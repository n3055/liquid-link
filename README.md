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
   ```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
