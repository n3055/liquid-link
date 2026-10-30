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

// [PCN State] ApplyOfflineStateStatement validates and commits an off-chain co-signed state sequence update.
func (m *managerImpl) ApplyOfflineStateStatement(ctx context.Context, statement *OfflineStateStatement) (*Channel, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if statement == nil {
		return nil, errors.New("nil offline state statement")
	}
	if statement.ChannelID == "" {
		return nil, errors.New("channel id required in offline statement")
	}

	ch, err := m.registry.GetChannel(ctx, statement.ChannelID)
	if err != nil {
		return nil, err
	}

	if ch.State == StateClosed {
		return nil, ErrInvalidState
	}

	if statement.Sequence <= ch.Sequence {
		return nil, fmt.Errorf("offline sequence %d must be strictly higher than current sequence %d", statement.Sequence, ch.Sequence)
	}

	// Invariant: Balance conservation check with integer overflow protection
	currentTotal := ch.Share.BalancePartyA + ch.Share.BalancePartyB
	if statement.BalanceA > currentTotal || statement.BalanceB > currentTotal || statement.BalanceA+statement.BalanceB != currentTotal {
		return nil, fmt.Errorf("balance invariant violated: statement balances %d/%d invalid for channel total %d", statement.BalanceA, statement.BalanceB, currentTotal)
	}

	// Mandatory cryptographic signature verification
	if len(statement.PublicKey) == 0 || len(statement.Signature) == 0 {
		return nil, errors.New("cryptographic signature and public key required on offline statement")
	}

	gen := crypto.NewGenerator()
	signerAddr := gen.DeriveAddress(crypto.PublicKey(statement.PublicKey))
	if signerAddr != ch.Share.PartyA && signerAddr != ch.Share.PartyB {
		return nil, fmt.Errorf("signer %s is not an authorized channel participant (%s or %s)", signerAddr, ch.Share.PartyA, ch.Share.PartyB)
	}

	verifier := crypto.NewVerifier()
	stateHash := sha256.Sum256([]byte(fmt.Sprintf("%s|%d|%d|%d", statement.ChannelID, statement.Sequence, statement.BalanceA, statement.BalanceB)))
	if !verifier.Verify(stateHash[:], statement.Signature, crypto.PublicKey(statement.PublicKey)) {
		return nil, errors.New("invalid cryptographic signature on offline statement")
	}

	// Update channel state
	ch.Sequence = statement.Sequence
	ch.Share.BalancePartyA = statement.BalanceA
	ch.Share.BalancePartyB = statement.BalanceB
	ch.Signature = statement.Signature

	if err := m.registry.RegisterChannel(ctx, ch); err != nil {
		return nil, fmt.Errorf("failed to register updated channel: %w", err)
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
	ch.Bond.Expiry = time.Now().Add(24 * time.Hour)
	return m.registry.RegisterChannel(ctx, ch)
}

func (m *managerImpl) ChallengeSettlement(ctx context.Context, channelID string, higherSeqChannel *Channel) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if higherSeqChannel == nil {
		return errors.New("nil challenge channel state")
	}

	ch, err := m.registry.GetChannel(ctx, channelID)
	if err != nil {
		return err
	}

	if ch.State != StateClosing {
		return errors.New("channel is not in closing state")
	}

	if higherSeqChannel.Sequence <= ch.Sequence {
		return fmt.Errorf("challenge sequence %d must be strictly higher than current sequence %d", higherSeqChannel.Sequence, ch.Sequence)
	}

	// Update channel state with higher sequence numbers
	ch.Sequence = higherSeqChannel.Sequence
	ch.Share = higherSeqChannel.Share
	ch.Signature = higherSeqChannel.Signature

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
