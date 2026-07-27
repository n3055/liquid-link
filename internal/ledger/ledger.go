package ledger

import (
	"context"

	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/token"
)

// AccountState represents the current state of a ledger account.
type AccountState struct {
	Address crypto.Address `json:"address"`
	Balance token.Amount   `json:"balance"`
	Nonce   uint64         `json:"nonce"`
}

// Ledger defines the core interface to inspect and modify the global state of accounts.
type Ledger interface {
	// GetAccountState retrieves the balance and nonce for an address.
	GetAccountState(ctx context.Context, addr crypto.Address) (*AccountState, error)

	// ApplyTransaction applies a state transition for a transaction.
	ApplyTransaction(ctx context.Context, sender, receiver crypto.Address, amount token.Amount, nonce uint64) error

	// CommitState commits a batch of changes (from a block) to the state store.
	CommitState(ctx context.Context, blockHeight uint64) error

	// RollbackState rolls back any uncommitted state changes.
	RollbackState(ctx context.Context) error
}

// StateRepository defines persistent read/write operations for account state.
type StateRepository interface {
	// SetAccountState saves the account state.
	SetAccountState(ctx context.Context, state *AccountState) error
	// GetAccountState retrieves the account state.
	GetAccountState(ctx context.Context, addr crypto.Address) (*AccountState, error)
}
