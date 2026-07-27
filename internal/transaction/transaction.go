package transaction

import (
	"context"
	"time"

	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/token"
)

// Transaction represents a signed request to transfer tokens between accounts.
type Transaction struct {
	ID        string         `json:"id"`        // Unique hash of the transaction data
	Sender    crypto.Address `json:"sender"`    // Address of the sender
	Receiver  crypto.Address `json:"receiver"`  // Address of the receiver
	Amount    token.Amount   `json:"amount"`    // Number of tokens to transfer
	Nonce     uint64         `json:"nonce"`     // Prevent replay attacks
	Timestamp time.Time      `json:"timestamp"` // Time the transaction was created
	Signature []byte         `json:"signature"` // Cryptographic signature of Sender
}

// Validator defines rules for validating transactions before they enter the mempool or ledger.
type Validator interface {
	// VerifySignature validates that the transaction signature matches the sender's public key.
	VerifySignature(ctx context.Context, tx *Transaction) error

	// VerifyBalance checks if the sender has sufficient funds.
	VerifyBalance(ctx context.Context, tx *Transaction) error

	// VerifyNonce ensures the transaction nonce is sequentially correct.
	VerifyNonce(ctx context.Context, tx *Transaction) error
}

// Repository defines the interface for persisting and querying transactions.
type Repository interface {
	// Put stores a transaction in the database.
	Put(ctx context.Context, tx *Transaction) error
	// Get retrieves a transaction by ID.
	Get(ctx context.Context, id string) (*Transaction, error)
	// ListByAddress lists transactions associated with a specific address.
	ListByAddress(ctx context.Context, addr crypto.Address) ([]*Transaction, error)
}
