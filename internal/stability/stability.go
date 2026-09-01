package stability

import (
	"context"
	"errors"
	"time"

	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/token"
)

var (
	ErrInsufficientShares = errors.New("insufficient staked shares")
	ErrInsufficientBonds  = errors.New("insufficient bonds held")
	ErrNotContraction     = errors.New("bonds can only be purchased during price contraction (LQD < ₹1.00)")
	ErrNotExpansion       = errors.New("bonds can only be redeemed during price expansion (LQD > ₹1.00)")
)

// PegState represents the market state of LQD relative to ₹1.00 INR.
type PegState string

const (
	StateExpansion   PegState = "expansion"   // LQD > ₹1.00 (Mint & distribute seigniorage interest to stakers)
	StateContraction PegState = "contraction" // LQD < ₹1.00 (Sell discounted LQD-Bonds to contract supply)
	StatePegged      PegState = "pegged"      // LQD == ₹1.00 (Balanced equilibrium)
)

// Status provides a real-time summary of protocol stability metrics.
type Status struct {
	Price                 float64      `json:"price"`                   // Current LQD price in INR
	TargetPeg             float64      `json:"target_peg"`              // Target peg (₹1.00)
	State                 PegState     `json:"state"`                   // Current state
	TotalStakedShares     token.Amount `json:"total_staked_shares"`     // Total LQD-Shares staked
	TotalOutstandingBonds token.Amount `json:"total_outstanding_bonds"` // Total LQD-Bonds waiting for redemption
	CurrentAPY            float64      `json:"current_apy"`             // Staking yield APY percentage
	Epoch                 uint64       `json:"epoch"`                   // Current stability epoch counter
}

// PricePoint represents a historical price measurement for charting.
type PricePoint struct {
	Timestamp int64   `json:"timestamp"`
	Price     float64 `json:"price"`
	Volume    uint64  `json:"volume"`
}

// StakingPosition tracks a user's staked shares and accrued interest.
type StakingPosition struct {
	Address      crypto.Address `json:"address"`
	StakedShares token.Amount   `json:"staked_shares"`
	AccruedYield token.Amount   `json:"accrued_yield"`
	LastEpoch    uint64         `json:"last_epoch"`
}

// BondPosition tracks a user's held LQD-Bonds.
type BondPosition struct {
	Address       crypto.Address `json:"address"`
	BondsHeld     token.Amount   `json:"bonds_held"`
	PurchasePrice float64        `json:"purchase_price"`
	RedemptionVal token.Amount   `json:"redemption_val"`
}

// Engine manages the algorithmic seigniorage stability mechanism.
type Engine interface {
	// GetStatus returns the current stability metrics and APY.
	GetStatus(ctx context.Context) (*Status, error)

	// GetPriceHistory returns price points for UI charting.
	GetPriceHistory(ctx context.Context) ([]PricePoint, error)

	// StakeShares locks LQD/Shares to earn seigniorage interest payouts.
	StakeShares(ctx context.Context, user crypto.Address, amount token.Amount) (*StakingPosition, error)

	// UnstakeShares releases staked shares and claims accumulated yield.
	UnstakeShares(ctx context.Context, user crypto.Address, amount token.Amount) (*StakingPosition, error)

	// GetStakingPosition retrieves the user's staking position.
	GetStakingPosition(ctx context.Context, user crypto.Address) (*StakingPosition, error)

	// BuyBonds purchases discounted LQD-Bonds during contraction.
	BuyBonds(ctx context.Context, user crypto.Address, lqdAmount token.Amount) (*BondPosition, error)

	// RedeemBonds redeems LQD-Bonds for 1 LQD + interest during expansion.
	RedeemBonds(ctx context.Context, user crypto.Address, bondAmount token.Amount) (*BondPosition, error)

	// GetBondPosition retrieves the user's bond position.
	GetBondPosition(ctx context.Context, user crypto.Address) (*BondPosition, error)

	// SetMarketPrice updates the live market price of LQD token in INR.
	SetMarketPrice(price float64)

	// ProcessEpoch executes periodic seigniorage rebalancing and yield distribution.
	ProcessEpoch(ctx context.Context) error

	// StartBackgroundWorker launches periodic algorithmic seigniorage rebalancing.
	StartBackgroundWorker(ctx context.Context, interval time.Duration)
}
