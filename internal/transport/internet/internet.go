package internet

import (
	"context"
	"fmt"
	"sync"

	"github.com/n3055/liquid-link/internal/config"
	"github.com/n3055/liquid-link/internal/transport"
)

// Transport implements transport.Transport over standard TCP/IP.
type Transport struct {
	cfg     *config.InternetConfig
	handler transport.Handler
	mu      sync.RWMutex
	running bool
}

// NewTransport instantiates a new Internet TCP transport.
func NewTransport(cfg *config.InternetConfig) *Transport {
	return &Transport{
		cfg: cfg,
	}
}

// Start opens socket listeners and begins network polling.
func (t *Transport) Start(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.running {
		return fmt.Errorf("internet transport already running")
	}

	t.running = true
	// TODO: Initialize TCP listeners and peer dialers
	return nil
}

// Stop closes all active connections and listener sockets.
func (t *Transport) Stop(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return nil
	}

	t.running = false
	// TODO: Tear down active TCP connections
	return nil
}

// Send transmits a message envelope to a specific target IP/Port.
func (t *Transport) Send(ctx context.Context, target string, msg *transport.Message) error {
	// TODO: Dial target or reuse connection, serialize, and write
	return nil
}

// Broadcast sends a message to all bootstrap and discovered IP peers.
func (t *Transport) Broadcast(ctx context.Context, msg *transport.Message) error {
	// TODO: Loop over active connections and send
	return nil
}

// RegisterHandler registers the router/handler for incoming TCP messages.
func (t *Transport) RegisterHandler(ctx context.Context, handler transport.Handler) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.handler = handler
}

// Medium returns standard internet transport type.
func (t *Transport) Medium() transport.Medium {
	return transport.MediumInternet
}
