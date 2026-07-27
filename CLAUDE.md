# CLAUDE.md — LiquidLink Engineering Handbook & Architecture Guide

This handbook serves as the permanent system architecture document, coding guidelines, development workflow tracker, and contributor guide for the **LiquidLink** repository.

---

## 1. Project Vision
LiquidLink is a research prototype and educational decentralized payment channel network (PCN) built to bridge crypto assets directly to fiat networks across transport-agnostic channels. The core philosophy is to support communication over diverse channels—from high-bandwidth Internet connections to low-bandwidth, local Bluetooth Mesh networks—without affecting the core transaction logic or consensus protocols.

## 2. Goals
- **Transport Agnosticism:** Complete decoupling of payment validation/signing logic from the underlying transmission layer.
- **Off-Chain Speed:** Low-latency payment routing using state channels (Share/Bond models).
- **Crypto-to-Fiat Bridging:** Verification of Unique Transaction References (UTRs) and cryptographic proof receipts to bridge on-chain assets with real-world fiat gateways.
- **Robust Permissioned Consensus:** Reliable block validation and majority voting consensus among known Payment Service Provider (PSP) nodes.
- **High Readability:** Simple, clean, decoupled Go code designed as an educational model.

## 3. Non-Goals
- **Permissionless Mainnet:** LiquidLink is not a permissionless public blockchain (e.g., Bitcoin, Ethereum). There is no proof-of-work or proof-of-stake.
- **High-Frequency Trading System:** Not optimized for microsecond-level trading execution.
- **Production Fiat Gateway:** The fiat settlement layer is simulated/mocked via receipts and UTR mapping, rather than integration with active commercial banking APIs.

## 4. Architecture Overview
The architecture is structured to flow linearly from identity keys to the physical transport layer.

```
Wallet
  ↓
Signed Transaction
  ↓
Ledger
  ↓
Blockchain
  ↓
Payment Channel Network (PCN)
  ↓
Routing
  ↓
Transport (Internet / Bluetooth)
  ↓
Settlement (Fiat Bridge)
```

## 5. Layered Architecture
1. **Application/UI Layer (`cmd/`, `internal/api/`):** Exposes JSON-RPC/REST APIs and CLI tools.
2. **Domain/Orchestration Layer (`internal/node/`):** Manages component lifecycles, service bootups, and wiring.
3. **Logic Layer (`internal/pcn/`, `internal/routing/`, `internal/consensus/`):** Implements transaction validation, channel balances, routing algorithms, and voting mechanisms.
4. **Data & Storage Layer (`internal/storage/`, `internal/blockchain/`, `internal/ledger/`):** Encapsulates BadgerDB operations, block storage, and ledger state transitions.
5. **Transport Layer (`internal/transport/`):** Transmits envelope messages agnostic of their contents.

## 6. Package Responsibilities
- **`cmd/`**: Entrypoints for node daemon (`node/`) and command-line client (`cli/`).
- **`internal/api`**: REST/HTTP routes using the Chi router for checking status, wallets, channels, and bridge settlements.
- **`internal/block`**: Defines `Block` and `Header` schema, serialization, and hashes.
- **`internal/blockchain`**: Implements block sequence verification, chain rules, and persistence wrappers.
- **`internal/config`**: Configuration structure (`config.yaml`) parser.
- **`internal/consensus`**: Handles voting decisions (`ACCEPT`/`REJECT`), block proposing, and leader election.
- **`internal/crypto`**: Ed25519 key derivation, signature creation, and validation.
- **`internal/ledger`**: Tracks account balances, nonces, and transaction execution state transitions.
- **`internal/node`**: Central coordinator bootstrapper that configures storage and runs tasks.
- **`internal/pcn`**: Orchestrates state channels, holds collateral balances, and implements share/bond models.
- **`internal/peer`**: Handles peer registries, discovery broadcasts, and liveness signals.
- **`internal/receipt`**: Links fiat settlement UTRs to crypto transactions and verifies receipts.
- **`internal/routing`**: Finds optimal multi-hop paths across the network nodes.
- **`internal/storage`**: Implements basic BadgerDB storage engines.
- **`internal/token`**: Logic for token representation, minting, and burning.
- **`internal/transaction`**: Houses the transaction model, nonces, and signature verification.
- **`internal/transport`**: Provides transport interfaces and universal message envelopes.
- **`internal/transport/internet`**: IP/TCP transport driver.
- **`internal/transport/bluetooth`**: Bluetooth Mesh transport driver.
- **`internal/wallet`**: Handles local key storage and transaction signing.

