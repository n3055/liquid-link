package transport

import (
	"context"
)

// Medium represents the type of communication transport.
type Medium string

const (
	MediumInternet  Medium = "internet"
	MediumBluetooth Medium = "bluetooth"
)

// Message represents the universal payload envelope sent over any transport.
type Message struct {
	Type      string `json:"type"`      // Type of protocol message (e.g., "proposal", "vote", "tx")
	Payload   []byte `json:"payload"`   // Serialized message payload
	SenderID  string `json:"sender_id"` // ID of the sending node
	Signature []byte `json:"signature"` // Transport-level signature (if applicable)
}

// Handler handles incoming messages from a transport.
type Handler interface {
	// HandleMessage processes a received message.
	HandleMessage(ctx context.Context, msg *Message) error
}

// Transport defines the common behavior that any communication medium must satisfy.
type Transport interface {
	// Start opens the transport listeners and begins accepting connections/packets.
	Start(ctx context.Context) error

	// Stop cleanly shuts down the transport.
	Stop(ctx context.Context) error

	// Send transmits a message to a specific peer address.
	Send(ctx context.Context, target string, msg *Message) error

	// Broadcast transmits a message to all connected peers.
	Broadcast(ctx context.Context, msg *Message) error

	// RegisterHandler binds a message handler to receive incoming payloads.
	RegisterHandler(ctx context.Context, handler Handler)

	// Medium returns the underlying communication transport medium.
	Medium() Medium
}
