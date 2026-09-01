package fiat

import (
	"context"
	"sync"

	"github.com/n3055/liquid-link/internal/crypto"
)

const (
	// DefaultInitialINRBalance represents initial simulated fiat credit assigned to new user wallets.
	DefaultInitialINRBalance = 1000.00
)

type managerImpl struct {
	mu       sync.RWMutex
	accounts map[crypto.Address]*Account
}

// [Fiat Bank] NewManager initializes a thread-safe simulated INR fiat bank manager.
func NewManager() Manager {
	return &managerImpl{
		accounts: make(map[crypto.Address]*Account),
	}
}

// [Fiat Bank] getOrCreateAccount retrieves or initializes an INR account with default starting balance.
func (m *managerImpl) getOrCreateAccount(user crypto.Address) *Account {
	acc, exists := m.accounts[user]
	if !exists {
		acc = &Account{
			User:      user,
			Available: DefaultInitialINRBalance,
			Reserved:  0.0,
		}
		m.accounts[user] = acc
	}
	return acc
}

func (m *managerImpl) CreditINR(ctx context.Context, user crypto.Address, amount float64) error {
	if amount <= 0 {
		return ErrInvalidINRAmount
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	acc := m.getOrCreateAccount(user)
	acc.Available += amount
	return nil
}

func (m *managerImpl) DebitINR(ctx context.Context, user crypto.Address, amount float64) error {
	if amount <= 0 {
		return ErrInvalidINRAmount
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	acc := m.getOrCreateAccount(user)
	if acc.Available < amount {
		return ErrInsufficientINRBalance
	}

	acc.Available -= amount
	return nil
}

func (m *managerImpl) ReserveINR(ctx context.Context, user crypto.Address, amount float64) error {
	if amount <= 0 {
		return ErrInvalidINRAmount
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	acc := m.getOrCreateAccount(user)
	if acc.Available < amount {
		return ErrInsufficientINRBalance
	}

	acc.Available -= amount
	acc.Reserved += amount
	return nil
}

func (m *managerImpl) ReleaseINR(ctx context.Context, user crypto.Address, amount float64) error {
	if amount <= 0 {
		return ErrInvalidINRAmount
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	acc := m.getOrCreateAccount(user)
	if acc.Reserved < amount {
		// Cap release at reserved amount to protect precision errors
		amount = acc.Reserved
	}

	acc.Reserved -= amount
	acc.Available += amount
	return nil
}

func (m *managerImpl) TransferINR(ctx context.Context, fromBuyer crypto.Address, toSeller crypto.Address, execAmount float64, reservedDeduction float64) error {
	if execAmount <= 0 || reservedDeduction <= 0 {
		return ErrInvalidINRAmount
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	buyerAcc := m.getOrCreateAccount(fromBuyer)
	sellerAcc := m.getOrCreateAccount(toSeller)

	// Deduct reserved escrow from buyer
	if buyerAcc.Reserved < reservedDeduction {
		// If reserved is slightly lower, take what remains and adjust available if needed
		diff := reservedDeduction - buyerAcc.Reserved
		buyerAcc.Reserved = 0
		if buyerAcc.Available >= diff {
			buyerAcc.Available -= diff
		}
	} else {
		buyerAcc.Reserved -= reservedDeduction
		// Return price improvement surplus back to buyer's available balance if buyer ordered at higher limit
		if reservedDeduction > execAmount {
			surplus := reservedDeduction - execAmount
			buyerAcc.Available += surplus
		}
	}

	// Credit execution INR amount to seller's available balance
	sellerAcc.Available += execAmount

	return nil
}

func (m *managerImpl) GetAccount(ctx context.Context, user crypto.Address) (*Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	acc := m.getOrCreateAccount(user)
	// Return defensive copy
	return &Account{
		User:      acc.User,
		Available: acc.Available,
		Reserved:  acc.Reserved,
	}, nil
}

func (m *managerImpl) GetTotalPoolReserve(ctx context.Context) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := 0.0
	for _, acc := range m.accounts {
		total += acc.Available + acc.Reserved
	}
	return total, nil
}
