# LiquidLink System Architecture & Design Specification

## 1. Executive Summary

LiquidLink is an educational and research-grade decentralized payment channel network (PCN) and crypto-to-fiat bridge. The system enables low-latency off-chain transfers over heterogeneous transport channels (TCP/IP networks, offline Bluetooth Mesh), backed by an algorithmic seigniorage stability peg engine (LQD / Shares / Bonds) and a non-custodial fiat settlement bridge.

---

## 2. Layered Architecture

The protocol is structured into five decoupled conceptual layers:

```
┌─────────────────────────────────────────────────────────────────┐
│                     1. Application & Web UI                     │
│           Chi REST Router (/api/v1), Control Panel Dashboard    │
└────────────────────────────────┬────────────────────────────────┘
                                 │
┌────────────────────────────────▼────────────────────────────────┐
│                   2. Orchestration & Domain                     │
│           Node Container (internal/node), AssetManager          │
└────────────────────────────────┬────────────────────────────────┘
                                 │
┌────────────────────────────────▼────────────────────────────────┐
│                    3. State & Logic Layer                       │
│     PCN State Channels, Seigniorage Engine, Orderbook Engine    │
└────────────────────────────────┬────────────────────────────────┘
                                 │
┌────────────────────────────────▼────────────────────────────────┐
│               4. Consensus & Storage Layer                      │
│     Rotating Leader PBFT, SHA-256 Blockchain, BadgerDB Store    │
└────────────────────────────────┬────────────────────────────────┘
                                 │
┌────────────────────────────────▼────────────────────────────────┐
│                     5. Transport Layer                          │
│     Universal Envelope, TCP/IP Driver, BLE Mesh Driver          │
└─────────────────────────────────────────────────────────────────┘
```

### Layer Descriptions

1. **Application & Web UI Layer (`internal/api/`)**
   - Exposes JSON REST APIs (`/api/v1/*`) and renders the real-time node dashboard (`handleDashboard`). Handles request decoding, parameter validation, and HTTP response formatting.
2. **Orchestration & Domain Layer (`internal/node/`, `internal/token/`)**
   - Bootstraps storage repositories, persistent wallet keys, signers, and injects dependencies into sub-services. `AssetManager` provides high-level `Mint`, `Burn`, `Credit`, and `Debit` state transition methods.
3. **State & Logic Layer (`internal/pcn/`, `internal/stability/`, `internal/orderbook/`)**
   - Manages off-chain bilateral state channels, calculates seigniorage peg adjustments (expansion dividends / contraction bonds), and matches LQD/INR limit orders.
4. **Consensus & Storage Layer (`internal/blockchain/`, `internal/consensus/`, `internal/storage/`)**
   - Maintains canonical SHA-256 block sequences, executes rotating leader PBFT majority voting (51%+ threshold), and handles persistent key-value storage via BadgerDB.
5. **Transport Layer (`internal/transport/`)**
   - Encapsulates universal protocol envelopes (`transport.Message`) signed with Ed25519 keys, transmitting messages across Internet TCP/IP or offline Bluetooth Mesh.

---

## 3. Core Component Dependencies & Constructor Wiring

All logic components rely on interface abstractions and constructor dependency injection (`New<Service>`):

```
Store (BadgerDB)
   ├── StateRepository → Ledger → LedgerBalanceAdapter → AssetManager
   ├── BlockRepository → Blockchain → ConsensusEngine
   ├── HolderRegistry  → PCN Manager
   ├── ReceiptRegistry → Fiat Bridge
   └── PeerRegistry
```

---

## 4. End-to-End Transaction Lifecycles

### 4.1 Direct On-Chain Transfer Sequence
1. Sender submits transaction via REST API `POST /api/v1/wallet/transfer`.
2. `AssetManager.Debit` deducts amount from sender account state in BadgerDB ledger.
3. `AssetManager.Credit` adds amount to receiver account state.
4. Asynchronous `syncPeerCredit` notifies peer nodes via `POST /api/v1/wallet/credit-sync` to update peer balance state without altering circulating total supply.
5. Proposer packages transaction into block; PBFT consensus validators sign and commit canonical block.

### 4.2 Off-Chain State Channel Payment Sequence
1. Party A and Party B open channel via `POST /api/v1/channels/open`. Initial collateral deposit is locked in BadgerDB.
2. Off-chain transfers execute by incrementing sequence counter $S \to S+1$ and updating balance split $(B_A, B_B)$.
3. Both parties co-sign state hash $H(\text{ChannelID} \parallel S \parallel B_A \parallel B_B)$.
4. Upon settlement, latest co-signed state statement is submitted to `FinalizeSettlement` to release locked collateral.

---

## 5. Security & Invariant Rules

- **Total Supply Preservation:** Account transfers (`Credit`/`Debit`) must preserve total circulating supply. Only authorized issuer routines (`Mint`/`Burn`) modify total supply.
- **Atomic Balance Updates:** `Debit` operations check balance sufficiency ($B \ge \text{amount}$) before modifying state. If subsequent credit operations fail, debits are automatically rolled back.
- **Thread Safety:** All in-memory state objects (ledger cache, orderbook maps, stability metrics) enforce concurrent access protection using standard `sync.RWMutex`.