## 7. Directory Structure
```
cmd/
  node/         # Node daemon main file
  cli/          # CLI utility main file
configs/        # YAML configuration files
docs/           # Design specifications and documentation
internal/
  api/          # REST API endpoints
  block/        # Block data structures
  blockchain/   # Chain management
  config/       # Config loader
  consensus/    # Leader election & majority voting
  crypto/       # Cryptographic keys and signing
  ledger/       # Balance state and ledger account models
  node/         # Service orchestrator
  peer/         # Discovery and peer records
  pcn/          # State channels (share/bond)
  receipt/      # Fiat bridge receipts & UTR mapping
  routing/      # Multi-hop PCN routing
  storage/      # BadgerDB wrapper
  token/        # Asset limits, minting/burning
  transaction/  # Signed transaction definitions
  transport/    # Agnostic transport interfaces
    internet/   # TCP/IP implementation
    bluetooth/  # Bluetooth Mesh implementation
  wallet/       # Wallet keys storage
pkg/            # Shared helper packages
scripts/        # Automation / deployment helper scripts
test/           # System / integration test suites
data/           # BadgerDB local data directories (ignored by git)
```

## 8. Coding Standards
- **Clean Architecture:** Keep packages highly decoupled. Internal components must rely on interfaces, never on concrete implementations of other packages.
- **SOLID Principles:** Favor single responsibility classes/functions. Open/Closed extension.
- **Dependency Injection:** Inject all dependencies via structures constructors. No global states or singletons.
- **Context Handling:** Every function performing I/O, database access, or network communication must accept `context.Context` as the first argument.

## 9. Go Style Guide
- **Go Version:** Go 1.26
- **Formatting:** Code must be formatted using `gofmt -s -w .` before commit.
- **Exported Entities:** All public types, functions, structs, and interfaces must have comprehensive GoDoc comments.
- **Size Limits:** Files must be kept short, ideally under 300 lines. Split logical tasks into multiple files when necessary.

## 10. Naming Conventions
- **Interfaces:** Standard naming like `Reader`, `Writer`, `Validator`, `Registry`. Avoid suffixing with `Interface`.
- **Variables:** Use camelCase. Keep scope-localized variables short (e.g. `ctx`, `tx`, `err`).
- **File Names:** All lower_case snake_case (e.g., `state_channel.go`).

## 11. Error Handling
- Never ignore errors.
- Wrap errors with logical context using `fmt.Errorf("failed to do action: %w", err)`.
- Custom errors should be defined near their usage, permitting `errors.Is` comparison.

## 12. Logging
- Use standard `log/slog` for structured, leveled logging.
- Avoid logging sensitive user secrets (private keys, personal info).
- Log at appropriate levels:
  - `DEBUG`: Verbose message roundtrips, database hits.
  - `INFO`: Startups, shutdowns, channel openings, block additions.
  - `WARN`: Peer dropouts, retries, minor validation failures.
  - `ERROR`: System crashes, DB failures, signature fraud.

## 13. Testing Philosophy
- **Unit Tests:** Run unit tests via `go test -v -race ./...`. Ensure every package has localized test coverage.
- **Fuzzing & Mocking:** Implement interface mocks for network and storage to test edge scenarios.
- **Integration Tests:** Use `test/` directory to coordinate end-to-end multi-node simulations.

## 14. Dependency Injection
- Always use **Constructor Injection**:
  ```go
  func NewService(db Repository, net Transport) *Service {
      return &Service{db: db, net: net}
  }
  ```

## 15. Repository Pattern
- Decouple data access logic. Define repository interfaces in the domain package (e.g., `blockchain.Repository` or `ledger.StateRepository`) and implement them in the infrastructure layer (`internal/storage`).

## 16. Transport Abstraction
- The core payment and state transition logic must not know how a package travels.
- A `transport.Message` represents the universal serializable wrapper.
- All drivers (Internet, Bluetooth Mesh) implement `transport.Transport`.

## 17. Consensus Philosophy
- Permissioned, known validator set (PSP nodes).
- Round-robin/time-based **Leader Rotation** proposing blocks.
- **Majority Voting (51%+)** of active peers commits blocks canonical to the local ledger copy.

