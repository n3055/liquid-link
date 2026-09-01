# LiquidLink Validator Node Operational Guide

## 1. Overview

LiquidLink runs a **Permissioned Proof-of-Authority / Leader-Rotation PBFT Consensus Network**. Payment Service Provider (PSP) validator nodes maintain canonical block sequences, validate transaction nonces and cryptographic signatures, vote on proposed blocks, and execute state channel dispute settlements.

---

## 2. Multi-Validator Architecture

```
                       ┌─────────────────────────┐
                       │   Proposer (Leader)     │
                       │   Node A (Port 8080)    │
                       └───────────┬─────────────┘
                                   │
                    1. Propose Block (SHA-256)
                                   │
               ┌───────────────────┴───────────────────┐
               │                                       │
               ▼                                       ▼
 ┌───────────────────────────┐           ┌───────────────────────────┐
 │   Validator (Peer Node)   │           │   Validator (Peer Node)   │
 │    Node B (Port 8081)     │           │    Node C (Port 8082)     │
 └─────────────┬─────────────┘           └─────────────┬─────────────┘
               │                                       │
               └───────────────────┬───────────────────┘
                                   │
                      2. Submit PBFT Accept Votes
                                   │
                                   ▼
                       51%+ Majority Quorum Achieved
                       Block Appended to Canonical Chain
```

---

## 3. Configuration & Startup

Validator nodes are configured via YAML files in `configs/`:

```yaml
# configs/node_a.yaml
node:
  id: "node-a"
  port: 8080
  log_level: "info"

storage:
  path: "./data/node-a/badger"

consensus:
  leader_rotation_interval: "30s"
  vote_timeout: "5s"
  majority_threshold: 0.51

transports:
  internet:
    enabled: true
    bind_address: "0.0.0.0:9090"
    bootstrap_peers:
      - "127.0.0.1:9091"
```

### Launching Validators

```bash
# Terminal 1: Launch Validator Node A
make run-a

# Terminal 2: Launch Validator Node B
make run-b
```

---

## 4. Consensus & Voting Protocol

1. **Leader Election**: A deterministic leader selector selects the block proposer based on round-robin block heights:
   $$\text{Leader}(H) = \text{Validators}[H \bmod N]$$
2. **Block Proposal**: The leader packages pending mempool transactions into a `Block` struct, computes the SHA-256 header hash, and signs the proposal with Ed25519.
3. **Validation**: Peer validators verify block index, parent hash, transaction signatures, nonces, and account balances.
4. **Majority Quorum**: Once votes exceed the majority threshold ($>51\%$), the block is committed to BadgerDB, updating canonical account ledger balances across all nodes.
