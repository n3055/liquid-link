package orderbook

import (
	"context"
	"errors"
	"time"

	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/token"
)

var (
	ErrInvalidPrice           = errors.New("price must be greater than zero")
	ErrInvalidAmount          = errors.New("amount must be greater than zero")
	ErrInvalidSide            = errors.New("order side must be 'buy' or 'sell'")
	ErrOrderNotFound          = errors.New("order not found")
	ErrUnauthorized           = errors.New("unauthorized order modification")
	ErrOrderCompleted         = errors.New("order is already filled or cancelled")
	ErrInsufficientINRBalance = errors.New("insufficient available INR balance to place buy order")
	ErrInsufficientLQDBalance = errors.New("insufficient LQD token balance to place sell order")
)

// OrderSide specifies whether an order is buying or selling tokens.
type OrderSide string

const (
	SideBuy  OrderSide = "buy"  // Bid: Buying LQD with INR
	SideSell OrderSide = "sell" // Ask: Selling LQD for INR
)

// OrderStatus tracks the current lifecycle state of a limit order.
type OrderStatus string

const (
	StatusOpen      OrderStatus = "open"
	StatusPartial   OrderStatus = "partial"
	StatusFilled    OrderStatus = "filled"
	StatusCancelled OrderStatus = "cancelled"
)

// Order represents a limit order placed in the LQD/INR orderbook.
type Order struct {
	ID        string         `json:"id"`
	User      crypto.Address `json:"user"`
	Side      OrderSide      `json:"side"`
	Price     float64        `json:"price"`  // Price per LQD in INR
	Amount    token.Amount   `json:"amount"` // Target token quantity
	Filled    token.Amount   `json:"filled"` // Quantity filled so far
	Status    OrderStatus    `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
}

// Trade represents an executed transaction between a buyer and a seller.
type Trade struct {
	ID          string         `json:"id"`
	BuyOrderID  string         `json:"buy_order_id"`
	SellOrderID string         `json:"sell_order_id"`
	Buyer       crypto.Address `json:"buyer"`
	Seller      crypto.Address `json:"seller"`
	Price       float64        `json:"price"`  // Price in INR per token
	Amount      token.Amount   `json:"amount"` // Quantity traded
	Timestamp   time.Time      `json:"timestamp"`
}

// PriceLevel aggregates total quantity available at a given price point.
type PriceLevel struct {
	Price    float64      `json:"price"`
	Quantity token.Amount `json:"quantity"`
}

// Depth holds top bid and ask levels on the orderbook.
type Depth struct {
	Bids []PriceLevel `json:"bids"` // Highest bids first
	Asks []PriceLevel `json:"asks"` // Lowest asks first
}

// Engine defines operations for the LQD/INR limit orderbook matching engine.
type Engine interface {
	// PlaceOrder submits a new limit order and immediately executes matches.
	PlaceOrder(ctx context.Context, order *Order) (*Order, []*Trade, error)

	// CancelOrder cancels an active open order.
	CancelOrder(ctx context.Context, orderID string, user crypto.Address) error

	// GetDepth returns the aggregated orderbook depth (bids & asks).
	GetDepth(ctx context.Context) (*Depth, error)

	// GetTrades returns recent executed trades up to limit.
	GetTrades(ctx context.Context, limit int) ([]*Trade, error)

	// GetUserOrders returns all orders belonging to a user address.
	GetUserOrders(ctx context.Context, user crypto.Address) ([]*Order, error)

	// GetMarketPrice returns the current clearing market price in INR.
	GetMarketPrice(ctx context.Context) float64
}
