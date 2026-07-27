package bluetooth

import (
	"context"
	"fmt"
	"sync"

	"github.com/n3055/liquid-link/internal/config"
	"github.com/n3055/liquid-link/internal/transport"
)

// Transport implements transport.Transport over Bluetooth Mesh.
type Transport struct {
	cfg     *config.BluetoothConfig
	handler transport.Handler
	mu      sync.RWMutex
	running bool
}

// NewTransport instantiates a new Bluetooth Mesh transport.
func NewTransport(cfg *config.BluetoothConfig) *Transport {
	return &Transport{
		cfg: cfg,
	}
}

// Start initiates the local Bluetooth interface and begins scanning/advertising.
func (t *Transport) Start(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.running {
		return fmt.Errorf("bluetooth transport already running")
	}

	t.running = true
	// TODO: Initialize Bluetooth adapter scanning/advertising
	return nil
}

// Stop terminates active Bluetooth service advertising and discovery.
func (t *Transport) Stop(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return nil
	}

	t.running = false
	// TODO: Shutdown bluetooth hardware bindings
	return nil
}

// Send broadcasts or relays a packet through the Bluetooth mesh to a target MAC/node.
func (t *Transport) Send(ctx context.Context, target string, msg *transport.Message) error {
	// TODO: Assemble packets and transmit via advertising/mesh routing
	return nil
}

// Broadcast sends a packet via mesh flood routing to all nearby nodes.
func (t *Transport) Broadcast(ctx context.Context, msg *transport.Message) error {
	// TODO: Flood-route packet to all neighbors
	return nil
}

// RegisterHandler registers the message handler for incoming Bluetooth signals.
func (t *Transport) RegisterHandler(ctx context.Context, handler transport.Handler) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.handler = handler
}

// Medium returns standard bluetooth transport type.
func (t *Transport) Medium() transport.Medium {
	return transport.MediumBluetooth
}
