package bluetooth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/n3055/liquid-link/internal/config"
	"github.com/n3055/liquid-link/internal/transport"
)

var (
	ErrTransportClosed = errors.New("bluetooth transport is not active")
)

// Universal BLE Protocol Message Types
const (
	TypePCNOffline      = "pcn_offline_payment"
	TypeDirectTxOffline = "direct_tx_offline"
)

// MeshTransport implements transport.Transport over Bluetooth Low Energy (BLE) Mesh.
type MeshTransport struct {
	cfg         *config.BluetoothConfig
	handler     transport.Handler
	mu          sync.RWMutex
	running     bool
	seenPackets map[string]time.Time
	peers       map[string]bool
	queue       []*transport.Message
}

// NewMeshTransport instantiates a new BLE Mesh Transport driver.
func NewMeshTransport(cfg *config.BluetoothConfig) transport.Transport {
	return &MeshTransport{
		cfg:         cfg,
		seenPackets: make(map[string]time.Time),
		peers:       make(map[string]bool),
		queue:       make([]*transport.Message, 0),
	}
}

func (t *MeshTransport) Start(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.running {
		return fmt.Errorf("bluetooth mesh transport already running")
	}

	t.running = true
	return nil
}

func (t *MeshTransport) Stop(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return nil
	}

	t.running = false
	return nil
}

func (t *MeshTransport) Send(ctx context.Context, target string, msg *transport.Message) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return ErrTransportClosed
	}

	if msg == nil {
		return errors.New("cannot send nil message")
	}

	// Deduplicate packet by hash
	pktKey := getPacketKey(msg)
	t.seenPackets[pktKey] = time.Now()

	if t.handler != nil {
		_ = t.handler.HandleMessage(ctx, msg)
	}

	return nil
}

func (t *MeshTransport) Broadcast(ctx context.Context, msg *transport.Message) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return ErrTransportClosed
	}

	if msg == nil {
		return errors.New("cannot broadcast nil message")
	}

	pktKey := getPacketKey(msg)
	if _, seen := t.seenPackets[pktKey]; seen {
		return nil // Avoid circular packet flooding loops
	}

	t.seenPackets[pktKey] = time.Now()

	if t.handler != nil {
		_ = t.handler.HandleMessage(ctx, msg)
	}

	return nil
}

func (t *MeshTransport) RegisterHandler(ctx context.Context, handler transport.Handler) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.handler = handler
}

func (t *MeshTransport) Medium() transport.Medium {
	return transport.MediumBluetooth
}

// [Bluetooth Transport] QueueMessage buffers an offline message for subsequent transmission.
func (t *MeshTransport) QueueMessage(msg *transport.Message) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.queue = append(t.queue, msg)
}

// [Bluetooth Transport] GetQueuedMessages retrieves all buffered offline messages.
func (t *MeshTransport) GetQueuedMessages() []*transport.Message {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]*transport.Message, len(t.queue))
	copy(result, t.queue)
	return result
}

// [Bluetooth Transport] ClearQueue flushes the offline message queue.
func (t *MeshTransport) ClearQueue() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.queue = make([]*transport.Message, 0)
}

func getPacketKey(msg *transport.Message) string {
	h := sha256.Sum256(msg.Payload)
	return fmt.Sprintf("%s|%s|%s", msg.SenderID, msg.Type, hex.EncodeToString(h[:8]))
}