## 18. Blockchain Design
- Hash-chained sequence where each block references the previous block's SHA-256 hash.
- Consists of transaction batches signed by sending accounts.
- The block proposer signs the header hash; voting nodes sign the consensus ballot.

## 19. Transaction Lifecycle
1. User creates transaction and signs it using the Ed25519 Private Key.
2. The transaction is submitted via the node HTTP REST endpoint.
3. Node validates signature, nonce, and sender balance.
4. Transaction is held in local pool.
5. The rotated leader packages transactions into a block proposal.
6. Peers validate and vote.
7. Upon majority agreement, block is appended and account balances updated.

## 20. Wallet Design
- Local wallet stores cryptographic identity keys.
- Operates disconnected from the main loop; key signatures are produced locally and only signed payloads are sent to nodes.

## 21. Payment Channel Design
- **Share Model:** Establishes local ledger balance splits: `BalanceA` and `BalanceB`.
- **Bond Model:** Locks token collateral within a contract/registry to secure state channels.
- Bilateral signatures authorize state revisions off-chain.
- Dispute settlement period guarantees dispute protection.

## 22. Token Model
- Assets exist as a fixed-supply or minter-controlled digital token.
- `Amount` is measured as unsigned integers to prevent floating-point errors.
- Token generation is regulated via Authorized Issuer Nodes issuing `Mint`/`Burn` events.

## 23. Receipt Model
- Cryptographic confirmation receipts represent verified bridge settlements.
- Bridges lock token assets internally and issue a Unique Transaction Reference (UTR).
- The presentation of a valid receipt triggers matching asset liquidation or release.

## 24. API Design Guidelines
- Strictly JSON requests and responses over REST paths.
- Endpoints structured under `/api/v1/`.
- Handle HTTP statuses correctly:
  - `200 OK` / `201 Created` / `202 Accepted`
  - `400 Bad Request` / `401 Unauthorized` / `404 Not Found`
  - `500 Internal Server Error`

## 25. Database Guidelines
- Persistent storage must use **BadgerDB**.
- Perform lookups and writes inside transactions:
  - Read-only queries using `View`.
  - Writes using `Update`.
- Wrap database entries using standard Go structures serialized as JSON or binary byte-arrays.

## 26. Git Workflow
- Work must occur on feature branches branched off `main`.
- Clean branch histories: squash commits when merging to keep history clean.

## 27. Branch Naming
- `feature/name-of-feature`
- `bugfix/issue-description`
- `refactor/component-name`

## 28. Conventional Commits
Format commit messages with structural semantics:
- `feat: add Ed25519 signature checks`
- `fix: correct transaction nonce increment`
- `docs: update setup guidelines`
- `test: add routing simulation tests`

## 29. Pull Request Checklist
- [ ] Code compiles on Go 1.26.
- [ ] Code formatted with `gofmt`.
- [ ] Unit tests pass with `-race`.
- [ ] No circular dependencies introduced.
- [ ] GoDoc comments present on all public methods and structures.

## 30. Code Review Checklist
- Check for global state leakages.
- Ensure proper lock patterns are used for concurrent thread interactions.
- Confirm database transactions are closed cleanly.
- Verify error messages wrap standard return codes.

## 31. Definition of Done
- Complete implementation of the feature logic.
- Associated unit and integration tests passing.
- Documented changes inside the README or API manuals.
- Code reviewed and approved by another team contributor.

## 32. Documentation Requirements
- Keep architectural graphs updated.
- Every directory must contain `doc.go` clarifying the package purpose.
- Maintain clear README setup steps for local testing.

## 33. AI Collaboration Guidelines
- Develop incrementally. Prefer small PRs representing single packages.
- Always write interface specifications first before generating logic.
- Keep structural implementations clear and understandable for undergraduate students.

## 34. Project Roadmap
- **Week 1:** Core Digital Asset, Wallet generation, local ledger, transaction signatures, REST API & CLI.
- **Week 2:** Distributed Network, discovery, synchronizing, consensus voting.
- **Week 3:** PCN routing and channels settlement models.
- **Week 4:** Settlement bridge interface, web demo, final docs.

## 35. Future Enhancements
- Blind signatures for privacy protection.
- Support for complex routing fee optimizations.
- Implementation of hardware-secured bluetooth mesh credentials.

---

### Command Guide

**Build and compile binaries:**
```bash
make build
```

**Run local unit tests:**
```bash
make test
```

**Clean local compilation artifacts:**
```bash
make clean
```

**Run the node application:**
```bash
make run
```
