package receipt

import (
	"context"
	"time"

	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/token"
)

// Receipt represents a signed cryptographic proof of a successful crypto-to-fiat transfer.
type Receipt struct {
	ID        string         `json:"id"`
	UTR       string         `json:"utr"`       // Unique Transaction Reference from fiat settlement
	Address   crypto.Address `json:"address"`   // Claimant/receiver address
	Amount    token.Amount   `json:"amount"`    // Bridged amount
	Timestamp time.Time      `json:"timestamp"` // Settlement time
	Signature []byte         `json:"signature"` // Bridge operator signature
}

// UTRMapping connects a Unique Transaction Reference (bank transfer ID) to local transactions.
type UTRMapping struct {
	UTR           string    `json:"utr"`
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"` // e.g. "pending", "linked", "settled"
	CreatedAt     time.Time `json:"created_at"`
}

// Registry stores and validates receipts and maps them to UTRs.
type Registry interface {
	// RegisterReceipt records a validated bridge receipt.
	RegisterReceipt(ctx context.Context, r *Receipt) error

	// VerifyReceipt checks the signature and authenticity of a receipt.
	VerifyReceipt(ctx context.Context, r *Receipt) (bool, error)

	// MapUTR links a fiat UTR to a local crypto transaction.
	MapUTR(ctx context.Context, utr string, txID string) error

	// GetUTRMapping retrieves the mapping details for a given UTR.
	GetUTRMapping(ctx context.Context, utr string) (*UTRMapping, error)

	// ListUTRMapping returns all UTR mappings.
	ListUTRMapping(ctx context.Context) ([]*UTRMapping, error)
}

// Bridge represents the crypto-to-fiat gateway orchestrator.
type Bridge interface {
	// InitiateBridgeRequest locks crypto assets to trigger a fiat payment.
	InitiateBridgeRequest(ctx context.Context, to crypto.Address, amount token.Amount) (*UTRMapping, error)

	// SettleBridgeRequest completes the bridging process once fiat payment receipt is presented.
	SettleBridgeRequest(ctx context.Context, utr string, receipt *Receipt) error
}
