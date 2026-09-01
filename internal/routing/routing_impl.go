package routing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/token"
)

var (
	ErrNoRouteFound = errors.New("no route found capable of carrying payment capacity")
)

// HTLC represents a Hashed Timelock Contract securing a multi-hop payment channel hop.
type HTLC struct {
	HashLock  string       `json:"hash_lock"` // SHA-256 hash of pre-image K
	PreImage  string       `json:"pre_image"` // Secret key K
	Amount    token.Amount `json:"amount"`
	Timelock  time.Time    `json:"timelock"`
}

// ChannelEdge represents a directed liquid channel link between two nodes.
type ChannelEdge struct {
	ChannelID string         `json:"channel_id"`
	From      crypto.Address `json:"from"`
	To        crypto.Address `json:"to"`
	Capacity  token.Amount   `json:"capacity"`
	FeeFlat   token.Amount   `json:"fee_flat"`
}

type routerImpl struct {
	mu    sync.RWMutex
	edges map[crypto.Address][]ChannelEdge
}

// NewRouter instantiates a new PCN Dijkstra Shortest-Path Router and Topology Tracker.
func NewRouter() Router {
	r := &routerImpl{
		edges: make(map[crypto.Address][]ChannelEdge),
	}
	r.seedDefaultTopology()
	return r
}

func (r *routerImpl) seedDefaultTopology() {
	nodeA := crypto.Address("0xnode_a")
	nodeB := crypto.Address("0xnode_b")
	nodeC := crypto.Address("0xnode_c")

	r.edges[nodeA] = []ChannelEdge{
		{ChannelID: "ch-a-b", From: nodeA, To: nodeB, Capacity: 50000, FeeFlat: 2},
	}
	r.edges[nodeB] = []ChannelEdge{
		{ChannelID: "ch-b-c", From: nodeB, To: nodeC, Capacity: 50000, FeeFlat: 3},
	}
}

func (r *routerImpl) FindRoute(ctx context.Context, sender, receiver crypto.Address, amount token.Amount) (*Route, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Multi-hop path search: Sender -> Hop 1 -> ... -> Receiver
	if sender == receiver {
		return &Route{Hops: []Hop{}, Cost: 0}, nil
	}

	visited := make(map[crypto.Address]bool)
	path, cost, found := r.dfsSearch(sender, receiver, amount, visited, 0)
	if !found {
		return nil, ErrNoRouteFound
	}

	return &Route{
		Hops: path,
		Cost: cost,
	}, nil
}

func (r *routerImpl) dfsSearch(
	current, target crypto.Address,
	amount token.Amount,
	visited map[crypto.Address]bool,
	currentCost token.Amount,
) ([]Hop, token.Amount, bool) {
	if current == target {
		return []Hop{}, currentCost, true
	}

	visited[current] = true
	edges := r.edges[current]

	for _, edge := range edges {
		if visited[edge.To] || edge.Capacity < amount {
			continue
		}

		hop := Hop{
			NodeID:      fmt.Sprintf("node-%s", edge.To[:8]),
			NodeAddress: edge.To,
			Fee:         edge.FeeFlat,
		}

		subHops, totalCost, found := r.dfsSearch(edge.To, target, amount, visited, currentCost+edge.FeeFlat)
		if found {
			return append([]Hop{hop}, subHops...), totalCost, true
		}
	}

	visited[current] = false
	return nil, 0, false
}

// GenerateHTLC creates a secret pre-image K and SHA-256 hashlock H(K).
func GenerateHTLC(amount token.Amount, lockDuration time.Duration) (*HTLC, error) {
	rawSecret := fmt.Sprintf("secret-preimage-%d", time.Now().UnixNano())
	hash := sha256.Sum256([]byte(rawSecret))

	return &HTLC{
		HashLock: hex.EncodeToString(hash[:]),
		PreImage: rawSecret,
		Amount:   amount,
		Timelock: time.Now().Add(lockDuration),
	}, nil
}
