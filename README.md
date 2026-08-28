# LiquidLink

LiquidLink is a research prototype and educational decentralized crypto-to-fiat bridge built on top of a Payment Channel Network (PCN). 

The goal of this project is to construct a **transport-agnostic payment network** capable of securely transferring custom digital assets across multiple communication mediums, such as standard Internet (TCP/IP) connections and local Bluetooth Mesh networks, without affecting transaction logic or consensus protocols.

## Key Features

- **Transport Agnostic:** The exact same signed transaction can be transferred over any supported transport.
- **Off-Chain State Channels:** Built-in support for state-channels utilizing both **Share** and **Bond** models.
- **Crypto-to-Fiat Bridge:** Cryptographic proofs and Unique Transaction Reference (UTR) mapping to clear fiat settlements.
- **Permissioned Consensus:** Decentralized validation ledger using simple majority-voting consensus and rotating leaders.

## Repository Structure

The project follows a **Clean Architecture** style:

- `cmd/`: Node daemon and CLI client entrypoints.
- `configs/`: YAML node configurations.
- `internal/`: Decoupled packages implementing core cryptography, consensus, routing, channels, ledger storage, API controllers, and transport adapters.
- `pkg/`: Shared utility packages.

See [CLAUDE.md](CLAUDE.md) for the complete engineering handbook and architectural specification.

## Development Setup

### Prerequisites

- Go 1.26+ installed locally.

### Getting Started

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd liquid-link
   ```

2. **Build the project:**
   ```bash
   make build
   ```

3. **Run local unit tests:**
   ```bash
   make test
   ```

4. **Launch the node daemon:**
   ```bash
   make run
   # Or run manually:
   # ./bin/liquid-node -config configs/config.yaml
   ```

## Web UI Dashboard

Once the node daemon is running, open your browser to access the interactive LiquidLink Console:

- **Dashboard URL:** [http://localhost:8080/](http://localhost:8080/) or [http://localhost:8080/ui](http://localhost:8080/ui)

### Web Dashboard Features

- **Real-Time Node Metrics:** View Node ID, Port, Wallet Address, Token Balance, Total Token Supply, Block Height, Connected Peers, and Node Health.
- **Token Operations:** 
  - **Transfer Tokens:** Send tokens directly to any counterparty address.
  - **Mint Tokens:** Issue new token supply to specified target addresses or the node's default wallet.
- **Payment Channel Network (PCN):**
  - **Open State Channels:** Deposit collateral and establish bilateral off-chain payment channels with counterparties.
  - **Update State:** Adjust channel balance allocations (sending/receiving) and increment sequence numbers.
  - **Settle Channels:** Initiate dispute/cooperative settlement or finalize channel closure.
- **Fiat Settlement Bridge:**
  - **Bridge Requests:** Lock crypto assets to generate Unique Transaction References (UTRs) for fiat clearing.
  - **Receipt Verification:** Verify cryptographic receipts and finalize fiat-to-crypto bridge requests.
- **Blockchain Explorer:** Inspect recent blocks, block hashes, nonces, timestamps, and transaction counts.

## REST API Endpoints

The node daemon also exposes JSON endpoints under `/api/v1`:

| Endpoint | Method | Description |
| --- | --- | --- |
| `/api/v1/status` | GET | Node status, metrics, block height, and peer count |
| `/api/v1/wallet/info` | GET | Wallet address and local balance |
| `/api/v1/wallet/balance` | GET | Query account balance (`?address=...`) |
| `/api/v1/wallet/transfer` | POST | Transfer tokens (`receiver`, `amount`) |
| `/api/v1/wallet/mint` | POST | Mint tokens (`address`, `amount`) |
| `/api/v1/channels/list` | GET | List registered state channels |
| `/api/v1/channels/{id}` | GET | Get state channel details |
| `/api/v1/channels/open` | POST | Open new state channel (`counterparty`, `amount`) |
| `/api/v1/channels/update` | POST | Update channel state (`channel_id`, `amount`, `sending`) |
| `/api/v1/channels/settle` | POST | Initiate/finalize settlement (`channel_id`, `action`) |
| `/api/v1/bridge/fiat-request` | POST | Initiate fiat bridge request (`receiver`, `amount`) |
| `/api/v1/bridge/utr/{utr}` | GET | Fetch UTR mapping details |
| `/api/v1/bridge/verify-receipt` | POST | Verify receipt & settle request (`utr`, `receipt_id`) |
| `/api/v1/blockchain/height` | GET | Current block height and tip block |
| `/api/v1/blockchain/blocks` | GET | List recent blockchain blocks |
| `/api/v1/peers` | GET | List registered network peers |
| `/api/v1/peers/register` | POST | Register network peer (`id`, `endpoint`, `address`) |

## License

This project is licensed under the MIT License - see the LICENSE file for details.
