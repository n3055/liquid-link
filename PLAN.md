# PLAN: Architectural Explanation & System Design Specification for LiquidLink

## Context
This document serves as the comprehensive architectural specification and conceptual guide for **LiquidLink**, detailing its core mission, business value proposition (solving UPI MDR charges), distinction from CBDC, and deep technical implementation in code.

---

## 1. Executive Summary & One-Line Pitch

> **One-Line Pitch:**
> **LiquidLink is an open, transport-agnostic, peer-to-peer off-chain payment network backed by an algorithmic seigniorage INR stablecoin (LQD) and cryptographic UTR fiat bridges, enabling zero-fee instant micro-transactions without bank MDR charges.**

### Core Purpose
LiquidLink connects digital asset state channels with traditional fiat banking rails. It allows users and merchants to transact off-chain instantly over any network medium (TCP/IP, Bluetooth Mesh) with 1:1 INR peg stability, settling with fiat bank reference numbers (UTR) and automated UPI payouts.

---

## 2. How LiquidLink Solves UPI MDR Charges

### The UPI MDR Problem
1. **Merchant Discount Rate (MDR):** Traditional card networks and payment gateways charge merchants 0.5%–2.5% per transaction.
2. **Infrastructure & Gateway Costs:** While basic P2P UPI in India is subsidized, P2M (Person-to-Merchant) over credit-on-UPI, PPI wallets, payment aggregators, and cross-border gateways impose significant interchange fees, gateway margins, and server bandwidth overheads.
3. **Micro-transaction Bottleneck:** Sending thousands of ₹10–₹100 transactions directly through bank core banking servers (CBS) leads to server congestion, rate limits, and non-zero processing costs per transaction for financial institutions.

### LiquidLink's Solution
1. **Zero-Fee Off-Chain Payments (PCN):** Payments between buyers and sellers execute inside bilateral off-chain State Channels (`internal/pcn`). Co-signed state sequence updates occur locally in micro-seconds with **0 MDR fees** and **0 blockchain gas fees**.
2. **Net Batch Settlement:** Instead of bombarding bank servers with 1,000 separate micro-transactions, LiquidLink aggregates off-chain transaction state and executes a single net settlement only when opening/closing state channels or rebalancing fiat reserves.
3. **Intermediary Elimination:** LQD stablecoins maintain a 1:1 algorithmic peg with INR (₹1.00), bypassing card networks, acquiring banks, and payment aggregators that extract MDR fees.

---

## 3. How LiquidLink Differs from CBDC (e₹ / Digital Rupee)

| Feature | CBDC (e₹ / Digital Rupee) | LiquidLink |
| :--- | :--- | :--- |
| **Control & Governance** | Centralized & strictly permissioned by the Reserve Bank of India (RBI). | Decentralized / Multi-validator consensus running leader-based PBFT (`internal/consensus`). |
| **Monetary Policy** | Central Bank fiat printing; manual issuance by central authority. | **Algorithmic Seigniorage Engine** (`internal/stability`): Automated expansion/contraction via LQS shares and LQB bonds. |
| **Privacy & Custody** | Centralized ledger tracks every account; banks/RBI can freeze accounts. | **Non-Custodial Client-Side Keys**: Ed25519 keypairs generated in browser via WebCrypto API; user holds private key. |
| **Transport & Offline** | Requires central bank server connection or specialized bank HSM hardware. | **Transport Agnostic & BLE Mesh** (`internal/transport/bluetooth`): Runs offline over Bluetooth Low Energy mesh networking without internet/bank connectivity. |
| **Bridge & Interoperability** | Siloed in central bank system. | Native **Cryptographic Fiat Bridge** (`internal/receipt`, `internal/fiat`) with automated UTR verification and UPI gateway integration. |

---

## 4. Deep Code-Level Architectural Breakdown

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

