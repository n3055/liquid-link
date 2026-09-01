# LiquidLink

LiquidLink is a high-performance payment network and crypto-to-fiat bridge built on top of a Payment Channel Network (PCN) with an Algorithmic Seigniorage Stability Engine and non-custodial browser wallet capabilities.

The project provides a **transport-agnostic payment network** capable of transferring digital assets across multiple communication mediums (Internet TCP/IP, Bluetooth Mesh) and settling fiat transactions cleanly with local bank reference numbers (UTR) and automated UPI payout webhooks.

---

## 📚 Documentation & Guides

- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — System layer architecture, component wiring, and end-to-end transaction lifecycles.
- [docs/VALIDATOR_GUIDE.md](docs/VALIDATOR_GUIDE.md) — Multi-validator consensus setup, leader rotation, and operational guide.
- [docs/DEMO_GUIDE.md](docs/DEMO_GUIDE.md) — Step-by-step interactive Web UI walkthrough and feature demonstration manual.
- [docs/PCN_SPEC.md](docs/PCN_SPEC.md) — State channel state machines, sequence co-signatures, and dispute timelock protocols.
- [docs/STABILITY.md](docs/STABILITY.md) — LQD/INR limit orderbook matching engine, seigniorage expansion/contraction, and yield math.
- [docs/FIAT_BRIDGE.md](docs/FIAT_BRIDGE.md) — Cryptographic UTR mappings, receipt verification, and fiat pool settlement workflows.
- [docs/TRANSPORT.md](docs/TRANSPORT.md) — Transport abstraction, TCP/IP driver, and Bluetooth Mesh flooding drivers.

---

## Key Features

- **Non-Custodial Client Wallet:** Client-side Web Crypto key generation and cryptographic Ed25519 transaction signing performed directly in the browser (MetaMask / Phantom style).
- **Algorithmic Seigniorage Engine (LQD / Shares / Bonds):**
  - **LQD (Liquid Stablecoin):** Payment token targeted at 1 LQD = ₹1.00 INR.
  - **LQD-Shares (LQS):** Equity tokens staked by users to capture protocol seigniorage interest payouts during expansion ($>\text{₹}1.00$).
  - **LQD-Bonds (LQB):** Discounted bonds issued during market contraction ($<\text{₹}1.00$). Users burn LQD to buy bonds and earn high interest yield when LQD restores peg.
- **Transport Agnostic:** The exact same signed transaction can be transferred over any supported transport layer (HTTP, P2P, Bluetooth Mesh).
- **Off-Chain State Channels (PCN):** Zero-latency bilateral state channels using co-signed state statements and dispute challenge windows.
- **Crypto-to-Fiat Bridge:** UTR mapping, cryptographic receipt verification, and automated UPI payout gateway endpoints.
- **Permissioned Validator Consensus:** Leader-based PBFT consensus engine with rotating validators and Ed25519 vote tallies.
- **Persistent Key Storage:** Automatic node keypair persistence in local disk storage (`data/<node>/wallets/`).

---

## Repository Structure

The project follows a **Clean Architecture** structure:

- `cmd/`: Node daemon (`cmd/node/main.go`) and CLI client (`cmd/cli/main.go`).
- `configs/`: Node YAML configurations (`node_a.yaml`, `node_b.yaml`).
- `internal/`:
  - `api/`: REST HTTP router, endpoints, dashboard HTML/JS template, and UPI payout handlers.
  - `blockchain/`: On-chain block repository, block structures, and chain validation.
  - `consensus/`: Leader selection, PBFT voting engine, and vote validation.
  - `crypto/`: Ed25519 key generation, signers, verifiers, and address derivation.
  - `ledger/`: BadgerDB transaction ledger and account balance state machine.
  - `node/`: Dependency injection container and node lifecycle daemon.
  - `pcn/`: Off-chain payment channel manager, state statements, and settlement bonds.
  - `peer/`: P2P registry and node discovery.
  - `receipt/`: Fiat bridge receipt verification and UTR mapping.
  - `stability/`: Algorithmic seigniorage engine, staking APY yields, and bond redemptions.
  - `token/`: Asset manager (Mint, Burn, Credit, Debit).
  - `wallet/`: File-backed key repository and local wallet generation.

---

## Development Setup & Running Multi-Node System

