# AGENTS.md — AI Agent Guidance & Codebase Conventions

This document provides guidelines and system context specifically tailored for AI coding agents working on the **LiquidLink** repository.

---

## 1. GitHub CLI (`gh`) Account Configuration

- **Primary Repository Owner:** `n3055`
- **Secondary Account:** `rishi-anand_dex` (Rishi Delta)

### Account Switching Protocol for PRs
When interacting with GitHub CLI (`gh`) for creating Pull Requests or pushing branches to `n3055/liquid-link`:

1. **Switch to `n3055` before GitHub PR operations:**
   ```bash
   gh auth switch -u n3055
   ```
2. **Perform `git push` and `gh pr create` actions.**
3. **Switch back to `rishi-anand_dex` when done:**
   ```bash
   gh auth switch -u rishi-anand_dex
   ```

---

## 2. Research Architecture & Modular Documentation Mapping

Agents modifying logic must consult the relevant specification under `docs/`:

- `docs/ARCHITECTURE.md` — Layered architecture, constructor injection rules, component wiring.
- `docs/PCN_SPEC.md` — State channels, sequence co-signatures, dispute timelock state machines.
- `docs/STABILITY.md` — Seigniorage expansion/contraction, LQD/INR orderbook matching logic.
- `docs/FIAT_BRIDGE.md` — Fiat bridge UTR mapping, receipt verification, escrow rules.
- `docs/TRANSPORT.md` — Universal transport envelopes, TCP/IP & BLE Mesh drivers.

---

## 3. AI Codebase Navigation & Readability Rules

### 3.1 Function & Type Documentation
- **Mandatory Function Comments:** Every exported and package-level unexported function must have a GoDoc comment summarizing purpose, inputs, and return values.
- **Searchable Comment Keywords:** Include specific architectural keywords in comment headers so regex and symbol tools can locate handlers instantly:
  - `[Ledger]` — Ledger state repository and balance state transitions.
  - `[PCN State]` — Payment Channel Network state management and settlement.
  - `[Orderbook]` — Limit orderbook placement, depth, and matching engine.
  - `[Stability]` — Algorithmic seigniorage peg, staking yields, and bond redemptions.
  - `[Receipt Verification]` — UTR mapping and cryptographic receipt verification.
  - `[Transport Envelope]` — Multi-medium transport messages and drivers.

### 3.2 Modular Architecture & Single Responsibility
- **Loose Coupling:** Rely on interfaces rather than concrete struct implementations in domain/logic packages.
- **Constructor Injection:** Inject dependencies via `New<Service>(...)` constructors. Avoid global state or singletons.
- **File Scope & Size:** Keep Go source files under 300 lines. Split large packages into focused files (e.g. `dashboard.go`, `pcn_impl.go`).

### 3.3 Meaningful & Unambiguous Naming
- **Explicit Variable Names:** Avoid single-letter variables except trivial loop counters or short contexts (`ctx`, `err`, `tx`). Use explicit names (`channelID`, `targetBalance`, `receiptRegistry`).
- **Domain Alignment:** Align variable names strictly with the domain model (`UTRMapping`, `StateChannel`, `AssetManager`).

### 3.4 Comprehensive Test Coverage
- **Co-located Tests:** Place test files adjacent to source code (`*_test.go`).
- **Descriptive Scenario Names:** Name tests clearly describing the scenario (e.g., `TestOrderbook_UnbackedOrderDebitFailure`, `TestStateChannel_InitiateSettlement`).
- **Race Detection Verification:** Always verify code changes with `go test -v -race ./...`.

---

## 4. Workflow Checklist for AI Agents

1. **Understand & Inspect:** Read relevant package files and modular docs in `docs/` before editing. Run `go test -v -race ./...` to verify baseline.
2. **Implement:** Follow SOLID principles, inject dependencies via constructors, add searchable comment tags.
3. **Test:** Add unit tests covering happy path and edge cases.
4. **Format & Lint:** Execute `gofmt -s -w .` and `go vet ./...`.
5. **PR & Account Hygiene:** Use `gh auth switch -u n3055` to submit PRs, then `gh auth switch -u rishi-anand_dex` to reset account state.
