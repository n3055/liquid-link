package peer

import (
	"context"
	"time"

	"github.com/n3055/liquid-link/internal/crypto"
)

// Info holds public identification and location details of a peer.
type Info struct {
	ID        string         `json:"id"`
	Address   crypto.Address `json:"address"`
	Endpoint  string         `json:"endpoint"`
	LastSeen  time.Time      `json:"last_seen"`
	Connected bool           `json:"connected"`
}

// Registry stores, updates, and lists available peers on the network.
type Registry interface {
	// RegisterPeer adds or updates a peer in the local database.
	RegisterPeer(ctx context.Context, p *Info) error

	// UnregisterPeer removes a peer from the local registry.
	UnregisterPeer(ctx context.Context, id string) error

	// GetPeer retrieves details of a specific peer.
	GetPeer(ctx context.Context, id string) (*Info, error)

	// ListPeers returns all known peers.
	ListPeers(ctx context.Context) ([]*Info, error)
}

// Discovery service actively queries or advertises node existence to build the peer registry.
type Discovery interface {
	// Start begins background discovery loops (e.g., multicast DNS or gossip exchange).
	Start(ctx context.Context) error

	// Stop cleanly terminates discovery processes.
	Stop(ctx context.Context) error
}