### Prerequisites

- Go 1.22+ installed locally.

### Commands

```bash
# 1. Build binaries (bin/node, bin/cli)
make build

# 2. Run unit & integration tests
make test

# 3. Launch Node A (Port 8080)
make run-a

# 4. Launch Node B in a separate terminal (Port 8081)
make run-b
```

---

## Web UI Dashboard & Client Wallet

Open your browser to test the interactive LiquidLink Console:

- **Node A Dashboard:** [http://localhost:8080/](http://localhost:8080/)
- **Node B Dashboard:** [http://localhost:8081/](http://localhost:8081/)

### Dashboard Features

- **Client-Side Non-Custodial Wallet:** Generate, import, or export Ed25519 private keys directly in your browser. All transactions are signed client-side before transmission.
- **Dynamic Cartoon Identicons:** 100% unique, deterministic pixel art character avatars generated from wallet addresses.
- **Real-Time Price & Peg Chart:** Interactive SVG price history chart tracking LQD vs ₹1.00 INR.
- **Staking & Stability Dashboard:** Stake LQD-Shares to earn APY interest dividends, or buy/redeem LQD-Bonds during contraction/expansion.
- **Payment Channel Network (PCN):** Open bilateral state channels, execute off-chain sequence updates, and settle on-chain with dispute challenge timelocks.
- **Fiat Bridge & UPI Payout Simulation:** Generate UTR bridge requests, verify cryptographic receipts, and simulate automated UPI payout webhooks.

---

## REST API Endpoints (`/api/v1`)

| Endpoint | Method | Description |
| --- | --- | --- |
| `/api/v1/status` | GET | Node status, metrics, block height, and peer count |
| `/api/v1/wallet/info` | GET | Wallet address and local balance |
| `/api/v1/wallet/balance` | GET | Query account balance (`?address=...`) |
| `/api/v1/wallet/transfer` | POST | Direct token transfer (`receiver`, `amount`) |
| `/api/v1/wallet/mint` | POST | Mint token supply (`address`, `amount`) |
| `/api/v1/tx/broadcast` | POST | Broadcast pre-signed client transaction payload |
| `/api/v1/channels/list` | GET | List registered state channels |
| `/api/v1/channels/open` | POST | Open new state channel (`counterparty`, `amount`) |
| `/api/v1/channels/update` | POST | Update off-chain state (`channel_id`, `amount`, `sending`) |
| `/api/v1/channels/settle` | POST | Initiate/finalize channel settlement (`channel_id`, `action`) |
| `/api/v1/stability/status` | GET | Peg status, APY %, total staked shares, outstanding bonds |
| `/api/v1/stability/history` | GET | Price history data vs INR |
| `/api/v1/stability/stake` | POST | Stake LQD-Shares for interest yields |
| `/api/v1/stability/buy-bond` | POST | Buy discounted LQD-Bonds during contraction |
| `/api/v1/stability/redeem-bond` | POST | Redeem LQD-Bonds for LQD + interest during expansion |
| `/api/v1/bridge/fiat-request` | POST | Initiate fiat bridge request (`receiver`, `amount`) |
| `/api/v1/bridge/utr/{utr}` | GET | Fetch UTR mapping details |
| `/api/v1/bridge/verify-receipt` | POST | Verify receipt & settle request (`utr`, `receipt_id`) |
| `/api/v1/bridge/upi-payout` | POST | Automated UPI payout gateway integration / simulation |

---

## TODO & Roadmap

- [x] Multi-node YAML configurations and Makefile targets (`run-a`, `run-b`).
- [x] Persistent Ed25519 wallet key storage on disk.
- [x] Inter-node peer balance credit synchronization.
- [x] Deterministic pixel art cartoon identicon generator (MetaMask/Phantom style).
- [ ] Client-side non-custodial browser wallet (Web Crypto API & client JS signing).
- [ ] Algorithmic Seigniorage Stability Engine (`internal/stability`).
- [ ] Staking yield APY calculation and bond discount/redemption logic.
- [ ] Interactive SVG real-time price history chart in Web UI.
- [ ] Live notification feed drawer for mints, transfers, yields, and bridge events.
- [ ] Integration of real centralized UPI payout server SDK / Webhook payload verification.

---

## License

This project is licensed under the MIT License.
