package pcn

import (
	"context"
	"time"

	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/token"
)

// ChannelState represents the state of a payment channel.
type ChannelState string

const (
	StateOpen    ChannelState = "open"
	StateClosing ChannelState = "closing"
	StateClosed  ChannelState = "closed"
)

// Share represents the balance allocation of a payment channel between two participants.
type Share struct {
	PartyA        crypto.Address `json:"party_a"`
	PartyB        crypto.Address `json:"party_b"`
	BalancePartyA token.Amount   `json:"balance_party_a"`
	BalancePartyB token.Amount   `json:"balance_party_b"`
}

// Bond represents lockup parameters securing a payment channel.
type Bond struct {
	TxID             string       `json:"tx_id"` // Transaction locking the collateral
	LockedCollateral token.Amount `json:"locked_collateral"`
	Expiry           time.Time    `json:"expiry"`
}

// Channel represents an active payment channel between two nodes.
type Channel struct {
	ID        string       `json:"id"`
	State     ChannelState `json:"state"`
	Share     Share        `json:"share"`
	Bond      Bond         `json:"bond"`
	Sequence  uint64       `json:"sequence"`
	Signature []byte       `json:"signature"` // Co-signed state signature
}

// Manager handles the lifecycle of payment channels.
type Manager interface {
	// OpenChannel locks collateral and establishes a payment channel.
	OpenChannel(ctx context.Context, counterpart crypto.Address, initialDeposit token.Amount) (*Channel, error)

	// UpdateChannelState updates the balances of the channel, requiring mutual signatures.
	UpdateChannelState(ctx context.Context, channelID string, amount token.Amount, sending bool) (*Channel, error)

	// InitiateSettlement starts the dispute/settlement period.
	InitiateSettlement(ctx context.Context, channelID string) error

	// ChallengeSettlement challenges a closing channel with a higher sequence co-signed state payload.
	ChallengeSettlement(ctx context.Context, channelID string, higherSeqChannel *Channel) error

	// FinalizeSettlement closes the channel and distributes locked funds based on the last state.
	FinalizeSettlement(ctx context.Context, channelID string) error
}

// HolderRegistry tracks all active local channels and counterparty identities.
type HolderRegistry interface {
	// RegisterChannel stores a new channel details.
	RegisterChannel(ctx context.Context, ch *Channel) error

	// GetChannel retrieves active channel details.
	GetChannel(ctx context.Context, channelID string) (*Channel, error)

	// ListChannels lists all channels registered locally.
	ListChannels(ctx context.Context) ([]*Channel, error)
}