### A. Storage Engine: What is BadgerDB & How it Works
- **What is BadgerDB?** An embeddable, ultra-fast Key-Value (KV) database written in pure Go by Dgraph, implementing WISCKEY architecture (separating keys from values to minimize SSD write amplification).
- **Implementation (`internal/storage`, `internal/ledger`):**
  - Wrapped via `storage.BadgerStore` providing thread-safe ACID `View` (read-only) and `Update` (read-write) transactions.
  - Key namespaces:
    - `ledger:account:<address>` -> Serialized JSON `AccountState` (Balance, Nonce).
    - `pcn:channel:<channel_id>` -> Serialized JSON `Channel` state.
    - `receipt:id:<id>` & `receipt:utr:<utr>` -> Verified bridge receipts and bank UTR mappings.
    - `block:height:<index>` & `block:hash:<hash>` -> Chain blocks.
  - **Read-Write Caching Layer:** `ledgerImpl` uses a read-write mutex `sync.RWMutex` with an in-memory `map[crypto.Address]*AccountState` for microsecond balance lookups, persisting state to BadgerDB on writes.

### B. Validation & Consensus Mechanics
- **Validation Layers (`internal/consensus`, `internal/blockchain`, `internal/crypto`):**
  1. **Transaction Validation:** Verifies sender address Ed25519 signature against payload hash, checks account nonce (anti-replay), and ensures sender balance $\ge$ transfer amount.
  2. **Block Proposal Validation:** Checks block height (`Index == tip.Index + 1`), previous hash link (`PrevHash == tip.Hash`), Merkle tree root hash matching, timestamp monotonicity, and proposer authority against validator key map.
  3. **PBFT Vote Tallying:** Leader proposes block $\rightarrow$ Nodes evaluate rules $\rightarrow$ Nodes sign Ed25519 `VoteAccept`/`VoteReject` messages $\rightarrow$ When vote count meets threshold percentage (>66%), block is committed to BadgerDB blockchain and ledger state updates are finalized.

### C. Orderbook Matching Engine
- **Implementation (`internal/orderbook`):**
  - In-memory limit orderbook tracking `orders map[string]*Order`, bids (buy), and asks (sell).
  - **Escrow / Balance Reservation:**
    - Buy order: Immediately reserves required INR from buyer's fiat account (`fiatMgr.ReserveINR`).
    - Sell order: Verifies seller has sufficient LQD balance (`assetMgr.BalanceOf`).
  - **Matching Algorithm:**
    - **Price-Time Priority:** Matches incoming buy orders against lowest sell asks (`Price <= buyPrice`), sorted by price ascending, then creation timestamp.
    - **Maker Price Execution:** Trade executes at seller's limit price.
    - **Atomic Settlement:** Simultaneously debits/credits LQD asset manager balances AND transfers reserved INR fiat between buyer and seller in a single atomic step (`fiatMgr.TransferINR`).
    - **Price Discovery Ticker:** Executed trade price updates `lastPrice`, feeding real-time price signals into the Seigniorage Stability Engine (`stability.SetMarketPrice`).

### D. Minting, Stability Engine & Converting to Bonds/Shares
- **Implementation (`internal/stability`, `internal/token`):**
  - **Tri-Token Economic Model:**
    - **LQD (Liquid Stablecoin):** Payment medium targeted at ₹1.00 INR.
    - **LQS (LQD-Shares / Equity):** Staked during expansion ($P_{\text{LQD}} > \text{₹}1.00$). Earns APY interest yields from seigniorage minting rewards.
    - **LQB (LQD-Bonds / Debt):** Issued during contraction ($P_{\text{LQD}} < \text{₹}1.00$).
  - **Contraction Mechanism (Price < ₹1.00):**
    - Protocol needs to reduce LQD circulating supply to raise price back to ₹1.00.
    - Users call `BuyBonds(lqdAmount)`.
    - Protocol burns `lqdAmount` of LQD from user balance and issues discounted `LQB` bonds at a 1.15x ratio (e.g., 100 LQD $\rightarrow$ 115 LQB).
    - Supply contracts, restoring peg upward.
  - **Expansion Mechanism (Price > ₹1.00):**
    - Protocol needs to increase LQD supply to bring price down to ₹1.00.
    - Bondholders call `RedeemBonds(bondAmount)`: Redeems 1 LQB for 1 LQD + 5% expansion yield bonus, minting new LQD.
    - Remaining expanded seigniorage is distributed as staking APY dividends to LQS share stakers (`StakeShares`).

