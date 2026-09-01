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

## 2. AI Codebase Navigation & Readability Rules

To ensure large codebases remain easy for AI agents and human collaborators to index, grep, and modify:

### 2.1 Function & Type Documentation
- **Mandatory Function Comments:** Every exported and package-level unexported function must have a clear GoDoc comment summarizing its purpose, input parameters, and return values.
- **Searchable Keywords:** Include specific architectural keywords in comment headers (e.g. `[Ledger]`, `[PCN State]`, `[Receipt Verification]`) so regex and symbol search tools can locate handlers instantly.

### 2.2 Modular Architecture & Single Responsibility
- **Loose Coupling:** Favor interfaces over concrete implementations in logic layers.
- **Constructor Injection:** Inject dependencies via `New<Service>(...)` constructors instead of using global state or singletons.
- **File Length & Scope:** Keep Go source files short (ideally under 300 lines). Break large packages into focused files (e.g. `dashboard.go`, `pcn_impl.go`).

### 2.3 Meaningful & Unambiguous Naming
- **Explicit Variable Names:** Avoid single-letter variables except for trivial loop counters or short contexts (`ctx`, `err`, `tx`). Use clear names like `channelID`, `targetBalance`, `receiptRegistry`.
- **Domain Alignment:** Align variable and function names strictly with the domain model (`UTRMapping`, `StateChannel`, `AssetManager`).

### 2.4 Comprehensive Test Coverage
- **Co-located Tests:** Write test files alongside source files (`*_test.go`).
- **Descriptive Test Names:** Use descriptive test function names detailing the scenario being tested (e.g., `TestWalletTransfer_InsufficientBalance`, `TestStateChannel_InitiateSettlement`).
- **Race Detection:** Always verify changes with `make test` or `go test -v -race ./...`.

---

## 3. Workflow Checklist for AI Agents

1. **Understand & Inspect:** Read relevant package files before editing. Run tests to establish baseline.
2. **Implement:** Follow SOLID principles, inject dependencies, write clear function comments.
3. **Test:** Add unit tests covering happy paths and edge cases. Ensure `make test` passes.
4. **Format & Lint:** Run `gofmt -s -w .` and `go vet ./...`.
5. **PR & Account Hygiene:** Use `gh auth switch -u n3055` to submit PRs, then `gh auth switch -u rishi-anand_dex` to reset account state.
