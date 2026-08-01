package wallet

import (
	"context"
	"errors"

	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/transaction"
)

// ErrAlreadyExists is returned when attempting to save a wallet under an alias that is already in use.
var ErrAlreadyExists = errors.New("wallet already exists")

// Wallet manages key storage, address creation, and transaction signing.
type Wallet interface {
	// Address returns the primary address of the wallet.
	Address() crypto.Address

	// SignTransaction signs a transaction payload using the wallet's private key.
	SignTransaction(ctx context.Context, tx *transaction.Transaction) error

	// ExportPublicKey exports the raw public key of the wallet.
	ExportPublicKey() crypto.PublicKey
}

// Repository defines the interface for local wallet persistence.
type Repository interface {
	// Save stores a wallet securely in local storage.
	Save(ctx context.Context, alias string, wallet Wallet) error
	// Load retrieves a wallet by its alias from local storage.
	Load(ctx context.Context, alias string) (Wallet, error)
	// List returns all stored wallet addresses and their aliases.
	List(ctx context.Context) (map[string]crypto.Address, error)
}