### E. Off-Chain Payment Channels (PCN)
- **Implementation (`internal/pcn`):**
  - Bilateral state channels between Node A and Node B.
  - Opening: Parties lock initial LQD deposit on-chain / on-ledger.
  - Off-chain transfers: Parties exchange co-signed `StateStatement` JSON objects with incrementing sequence numbers (`Sequence: N+1`). Balance split updates off-chain instantaneously with 0 gas/fees.
  - Settlement: Either party submits the highest sequence co-signed state to `SettleChannel()`. After challenge timelock, final balances are unlocked and updated on the main ledger.

### F. Bluetooth Mesh Role (BLE Offline Payments)
- **Implementation (`internal/transport/bluetooth`, Web Bluetooth API):**
  - Transport abstraction `transport.Transport` interface implemented by `MeshTransport`.
  - Deduping mechanism: Maintains `seenPackets map[string]time.Time` keyed by SHA-256 hash of payload (`SenderID|Type|Hash`). Prevents broadcast flooding loops.
  - Offline flow: Customer browser signs Ed25519 transaction offline $\rightarrow$ Web Bluetooth API (`navigator.bluetooth`) broadcasts packet to nearby merchant/peer BLE devices $\rightarrow$ Merchant node receives, verifies signature offline, caches payload, and broadcasts packet to network when internet connectivity resumes.

### G. Cryptographic Fiat Bridge & UPI Payout Gateway
- **Implementation (`internal/receipt`, `internal/fiat`, `internal/api`):**
  - Generates unique Bank UTR reference mapping (`UTRMapping`).
  - Cryptographic receipt verification: `VerifyReceipt()` checks SHA-256 hash `ID|UTR|Address|Amount|Timestamp` signed by bank operator Ed25519 key.
  - UPI payout gateway simulation (`/api/v1/bridge/upi-payout`) connects verified UTR receipts with INR reserve pool payouts to recipient Virtual Payment Addresses (VPA, e.g., `user@upi`).

---

## 5. Tech Stack Summary

- **Core Backend Language:** Go 1.22+
- **Database / Ledger Storage:** BadgerDB v4 (`github.com/dgraph-io/badger/v4`)
- **Cryptography & Key Management:** Ed25519 (`crypto/ed25519` in Go & WebCrypto API in JS)
- **REST API Engine:** Chi Router (`github.com/go-chi/chi/v5`)
- **Web UI & Dashboard:** HTML5, CSS3, Vanilla JavaScript (zero external JS bundle dependencies), Web API, Web Bluetooth API (`navigator.bluetooth`)
- **Configuration & Build:** YAML (`gopkg.in/yaml.v3`), Makefile

---

## Files to modify
*Note: This is an architectural specification and system explanation document requested by the user.*
- `PLAN.md` (Updated with complete system architecture and business logic details)

## Reuse
- Existing Go package structure (`internal/ledger`, `internal/stability`, `internal/orderbook`, `internal/pcn`, `internal/transport`, `internal/receipt`, `internal/fiat`, `internal/consensus`).

## Steps
- [x] **Step 1: Document Executive Summary & One-Line Pitch**
- [x] **Step 2: Explain UPI MDR Charge Solution Logic**
- [x] **Step 3: Detail CBDC vs LiquidLink Key Differentiators**
- [x] **Step 4: Provide In-Depth Code-Level Subsystem Breakdown (BadgerDB, Validation, Orderbook, Seigniorage Engine, PCN, BLE Mesh, Fiat Bridge)**
- [x] **Step 5: Document Complete Tech Stack**

## Verification
- Code references checked and verified against `internal/` codebase implementation files (`storage.go`, `ledger_impl.go`, `orderbook_impl.go`, `stability_impl.go`, `pcn_impl.go`, `bluetooth_impl.go`, `receipt_impl.go`, `consensus_impl.go`).
