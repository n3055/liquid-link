package ledger

import (
	"context"
	"errors"
	"os"

	"github.com/dgraph-io/badger/v4"
	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/token"
)

// LedgerBalanceAdapter adapts ledger.StateRepository to token.BalanceRepository.
type LedgerBalanceAdapter struct {
	repo StateRepository
}

// NewLedgerBalanceAdapter returns a new token.BalanceRepository implementation.
func NewLedgerBalanceAdapter(repo StateRepository) token.BalanceRepository {
	return &LedgerBalanceAdapter{repo: repo}
}

// [Ledger] GetBalance retrieves the account balance for a given address from the state repository.
func (a *LedgerBalanceAdapter) GetBalance(ctx context.Context, addr crypto.Address) (token.Amount, error) {
	state, err := a.repo.GetAccountState(ctx, addr)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) || os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	return state.Balance, nil
}

// [Ledger] SetBalance sets the account balance for a given address in the state repository, preserving existing nonces.
func (a *LedgerBalanceAdapter) SetBalance(ctx context.Context, addr crypto.Address, balance token.Amount) error {
	state, err := a.repo.GetAccountState(ctx, addr)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) || os.IsNotExist(err) {
			state = &AccountState{
				Address: addr,
				Balance: balance,
				Nonce:   0,
			}
		} else {
			return err
		}
	} else {
		state.Balance = balance
	}

	return a.repo.SetAccountState(ctx, state)
}
