# LiquidLink: A Transport-Agnostic Payment Channel Network & Algorithmic Seigniorage Crypto-to-Fiat Bridge

LiquidLink is a novel decentralized payment network and crypto-to-fiat bridge. It combines **Transport-Agnostic Envelope Routing (TCP/IP & Bluetooth Mesh)**, **Optimistic Sequence-Based State Channels (PCN)**, an **Algorithmic Seigniorage Peg Engine (LQD / Shares / Bonds)**, and a **Non-Custodial Fiat Bridge with Cryptographic UTR Receipts**.

---

## 🔬 Core Architectural Pillars & Research Scope

```
                               ┌────────────────────────────────┐
                               │   LiquidLink Protocol Stack    │
                               └───────────────┬────────────────┘
                                               │
         ┌───────────────────┬─────────────────┼───────────────────┬───────────────────┐
         │                   │                 │                   │                   │
┌────────▼────────┐ ┌────────▼────────┐ ┌──────▼────────┐ ┌────────▼────────┐ ┌────────▼────────┐
│    Pillar I     │ │    Pillar II    │ │   Pillar III  │ │    Pillar IV    │ │     Pillar V    │
│  Transport-     │ │  Optimistic PCN │ │ Dual-Token Peg│ │ Cryptographic   │ │ Multi-Hop HTLC  │
│  Agnostic Mesh  │ │  Dispute Engine │ │ Seigniorage   │ │ Fiat Bridge UTR │ │ PCN Router      │
└─────────────────┘ └─────────────────┘ └───────────────┘ └─────────────────┘ └─────────────────┘
```

1. **Pillar I: Dual-Medium Transport-Agnostic Engine (DMT-Engine)**
   - Transmits Ed25519-signed transaction envelopes over standard TCP/IP networks or offline Bluetooth Low Energy (BLE) Mesh flooding without altering consensus rules or signature verification.
