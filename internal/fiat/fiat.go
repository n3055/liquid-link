package fiat

import (
	"context"
	"errors"

	"github.com/n3055/liquid-link/internal/crypto"
)

var (
	ErrInvalidINRAmount        = errors.New("amount must be greater than zero")
	ErrInsufficientINRBalance  = errors.New("insufficient available INR balance")
	ErrInsufficientINRReserved = errors.New("insufficient reserved INR balance")
	ErrAccountNotFound         = errors.New("INR bank account not found")
)

// [Fiat Bank] Account represents a user's simulated INR bank account balance and reserves.
type Account struct {
	User      crypto.Address `json:"user"`
	Available float64        `json:"available"` // Unlocked INR available for spending/bidding
	Reserved  float64        `json:"reserved"`  // INR locked in active open buy orders
}

// [Fiat Bank] Manager defines bank operations for simulated INR fiat balances and order escrow.
type Manager interface {
	// CreditINR credits INR to a user's available bank balance.
	CreditINR(ctx context.Context, user crypto.Address, amount float64) error

	// DebitINR debits available INR from a user's bank balance.
	DebitINR(ctx context.Context, user crypto.Address, amount float64) error

	// ReserveINR locks available INR into reserved escrow for open buy orders.
	ReserveINR(ctx context.Context, user crypto.Address, amount float64) error

	// ReleaseINR unlocks reserved INR back into available balance upon order cancellation.
	ReleaseINR(ctx context.Context, user crypto.Address, amount float64) error

	// TransferINR settles INR from buyer's reserved balance to seller's available balance upon trade match.
	TransferINR(ctx context.Context, fromBuyer crypto.Address, toSeller crypto.Address, execAmount float64, reservedDeduction float64) error

	// GetAccount returns a user's INR account status.
	GetAccount(ctx context.Context, user crypto.Address) (*Account, error)

	// GetTotalPoolReserve returns the aggregated fiat bank pool reserve across all accounts.
	GetTotalPoolReserve(ctx context.Context) (float64, error)
}
