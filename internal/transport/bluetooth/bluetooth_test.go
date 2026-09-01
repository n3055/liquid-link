package bluetooth_test

import (
	"context"
	"testing"

	"github.com/n3055/liquid-link/internal/config"
	"github.com/n3055/liquid-link/internal/transport"
	"github.com/n3055/liquid-link/internal/transport/bluetooth"
)

type dummyHandler struct {
	received []*transport.Message
}

func (h *dummyHandler) HandleMessage(ctx context.Context, msg *transport.Message) error {
	h.received = append(h.received, msg)
	return nil
}

func TestBluetoothMeshTransport(t *testing.T) {
	ctx := context.Background()
	cfg := &config.BluetoothConfig{Enabled: true, DeviceID: "hci0"}
	bt := bluetooth.NewMeshTransport(cfg)

	if err := bt.Start(ctx); err != nil {
		t.Fatalf("failed to start bluetooth transport: %v", err)
	}

	handler := &dummyHandler{}
	bt.RegisterHandler(ctx, handler)

	msg := &transport.Message{
		Type:     "proposal",
		SenderID: "node-a",
		Payload:  []byte("test-payload"),
	}

	if err := bt.Send(ctx, "node-b", msg); err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	if len(handler.received) != 1 {
		t.Errorf("expected 1 received message, got %d", len(handler.received))
	}

	// Broadcast duplicate packet (should be deduplicated)
	if err := bt.Broadcast(ctx, msg); err != nil {
		t.Fatalf("failed to broadcast message: %v", err)
	}

	if len(handler.received) != 1 {
		t.Errorf("expected duplicate packet deduplicated, got %d received", len(handler.received))
	}

	if err := bt.Stop(ctx); err != nil {
		t.Fatalf("failed to stop bluetooth transport: %v", err)
	}
}