2. **Pillar II: Optimistic Sequence-Based State Channels & Challenge Window**
   - Bilateral off-chain channels utilizing monotonically increasing sequence numbers ($S \in \mathbb{N}$). Disputes initiate a 24-hour challenge window ($T_{\text{dispute}}$) on-chain; higher sequence co-signed states ($S' > S$) penalize cheat attempts.
3. **Pillar III: Dual-Token Algorithmic Seigniorage Stability Protocol**
   - Pegs the LQD stablecoin to ₹1.00 INR using an orderbook-driven feedback loop:
     - **Expansion ($P_{\text{LQD}} > \text{₹}1.005$)**: Protocol mints LQD and distributes seigniorage yield to **LQD-Shares (LQS)** stakers.
     - **Contraction ($P_{\text{LQD}} < \text{₹}0.995$)**: Protocol burns LQD by issuing discounted **LQD-Bonds (LQB)** redeemable at 1.15x face value upon expansion.
4. **Pillar IV: Non-Custodial Fiat Bridge with Cryptographic UTR Receipts**
   - Connects on-chain asset locks with real-world bank Unique Transaction References (UTRs). Releasing collateral requires a valid Ed25519 operator receipt attesting to fiat settlement.
5. **Pillar V: Multi-Hop Hashed Timelock Payment Routing (HTLC-PCN)**
   - Computes optimal Dijkstra routing paths across intermediate nodes, locking funds via secret hash pre-images $H(K)$ to facilitate non-custodial multi-hop payments.

---

## 📚 Modular Specifications Index

Comprehensive architectural specifications are located in the `docs/` directory:

- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — System layer architecture, component wiring, and end-to-end transaction lifecycles.
- [docs/PCN_SPEC.md](docs/PCN_SPEC.md) — State channel state machines, sequence co-signatures, and dispute timelock protocols.
- [docs/STABILITY.md](docs/STABILITY.md) — LQD/INR limit orderbook matching engine, seigniorage expansion/contraction, and yield math.
- [docs/FIAT_BRIDGE.md](docs/FIAT_BRIDGE.md) — Cryptographic UTR mappings, receipt verification, and fiat pool settlement workflows.
- [docs/TRANSPORT.md](docs/TRANSPORT.md) — Transport abstraction, TCP/IP driver, and Bluetooth Mesh flooding drivers.

---

## 📁 Repository Organization

```
cmd/
  node/            # Node daemon entrypoint (cmd/node/main.go)
  cli/             # Command-line interface client (cmd/cli/main.go)
configs/           # Multi-node YAML configurations (node_a.yaml, node_b.yaml)
docs/              # Modular architectural design specifications
internal/
  api/             # Chi REST API, Web UI dashboard template, & handlers
  block/           # Block schemas & hashing routines
  blockchain/      # On-chain repository & sequence rules
  config/          # YAML configuration parser
  consensus/       # Rotating leader & 51% majority voting engine
  crypto/          # Ed25519 key generation, signers, verifiers
  ledger/          # StateRepository & BadgerDB account state machine
  node/            # Core orchestrator & constructor dependency injection
  orderbook/       # LQD/INR limit orderbook matching engine
  pcn/             # State channels, Share/Bond models, settlement
  peer/            # Peer registry & discovery broadcasts
  receipt/         # Fiat bridge UTR mapping & receipt verification
  routing/         # Topology tracker & multi-hop router interface
  stability/       # Seigniorage engine, staking APY, & bond redemption
  storage/         # BadgerDB key-value database engine wrapper
  token/           # AssetManager (Mint, Burn, Credit, Debit)
  transaction/     # Signed transaction payloads
  transport/       # Envelope abstractions & transport drivers
    internet/      # TCP/IP transport driver
    bluetooth/     # BLE Mesh transport driver
  wallet/          # Key storage & file repository
```

---

## 🛠️ Build & Quickstart

### Prerequisites
- Go 1.22+ installed locally.

```bash
# Build node and cli binaries
make build

# Execute unit and integration tests with race detection
make test

# Start Node A (Port 8080)
make run-a

# Start Node B in a separate terminal (Port 8081)
make run-b
```

### Web Console Access
- Node A Control Panel: [http://localhost:8080/](http://localhost:8080/)
- Node B Control Panel: [http://localhost:8081/](http://localhost:8081/)

---

## 📡 REST API v1 Quick Reference

| Method | Endpoint | Description |
| --- | --- | --- |
| `GET` | `/api/v1/status` | System health, block height, peer count, balance |
| `GET` | `/api/v1/wallet/info` | Local wallet address & token balance |
| `POST` | `/api/v1/wallet/create` | Generate new wallet & mint genesis allocation |
| `POST` | `/api/v1/wallet/transfer` | Execute on-chain token transfer |
| `POST` | `/api/v1/channels/open` | Open state channel with counterparty |
| `POST` | `/api/v1/channels/update` | Co-sign off-chain channel sequence state update |
| `POST` | `/api/v1/channels/settle` | Initiate or finalize channel dispute settlement |
| `GET` | `/api/v1/orderbook/depth` | Query LQD/INR limit orderbook depth |
| `POST` | `/api/v1/orderbook/order` | Submit limit buy/sell order |
| `GET` | `/api/v1/stability/status` | Current price, peg state, APY %, total staked |
| `POST` | `/api/v1/stability/stake` | Stake LQD-Shares for seigniorage expansion yield |
| `POST` | `/api/v1/stability/buy-bonds` | Purchase discounted LQD-Bonds during contraction |
| `POST` | `/api/v1/stability/redeem-bonds`| Redeem LQD-Bonds during expansion phase |
| `POST` | `/api/v1/bridge/fiat-request` | Initiate fiat bridge request (locks crypto, returns UTR) |
| `POST` | `/api/v1/bridge/verify-receipt` | Submit signed operator receipt to settle fiat bridge |
