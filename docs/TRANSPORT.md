# Universal Transport Abstraction & Dual-Medium Driver Specification

## 1. Overview

LiquidLink relies on a **Transport-Agnostic Engine (DMT-Engine)**. The protocol decouples transaction signatures, block headers, and consensus votes from the transmission layer. Universal envelopes (`transport.Message`) can travel across high-bandwidth Internet TCP/IP sockets or low-bandwidth offline Bluetooth Low Energy (BLE) Mesh flooding without modifying core transaction signatures or state machine logic.

---

## 2. Universal Protocol Envelope Schema

```go
type Medium string

const (
	MediumInternet  Medium = "internet"
	MediumBluetooth Medium = "bluetooth"
)

type MessageType string

const (
	TypeTxProposal     MessageType = "tx_proposal"
	TypeBlockProposal  MessageType = "block_proposal"
	TypeConsensusVote  MessageType = "consensus_vote"
	TypeChannelState   MessageType = "channel_state"
	TypePeerHeartbeat  MessageType = "peer_heartbeat"
)

type Message struct {
	ID        string      `json:"id"`
	Type      MessageType `json:"type"`
	Sender    string      `json:"sender"`
	Payload   []byte      `json:"payload"`   // Serialized transaction/block/state payload
	Signature []byte      `json:"signature"` // Sender Ed25519 signature over Payload
	Timestamp int64       `json:"timestamp"`
}

type Transport interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Send(ctx context.Context, target string, msg *Message) error
	Broadcast(ctx context.Context, msg *Message) error
	RegisterHandler(ctx context.Context, handler Handler)
	Medium() Medium
}
```

---

## 3. Dual-Medium Transport Drivers

### 3.1 Internet Driver (`internal/transport/internet/`)
- **Protocol**: Direct TCP/IP socket connections and HTTP REST peer relays.
- **Topology**: Fully-connected mesh or bootstrap-node topology.
- **Use Case**: High-bandwidth node-to-node block proposals, vote collection, and REST API client communication.

### 3.2 Bluetooth Mesh Driver (`internal/transport/bluetooth/`)
- **Protocol**: Bluetooth Low Energy (BLE) Advertising & GATT Service Mesh flooding.
- **Service UUID**: `00001101-0000-1000-8000-00805F9B34FB`
- **Topology**: Decentralized ad-hoc local mesh flooding. Nodes re-broadcast incoming signed envelopes to adjacent physical devices.
- **Use Case**: Offline local payment channel updates and peer discovery when Internet connectivity is unavailable.

---

## 4. Handler Registration & Verification Flow

1. **Envelope Assembly**: Source node packages signed payload $P$ into `transport.Message`, signing payload hash $H(P)$ with Ed25519 private key.
2. **Transmission**: `Transport.Send` or `Transport.Broadcast` transmits envelope over active transport medium.
3. **Receipt & Handling**: Target node `Handler` validates envelope signature using sender public key $K_{\text{sender}}$, rejecting tampered or invalid messages before passing payload $P$ to state engine.
