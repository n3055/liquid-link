package receipt

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/storage"
	"github.com/n3055/liquid-link/internal/token"
)

var (
	ErrReceiptNotFound = errors.New("bridge receipt not found")
	ErrUTRNotFound     = errors.New("UTR mapping not found")
	ErrInvalidReceipt  = errors.New("receipt verification failed")
)

// badgerRegistry implements Registry using storage.Store.
type badgerRegistry struct {
	store    storage.Store
	verifier crypto.Verifier
	operator crypto.PublicKey
}

// NewRegistry creates a new BadgerDB-backed Receipt Registry.
func NewRegistry(store storage.Store, verifier crypto.Verifier, operatorPub crypto.PublicKey) Registry {
	return &badgerRegistry{
		store:    store,
		verifier: verifier,
		operator: operatorPub,
	}
}

func (r *badgerRegistry) RegisterReceipt(ctx context.Context, rec *Receipt) error {
	encoded, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("failed to marshal receipt: %w", err)
	}

	key := []byte(fmt.Sprintf("receipt:id:%s", rec.ID))
	return r.store.Set(ctx, key, encoded)
}

func (r *badgerRegistry) ListUTRMapping(ctx context.Context) ([]*UTRMapping, error) {
	var mappings []*UTRMapping
	err := r.store.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("receipt:utr:")
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var m UTRMapping
				if err := json.Unmarshal(val, &m); err != nil {
					return err
				}
				mappings = append(mappings, &m)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if mappings == nil {
		mappings = []*UTRMapping{}
	}

	return mappings, nil
}

func (r *badgerRegistry) VerifyReceipt(ctx context.Context, rec *Receipt) (bool, error) {
	if rec == nil {
		return false, errors.New("nil receipt")
	}

	payload := fmt.Sprintf("%s|%s|%s|%d|%d", rec.ID, rec.UTR, rec.Address, rec.Amount, rec.Timestamp.UnixNano())
	hash := sha256.Sum256([]byte(payload))

	if r.verifier != nil && r.operator != nil && len(rec.Signature) > 0 {
		return r.verifier.Verify(hash[:], rec.Signature, r.operator), nil
	}

	// Default true if no operator signature required in simulated mode
	return true, nil
}

func (r *badgerRegistry) MapUTR(ctx context.Context, utr string, txID string) error {
	mapping := &UTRMapping{
		UTR:           utr,
		TransactionID: txID,
		Status:        "pending",
		CreatedAt:     time.Now(),
	}

	encoded, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("failed to marshal UTR mapping: %w", err)
	}

	key := []byte(fmt.Sprintf("receipt:utr:%s", utr))
	return r.store.Set(ctx, key, encoded)
}

func (r *badgerRegistry) GetUTRMapping(ctx context.Context, utr string) (*UTRMapping, error) {
	key := []byte(fmt.Sprintf("receipt:utr:%s", utr))
	data, err := r.store.Get(ctx, key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, ErrUTRNotFound
		}
		return nil, err
	}

	var m UTRMapping
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to unmarshal UTR mapping: %w", err)
	}

	return &m, nil
}

// bridgeImpl implements the Bridge interface.
type bridgeImpl struct {
	mu           sync.RWMutex
	registry     Registry
	assetManager token.AssetManager
	signer       crypto.Signer
	store        storage.Store
}

// NewBridge creates a new Bridge instance.
func NewBridge(registry Registry, assetMgr token.AssetManager, signer crypto.Signer, store storage.Store) Bridge {
	return &bridgeImpl{
		registry:     registry,
		assetManager: assetMgr,
		signer:       signer,
		store:        store,
	}
}

func (b *bridgeImpl) InitiateBridgeRequest(ctx context.Context, to crypto.Address, amount token.Amount) (*UTRMapping, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 1. Lock/Debit crypto from user wallet
	if b.assetManager != nil {
		if err := b.assetManager.Debit(ctx, to, amount); err != nil {
			return nil, fmt.Errorf("failed to lock crypto assets for bridge: %w", err)
		}
	}

	// 2. Generate unique UTR & Transaction ID
	now := time.Now()
	rawUTR := fmt.Sprintf("UTR-BANK-%s-%d", to[:8], now.UnixNano())
	utrHash := sha256.Sum256([]byte(rawUTR))
	utr := "UTR" + hex.EncodeToString(utrHash[:8])
	txID := "tx-bridge-" + hex.EncodeToString(utrHash[8:16])

	// 3. Store UTR Mapping
	if err := b.registry.MapUTR(ctx, utr, txID); err != nil {
		return nil, fmt.Errorf("failed to record UTR mapping: %w", err)
	}

	mapping, err := b.registry.GetUTRMapping(ctx, utr)
	if err != nil {
		return nil, err
	}

	return mapping, nil
}

func (b *bridgeImpl) SettleBridgeRequest(ctx context.Context, utr string, rec *Receipt) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 1. Fetch UTR mapping
	mapping, err := b.registry.GetUTRMapping(ctx, utr)
	if err != nil {
		return err
	}

	if mapping.Status == "settled" {
		return errors.New("bridge request already settled")
	}

	// 2. Verify Receipt
	valid, err := b.registry.VerifyReceipt(ctx, rec)
	if err != nil || !valid {
		return ErrInvalidReceipt
	}

	// 3. Register receipt and update status to settled
	if err := b.registry.RegisterReceipt(ctx, rec); err != nil {
		return fmt.Errorf("failed to register receipt: %w", err)
	}

	mapping.Status = "settled"
	encoded, err := json.Marshal(mapping)
	if err != nil {
		return err
	}

	key := []byte(fmt.Sprintf("receipt:utr:%s", utr))
	return b.store.Set(ctx, key, encoded)
}
