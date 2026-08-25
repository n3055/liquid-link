package token

import (
	"context"

	"github.com/n3055/liquid-link/internal/crypto"
)

// Amount represents the smallest unit of the custom token (e.g., micro-tokens).
type Amount uint64

// BalanceRepository defines the data access contract needed by AssetManager to read/write balances.
type BalanceRepository interface {
	GetBalance(ctx context.Context, addr crypto.Address) (Amount, error)
	SetBalance(ctx context.Context, addr crypto.Address, balance Amount) error
}

// AssetManager governs issuance, redemption, and balance checks of tokens.
type AssetManager interface {
	// Mint creates new tokens out of thin air (authorized issuer node only).
	Mint(ctx context.Context, to crypto.Address, amount Amount) error

	// Burn permanently removes tokens from circulation.
	Burn(ctx context.Context, from crypto.Address, amount Amount) error

	// BalanceOf returns the current balance of a specific address.
	BalanceOf(ctx context.Context, addr crypto.Address) (Amount, error)

	// Credit increases the balance of an address by the specified amount.
	Credit(ctx context.Context, addr crypto.Address, amount Amount) error

	// Debit decreases the balance of an address by the specified amount.
	Debit(ctx context.Context, addr crypto.Address, amount Amount) error

	// TotalSupply returns the total circulating supply of the token.
	TotalSupply(ctx context.Context) (Amount, error)
}
