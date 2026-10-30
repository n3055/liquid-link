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

func TestBluetoothOfflineQueue(t *testing.T) {
	cfg := &config.BluetoothConfig{Enabled: true, DeviceID: "hci0"}
	bt := bluetooth.NewMeshTransport(cfg)

	meshTransport, ok := bt.(interface {
		QueueMessage(msg *transport.Message)
		GetQueuedMessages() []*transport.Message
		ClearQueue()
	})
	if !ok {
		t.Fatalf("expected MeshTransport to implement queue methods")
	}

	if len(meshTransport.GetQueuedMessages()) != 0 {
		t.Errorf("expected empty queue, got %d", len(meshTransport.GetQueuedMessages()))
	}

	msg1 := &transport.Message{Type: bluetooth.TypePCNOffline, Payload: []byte("pcn-payload-1"), SenderID: "node-1"}
	msg2 := &transport.Message{Type: bluetooth.TypeDirectTxOffline, Payload: []byte("tx-payload-2"), SenderID: "node-1"}

	meshTransport.QueueMessage(msg1)
	meshTransport.QueueMessage(msg2)

	queued := meshTransport.GetQueuedMessages()
	if len(queued) != 2 {
		t.Fatalf("expected 2 queued messages, got %d", len(queued))
	}
	if queued[0].Type != bluetooth.TypePCNOffline {
		t.Errorf("expected first message type %s, got %s", bluetooth.TypePCNOffline, queued[0].Type)
	}

	meshTransport.ClearQueue()
	if len(meshTransport.GetQueuedMessages()) != 0 {
		t.Errorf("expected queue cleared, got %d", len(meshTransport.GetQueuedMessages()))
	}
}
