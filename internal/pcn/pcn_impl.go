package pcn

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
	"github.com/n3055/liquid-link/internal/ledger"
	"github.com/n3055/liquid-link/internal/storage"
	"github.com/n3055/liquid-link/internal/token"
)

var (
	ErrChannelNotFound = errors.New("payment channel not found")
	ErrInvalidState    = errors.New("channel is not in open state")
	ErrInsufficientBal = errors.New("insufficient balance in channel share")
)

// badgerHolderRegistry implements HolderRegistry using storage.Store.
type badgerHolderRegistry struct {
	store storage.Store
}

// NewHolderRegistry creates a new BadgerDB-backed HolderRegistry.
func NewHolderRegistry(store storage.Store) HolderRegistry {
	return &badgerHolderRegistry{store: store}
}

func (r *badgerHolderRegistry) RegisterChannel(ctx context.Context, ch *Channel) error {
	encoded, err := json.Marshal(ch)
	if err != nil {
		return fmt.Errorf("failed to marshal channel: %w", err)
	}

	key := []byte(fmt.Sprintf("pcn:channel:%s", ch.ID))
	return r.store.Set(ctx, key, encoded)
}

func (r *badgerHolderRegistry) GetChannel(ctx context.Context, channelID string) (*Channel, error) {
	key := []byte(fmt.Sprintf("pcn:channel:%s", channelID))
	data, err := r.store.Get(ctx, key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, ErrChannelNotFound
		}
		return nil, err
	}

	var ch Channel
	if err := json.Unmarshal(data, &ch); err != nil {
		return nil, fmt.Errorf("failed to unmarshal channel: %w", err)
	}

	return &ch, nil
}

func (r *badgerHolderRegistry) ListChannels(ctx context.Context) ([]*Channel, error) {
	var channels []*Channel
	err := r.store.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("pcn:channel:")
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var ch Channel
				if err := json.Unmarshal(val, &ch); err != nil {
					return err
				}
				channels = append(channels, &ch)
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

	return channels, nil
}

// managerImpl implements the Manager interface.
type managerImpl struct {
	mu           sync.RWMutex
	nodeAddr     crypto.Address
	signer       crypto.Signer
	ledger       ledger.Ledger
	registry     HolderRegistry
	assetManager token.AssetManager
}

// NewManager creates a new PCN Manager.
func NewManager(
	nodeAddr crypto.Address,
	signer crypto.Signer,
	l ledger.Ledger,
	registry HolderRegistry,
	assetMgr token.AssetManager,
) Manager {
	return &managerImpl{
		nodeAddr:     nodeAddr,
		signer:       signer,
		ledger:       l,
		registry:     registry,
		assetManager: assetMgr,
	}
}

func (m *managerImpl) OpenChannel(ctx context.Context, counterpart crypto.Address, initialDeposit token.Amount) (*Channel, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Unique channel ID
	idRaw := fmt.Sprintf("%s|%s|%d|%d", m.nodeAddr, counterpart, initialDeposit, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(idRaw))
	channelID := hex.EncodeToString(hash[:16])

	// Lock collateral (debit node balance via ledger/asset manager)
	if m.assetManager != nil {
		if err := m.assetManager.Debit(ctx, m.nodeAddr, initialDeposit); err != nil {
			return nil, fmt.Errorf("failed to lock collateral deposit: %w", err)
		}
	}

	ch := &Channel{
		ID:    channelID,
		State: StateOpen,
		Share: Share{
			PartyA:        m.nodeAddr,
			PartyB:        counterpart,
			BalancePartyA: initialDeposit,
			BalancePartyB: 0,
		},
		Bond: Bond{
			TxID:             fmt.Sprintf("tx-bond-%s", channelID[:8]),
			LockedCollateral: initialDeposit,
			Expiry:           time.Now().Add(24 * time.Hour),
		},
		Sequence: 1,
	}

	// Sign channel state
	stateHash := sha256.Sum256([]byte(fmt.Sprintf("%s|%d|%d|%d", ch.ID, ch.Sequence, ch.Share.BalancePartyA, ch.Share.BalancePartyB)))
	if m.signer != nil {
		sig, err := m.signer.Sign(stateHash[:])
		if err == nil {
			ch.Signature = sig
		}
	}

	if err := m.registry.RegisterChannel(ctx, ch); err != nil {
		if m.assetManager != nil {
			_ = m.assetManager.Credit(ctx, m.nodeAddr, initialDeposit)
		}
		return nil, fmt.Errorf("failed to register channel: %w", err)
	}

	return ch, nil
}

func (m *managerImpl) UpdateChannelState(ctx context.Context, channelID string, amount token.Amount, sending bool) (*Channel, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch, err := m.registry.GetChannel(ctx, channelID)
	if err != nil {
		return nil, err
	}

	if ch.State != StateOpen {
		return nil, ErrInvalidState
	}

	if sending {
		if ch.Share.BalancePartyA < amount {
			return nil, ErrInsufficientBal
		}
		ch.Share.BalancePartyA -= amount
		ch.Share.BalancePartyB += amount
	} else {
		if ch.Share.BalancePartyB < amount {
			return nil, ErrInsufficientBal
		}
		ch.Share.BalancePartyB -= amount
		ch.Share.BalancePartyA += amount
	}

	ch.Sequence++

	stateHash := sha256.Sum256([]byte(fmt.Sprintf("%s|%d|%d|%d", ch.ID, ch.Sequence, ch.Share.BalancePartyA, ch.Share.BalancePartyB)))
	if m.signer != nil {
		sig, err := m.signer.Sign(stateHash[:])
		if err == nil {
			ch.Signature = sig
		}
	}

	if err := m.registry.RegisterChannel(ctx, ch); err != nil {
		return nil, fmt.Errorf("failed to update channel: %w", err)
	}

	return ch, nil
}

func (m *managerImpl) InitiateSettlement(ctx context.Context, channelID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch, err := m.registry.GetChannel(ctx, channelID)
	if err != nil {
		return err
	}

	if ch.State != StateOpen {
		return ErrInvalidState
	}

	ch.State = StateClosing
	return m.registry.RegisterChannel(ctx, ch)
}

func (m *managerImpl) FinalizeSettlement(ctx context.Context, channelID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch, err := m.registry.GetChannel(ctx, channelID)
	if err != nil {
		return err
	}

	if ch.State != StateClosing && ch.State != StateOpen {
		return errors.New("channel cannot be finalized")
	}

	// Refund/Distribute balances to accounts
	if m.assetManager != nil {
		if ch.Share.BalancePartyA > 0 {
			if err := m.assetManager.Credit(ctx, ch.Share.PartyA, ch.Share.BalancePartyA); err != nil {
				return fmt.Errorf("failed to credit party A (%s): %w", ch.Share.PartyA, err)
			}
		}
		if ch.Share.BalancePartyB > 0 {
			if err := m.assetManager.Credit(ctx, ch.Share.PartyB, ch.Share.BalancePartyB); err != nil {
				if ch.Share.BalancePartyA > 0 {
					_ = m.assetManager.Debit(ctx, ch.Share.PartyA, ch.Share.BalancePartyA)
				}
				return fmt.Errorf("failed to credit party B (%s): %w", ch.Share.PartyB, err)
			}
		}
	}

	ch.State = StateClosed
	return m.registry.RegisterChannel(ctx, ch)
}
