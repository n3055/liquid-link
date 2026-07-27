package routing

import (
	"context"

	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/token"
)

// Hop represents an intermediary node in a payment route.
type Hop struct {
	NodeID      string         `json:"node_id"`
	NodeAddress crypto.Address `json:"node_address"`
	Fee         token.Amount   `json:"fee"` // Fee charged by this hop for routing
}

// Route represents an ordered list of hops from sender to receiver.
type Route struct {
	Hops []Hop        `json:"hops"`
	Cost token.Amount `json:"cost"` // Total routing cost
}

// Router calculates optimal paths for sending payments across the PCN.
type Router interface {
	// FindRoute computes a path from sender to receiver that can support the requested transfer amount.
	FindRoute(ctx context.Context, sender, receiver crypto.Address, amount token.Amount) (*Route, error)
}

// TopologyTracker maintains the local view of the payment channel network topology.
type TopologyTracker interface {
	// UpdateChannelUpdates registers a channel update (e.g. balance, capacity change).
	UpdateChannelUpdates(ctx context.Context, update interface{}) error

	// GetTopology returns the internal network graph representation.
	GetTopology(ctx context.Context) (interface{}, error)
}
