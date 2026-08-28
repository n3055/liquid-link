package ledger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/storage"
	"github.com/n3055/liquid-link/internal/token"
)

// SystemAddress represents the address used for minting tokens (e.g., genesis allocation).
const SystemAddress = crypto.Address("0x0000000000000000000000000000000000000000")

// badgerStateRepository implements StateRepository using storage.Store.
type badgerStateRepository struct {
	store storage.Store
}

// NewStateRepository creates a new BadgerDB-backed StateRepository.
func NewStateRepository(store storage.Store) StateRepository {
	return &badgerStateRepository{store: store}
}

func (r *badgerStateRepository) SetAccountState(ctx context.Context, state *AccountState) error {
	encoded, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal account state: %w", err)
	}

	key := []byte(fmt.Sprintf("ledger:account:%s", state.Address))
	return r.store.Set(ctx, key, encoded)
}

func (r *badgerStateRepository) GetAccountState(ctx context.Context, addr crypto.Address) (*AccountState, error) {
	key := []byte(fmt.Sprintf("ledger:account:%s", addr))
	data, err := r.store.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var state AccountState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account state: %w", err)
	}

	return &state, nil
}

// ledgerImpl implements the Ledger interface with an in-memory cache.
type ledgerImpl struct {
	mu    sync.RWMutex
	repo  StateRepository
	cache map[crypto.Address]*AccountState
}

// NewLedger creates a new Ledger instance.
func NewLedger(repo StateRepository) Ledger {
	return &ledgerImpl{
		repo:  repo,
		cache: make(map[crypto.Address]*AccountState),
	}
}

// [Ledger] SetAccountState updates the account state in both persistent store and in-memory cache.
func (l *ledgerImpl) SetAccountState(ctx context.Context, state *AccountState) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.repo.SetAccountState(ctx, state); err != nil {
		return err
	}
	l.cache[state.Address] = copyState(state)
	return nil
}

func (l *ledgerImpl) GetAccountState(ctx context.Context, addr crypto.Address) (*AccountState, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if state, cached := l.cache[addr]; cached {
		return copyState(state), nil
	}

	state, err := l.repo.GetAccountState(ctx, addr)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) || os.IsNotExist(err) {
			state = &AccountState{
				Address: addr,
				Balance: 0,
				Nonce:   0,
			}
		} else {
			return nil, err
		}
	}

	l.cache[addr] = copyState(state)
	return copyState(state), nil
}

func (l *ledgerImpl) ApplyTransaction(ctx context.Context, sender, receiver crypto.Address, amount token.Amount, nonce uint64) error {
	senderState, err := l.GetAccountState(ctx, sender)
	if err != nil {
		return err
	}

	receiverState, err := l.GetAccountState(ctx, receiver)
	if err != nil {
		return err
	}

	// Apply validations and balance updates
	if sender == SystemAddress {
		// SystemAddress is used for minting, bypassing balance check
		receiverState.Balance += amount
	} else {
		if senderState.Balance < amount {
			return fmt.Errorf("insufficient balance: sender %s has %d, transaction requires %d", sender, senderState.Balance, amount)
		}
		if nonce != senderState.Nonce+1 {
			return fmt.Errorf("nonce mismatch: sender %s has nonce %d, transaction has %d", sender, senderState.Nonce, nonce)
		}

		senderState.Balance -= amount
		senderState.Nonce = nonce
		receiverState.Balance += amount
	}

	l.mu.Lock()
	l.cache[sender] = senderState
	l.cache[receiver] = receiverState
	l.mu.Unlock()

	return nil
}

func (l *ledgerImpl) CommitState(ctx context.Context, blockHeight uint64) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, state := range l.cache {
		if err := l.repo.SetAccountState(ctx, state); err != nil {
			return fmt.Errorf("failed to commit account state for %s: %w", state.Address, err)
		}
	}

	l.cache = make(map[crypto.Address]*AccountState)
	return nil
}

func (l *ledgerImpl) RollbackState(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.cache = make(map[crypto.Address]*AccountState)
	return nil
}

func copyState(s *AccountState) *AccountState {
	return &AccountState{
		Address: s.Address,
		Balance: s.Balance,
		Nonce:   s.Nonce,
	}
}
