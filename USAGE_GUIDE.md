# LiquidLink Complete Operations & Validator Usage Guide

Welcome to the comprehensive operations guide for **LiquidLink** — the transport-agnostic, peer-to-peer off-chain payment network backed by an algorithmic seigniorage INR stablecoin (LQD), cryptographic UTR fiat bridges, and Bluetooth Low Energy (BLE) offline mesh execution.

---

## Table of Contents
1. [Architecture & System Model](#1-architecture--system-model)
2. [Standalone Pre-Compiled Binaries](#2-standalone-pre-compiled-binaries)
3. [Single Node Quickstart](#3-single-node-quickstart)
4. [Running Multi-Node Validator Consensus](#4-running-multi-node-validator-consensus)
5. [Two-Laptop Deployment (Separate Networks / Hotspot)](#5-two-laptop-deployment-separate-networks--hotspot)
6. [Web Bluetooth Offline Payments & Host Push Sync](#6-web-bluetooth-offline-payments--host-push-sync)
7. [Off-Chain State Channels (PCN) Workflow](#7-off-chain-state-channels-pcn-workflow)
8. [Algorithmic Seigniorage Engine & Peg Stability](#8-algorithmic-seigniorage-engine--peg-stability)
9. [Cryptographic Fiat Bridge & Automated UPI Simulation](#9-cryptographic-fiat-bridge--automated-upi-simulation)
10. [REST API Quick Reference](#10-rest-api-quick-reference)

---

## 1. Architecture & System Model

LiquidLink is designed to eliminate banking server bottlenecks and merchant discount rate (MDR) processing fees for micro-transactions:

```
+-----------------------------------------------------------------------------------+
|                                 WEB UI / DASHBOARD                                |
|  Non-Custodial WebCrypto Wallet (Ed25519) | Web Bluetooth API | SVG Price Chart  |
+-----------------------------------------+-----------------------------------------+
                                          | REST API / JSON (/api/v1)
+-----------------------------------------v-----------------------------------------+
|                                  REST API ROUTER                                  |
|                 Chi Router (api.go) | UPI Payout | Wallet & PCN Endpoints         |
+-------------------+---------------------+---------------------+-------------------+
                    |                     |                     |
+-------------------v---+   +-------------v-------+   +---------v-------------------+
| SEIGNIORAGE ENGINE    |   | PAYMENT CHANNEL NET |   | ORDERBOOK MATCHING ENGINE   |
| Price Peg: ₹1.00 INR  |   | State Channels (PCN)|   | Bids/Asks Limit Orderbook   |
| Expansion: Stake LQS  |   | Co-Signed Sequences |   | Escrow & Atomic INR Transfer|
| Contraction: LQB Bonds|   | Dispute Timelocks   |   | Price Discovery Ticker      |
+-------------------+---+   +-------------+-------+   +---------+-------------------+
                    |                     |                     |
+-------------------v---------------------v---------------------v-------------------+
|                          FIAT BRIDGE & RECEIPT REGISTRY                           |
|       Cryptographic UTR Verification | Bank Operator Ed25519 Signatures           |
+-----------------------------------------+-----------------------------------------+
                                          |
+-----------------------------------------v-----------------------------------------+
|                       BADGER DB LEDGER & STORAGE REPOSITORY                       |
|   storage.BadgerStore (WISCKEY KV) | Account Balances | Transaction History | Blocks|
+-----------------------------------------------------------------------------------+
```

---

## 2. Standalone Pre-Compiled Binaries

LiquidLink is written in pure Go and compiles to statically linked standalone binaries (`CGO_ENABLED=0`). Target systems do **not** require Go, GCC, or dynamic runtime libraries.

The repository includes pre-built static binaries in `bin/`:

| Binary File | Target OS / Architecture | Description |
| :--- | :--- | :--- |
| `bin/liquidlink` | Linux (Native) | Default node daemon for Linux machines |
| `bin/liquidlink-linux-amd64` | Linux x86_64 / AMD64 | Portable static binary for Ubuntu, Debian, Fedora, Arch, etc. |
| `bin/liquidlink-linux-arm64` | Linux ARM64 / AArch64 | Portable static binary for Raspberry Pi 4/5, Linux ARM laptops |
| `bin/liquidlink-darwin-arm64` | macOS ARM64 | Portable static binary for Apple Silicon (M1/M2/M3/M4 Macs) |
| `bin/cli` | Linux / CLI Tool | Command-line management tool |

### Building Fresh Binaries
If you have Go 1.22+ installed and make modifications, rebuild with:
```bash
make build       # Compiles bin/node and copies to bin/liquidlink
make build-all   # Compiles all multi-platform static binaries
```

---

## 3. Single Node Quickstart

### Step 1: Start the Daemon
```bash
./bin/liquidlink --config configs/config.yaml
```
Output:
```json
{"time":"2026-10-30T07:18:00.000Z","level":"INFO","msg":"generating new wallet for node","node_id":"node-1"}
{"time":"2026-10-30T07:18:00.002Z","level":"INFO","msg":"minted genesis allocation","address":"0x...","amount":100000}
{"time":"2026-10-30T07:18:00.003Z","level":"INFO","msg":"starting LiquidLink node","node_id":"node-1"}
{"time":"2026-10-30T07:18:00.004Z","level":"INFO","msg":"starting HTTP API Server","port":8080}
```

### Step 2: Open the Web Console
Visit:
```
http://localhost:8080
```
- A client-side non-custodial Ed25519 wallet is automatically created or loaded.
- Initial testnet balances (1,000 LQD and ₹1,000.00 INR) are funded for immediate interactive testing.

---

## 4. Running Multi-Node Validator Consensus

LiquidLink features an Ed25519 leader-based **PBFT Consensus Engine** (`internal/consensus`) with an automated block producer loop.

### Running Two Validator Nodes on One Machine

#### Terminal 1 (Node A — Leader / Validator on `:8080`):
```bash
make run-a
# Or directly:
./bin/liquidlink --config configs/node_a.yaml
```
Node A binds HTTP to port `8080` and P2P internet transport to port `9090`.

#### Terminal 2 (Node B — Validator on `:8081`):
```bash
make run-b
# Or directly:
./bin/liquidlink --config configs/node_b.yaml
```
Node B binds HTTP to port `8081` and P2P internet transport to port `9091`.

#### How Consensus Operates:
1. Every 5 seconds, the automated background block producer ticker (`startBlockProducer`) queries `blockchain.CurrentBlock()`.
2. The proposer generates `block.NewBlock(index, tip.Hash, ...)`.
3. The block is signed using the node's Ed25519 private key.
4. `consensus.ReceiveProposal()` validates block hash monotonicity, timestamp, and validator authority.
5. If the vote threshold (>50%) is satisfied, the block is appended to BadgerDB and block height increments across all nodes.

You can verify the live block height in the UI stat-strip or via curl:
```bash
curl http://localhost:8080/api/v1/blockchain/height
# Output: {"height":14}
```

---

## 5. Two-Laptop Deployment (Separate Networks / Hotspot)

To run LiquidLink across two different laptops (e.g. Laptop A on WiFi / Host network and Laptop B on a separate network or offline):

### Configuration File Setup

#### 1. On Laptop A (Host / Validator Node):
Copy `configs/node_a.yaml` and set `port: 8080`.
Find Laptop A's LAN IP address (e.g. `192.168.1.100` via `ip addr` or `ifconfig`).

Run on Laptop A:
```bash
./bin/liquidlink-linux-amd64 --config configs/node_a.yaml
```
Now Laptop A's API and Web UI are accessible on the network at `http://192.168.1.100:8080`.

#### 2. On Laptop B (Client / Offline Laptop):
Copy `bin/liquidlink-linux-amd64` (or `bin/liquidlink-darwin-arm64` if using a Mac) and `configs/node_b.yaml` to Laptop B via USB drive or Bluetooth file transfer:

In `configs/node_b.yaml`:
```yaml
node:
  id: "node-laptop-b"
  port: 8081
storage:
  path: "./data/laptop-b/badger"
transports:
  internet:
    enabled: true
    bind_address: "0.0.0.0:9091"
    bootstrap_peers:
      - "192.168.1.100:9090" # Laptop A's IP
  bluetooth:
    enabled: true
    device_id: "hci0"
    service_uuid: "00001101-0000-1000-8000-00805F9B34FB"
```

Run on Laptop B:
```bash
./bin/liquidlink-linux-amd64 --config configs/node_b.yaml
```
Open `http://localhost:8081` on Laptop B.

---

## 6. Web Bluetooth Offline Payments & Host Push Sync

One of LiquidLink's core capabilities is executing **zero-internet payments** between nearby devices over Bluetooth Low Energy (BLE), followed by batch cryptographic verification when reconnecting to the host.

### Step-by-Step Bluetooth Workflow

#### Step 1: Open the Bluetooth Console
On Laptop B (or in your browser), go to:
- Open `http://localhost:8081` (or your local dashboard).
- In the **State Channels** panel or in the **Profile Drawer**, click **"📡 BLE P2P Sync"** to open the Bluetooth P2P modal.

#### Step 2: Pair with Nearby BLE Device
1. Click **"📡 Pair Web Bluetooth Device"**.
2. Chrome / Edge will pop up the native Web Bluetooth device scanner.
3. Select Laptop A (or any Bluetooth device / phone).
4. The status badge will turn green: `Connected to <Device Name>`.
   *(Note: If Web Bluetooth is not supported on the browser or runs headless, LiquidLink automatically falls back to the node's local BLE mesh transport driver `hci0`).*

#### Step 3: Make an Offline Payment (Zero Internet)
1. Switch to the **"Make Offline Payment"** tab in the modal.
2. Select your payment medium:
   - **State Channel Payment (PCN Off-Chain):** Enter the Channel ID (e.g. from an existing bilateral channel with Host).
   - **Direct Signed Transfer (Ed25519):** Enter the Receiver's wallet address.
3. Enter the amount (e.g. `50 LQD`).
4. Click **"Sign & Queue Offline Payment ⚡"**.
5. What happens automatically:
   - The transaction sequence is incremented.
   - The payload is signed with your private key.
   - If a Web Bluetooth GATT connection is active, the payload is transmitted over the BLE characteristic.
   - The signed statement is saved into browser `localStorage` under `liquidlink_offline_queue_<port>`.
   - The payload is queued in the local daemon.

#### Step 4: Reconnect and Push to Host Node
When Laptop B reconnects to the network (or approaches Host Laptop A):
1. Switch to the **"Queued Tx"** tab in the Bluetooth modal.
2. Under **"Host Node Sync URL"**, verify or enter Host Laptop A's URL:
   ```
   http://192.168.1.100:8080
   ```
3. Click **"🚀 Push & Validate with Host"**.
4. The Host node receives the requests:
   - **For PCN Statements (`/api/v1/channels/offline-sync`):** Host verifies that sequence $N+1 > N$, checks total balance conservation invariant ($Bal_A + Bal_B == Deposit$), validates the Ed25519 signature, commits the new sequence to BadgerDB, and broadcasts peer sync.
   - **For Direct Transfers (`/api/v1/transport/broadcast-offline`):** Host cryptographically verifies the signature, debits sender, credits receiver, and syncs peer state.
5. All validated payments are cleared from the offline queue, and an instant celebration toast confirms:
   `"Successfully pushed & validated N payments on Host node! 🚀"`

---

## 7. Off-Chain State Channels (PCN) Workflow

Payment channels allow two parties to transact off-chain infinitely without incurring blockchain consensus or banking network MDR fees.

### Channel Lifecycle:

1. **Open Channel:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/channels/open \
     -H "Content-Type: application/json" \
     -d '{"counterparty":"0xrecipient_address","amount":500}'
   ```
   Locks 500 LQD collateral in escrow and sets `sequence: 1`.

2. **Off-Chain Balance Updates:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/channels/update \
     -H "Content-Type: application/json" \
     -d '{"channel_id":"<CHANNEL_ID>","amount":50,"sending":true}'
   ```
   Increments sequence to `2`, updates balances ($450 / 50$), and updates Ed25519 co-signatures.

3. **Dispute & Settlement Challenge:**
   Either party can initiate settlement:
   ```bash
   curl -X POST http://localhost:8080/api/v1/channels/settle \
     -H "Content-Type: application/json" \
     -d '{"channel_id":"<CHANNEL_ID>","action":"initiate"}'
   ```
   Starts the 24-hour dispute window (`StateClosing`). During this window, an honest party can submit a higher sequence co-signed statement via `/api/v1/channels/challenge` or `/api/v1/channels/offline-sync` to defeat old state attempts.

4. **Finalize Settlement:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/channels/settle \
     -H "Content-Type: application/json" \
     -d '{"channel_id":"<CHANNEL_ID>","action":"finalize"}'
   ```
   Locks are released and funds are credited to the parties on the primary ledger.

---

## 8. Algorithmic Seigniorage Engine & Peg Stability

LiquidLink maintains a **1:1 peg with the Indian Rupee (₹1.00 INR)** via an automated algorithmic seigniorage engine (`internal/stability`):

### Tri-Token Mechanics:
- **LQD (Liquid Stablecoin):** Circulating payment token targeted at ₹1.00 INR.
- **LQS (LQD-Shares / Equity):** Staked during expansion ($P_{\text{LQD}} > \text{₹}1.00$). Share stakers earn interest dividends minted from seigniorage expansion.
- **LQB (LQD-Bonds / Debt):** Issued during contraction ($P_{\text{LQD}} < \text{₹}1.00$). Users burn LQD at a 15% discount to receive bonds. When LQD returns to ₹1.00, bondholders redeem bonds for a 5% bonus profit.

### Interactive UI Tools:
- **Peg Equilibrium Progress Gauge:** Shows real-time peg proximity to ₹1.00.
- **Live Staking Yield Calculator:** Computes expected annual return based on current protocol APY.
- **Bond Profit Estimator:** Computes redemption yield when purchasing discounted bonds.

---

## 9. Cryptographic Fiat Bridge & Automated UPI Simulation

The Fiat Bridge links off-chain digital asset state channels with traditional Indian banking rails:

1. **Create UTR Bridge Request:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/bridge/fiat-request \
     -H "Content-Type: application/json" \
     -d '{"receiver":"0xmy_wallet","amount":500}'
   ```
   Returns a unique 16-character Bank UTR mapping (e.g. `UTR984bc28a...`).

2. **Verify Receipt & Settle:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/bridge/verify-receipt \
     -H "Content-Type: application/json" \
     -d '{"utr":"UTR984bc28a...","receipt_id":"rec-001"}'
   ```
   Verifies bank operator Ed25519 signature and credits user balance with 500 LQD.

3. **UPI Payout Webhook Simulation:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/bridge/upi-payout \
     -H "Content-Type: application/json" \
     -d '{"vpa":"merchant@upi","amount":500,"utr":"UTR984bc28a..."}'
   ```

---

## 10. REST API Quick Reference

| Endpoint | Method | Purpose |
| :--- | :--- | :--- |
| `/api/v1/status` | `GET` | Node health, balance, total supply, block height, peers |
| `/api/v1/wallet/create` | `POST` | Generate new Ed25519 wallet keypair |
| `/api/v1/wallet/balance?address=0x...` | `GET` | Query address token & fiat balance |
| `/api/v1/wallet/transfer` | `POST` | Execute transfer (`sender`, `receiver`, `amount`) |
| `/api/v1/wallet/sign-offline` | `POST` | Sign offline PCN statement or direct transaction with node key |
| `/api/v1/tx/broadcast` | `POST` | Broadcast pre-signed client transaction payload |
| `/api/v1/channels/open` | `POST` | Open new bilateral state channel |
| `/api/v1/channels/update` | `POST` | Update off-chain state sequence |
| `/api/v1/channels/settle` | `POST` | Initiate or finalize channel settlement |
| `/api/v1/channels/offline-sync` | `POST` | **Push & validate offline signed PCN statements** |
| `/api/v1/transport/bluetooth/status` | `GET` | Query BLE transport config & service UUID |
| `/api/v1/transport/broadcast-offline` | `POST` | **Broadcast & validate offline signed BLE payloads** |
| `/api/v1/orderbook/depth` | `GET` | Query limit orderbook bids and asks |
| `/api/v1/orderbook/order` | `POST` | Place limit buy/sell order |
| `/api/v1/stability/status` | `GET` | Seigniorage status, peg state, APY %, shares, bonds |
| `/api/v1/stability/stake` | `POST` | Stake LQD-Shares for expansion dividend yields |
| `/api/v1/stability/buy-bonds` | `POST` | Buy discounted bonds during contraction |
| `/api/v1/bridge/fiat-request` | `POST` | Generate bank UTR reference request |
| `/api/v1/bridge/verify-receipt` | `POST` | Cryptographic bank receipt verification & settlement |
| `/api/v1/fiat/pool` | `GET` | Total fiat bank reserve pool metrics |
| `/api/v1/blockchain/height` | `GET` | Current blockchain height |
| `/api/v1/blockchain/blocks` | `GET` | Recent PBFT committed chain blocks |
