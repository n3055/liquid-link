# PLAN: Fix INR Reserve Pool, Bond/Share UI Display & Protocol Monetary Context Wiring

## Context
The LiquidLink web console and backend contain several UI display and state wiring gaps regarding the fiat INR reserve pool, seigniorage bond/share tracking, and protocol monetary supply metrics:
1. **Missing INR Reserve Pool:** No total fiat bank reserve pool card/metric is configured or displayed in the UI or exposed via a dedicated `/api/v1/fiat/pool` endpoint.
2. **Bond/Share Zero Display Bug:** In `internal/stability/stability_impl.go`, when `store == nil` (default node execution mode), `loadStakingPosition`, `saveStakingPosition`, `loadBondPosition`, and `saveBondPosition` fail to retain user positions in memory, causing staked shares and bonds held to always return `0` after staking or bond purchases.
3. **Missing Protocol Total Bonds & Shares Context:** The UI and `/api/v1/stability/status` integration only render personal holdings and do not display total protocol staked shares and total outstanding bonds context.
4. **Missing Total Supply Converted to Bonds & Shares:** The UI lacks a metric card reflecting circulating supply converted into bonds and shares collateral.

## Approach
1. **Fix In-Memory Stability Persistence (`internal/stability/stability_impl.go`):**
   - Add thread-safe `stakingPositions map[crypto.Address]*StakingPosition` and `bondPositions map[crypto.Address]*BondPosition` fields to `engineImpl`.
   - Update `loadStakingPosition`, `saveStakingPosition`, `loadBondPosition`, and `saveBondPosition` to fallback to in-memory maps when `e.store == nil` (and maintain them concurrently when `e.store` is used).
2. **Expose Total INR Pool Reserve in Fiat Bank & API (`internal/fiat/`, `internal/api/`):**
   - Add `/api/v1/fiat/pool` REST endpoint in `internal/api/api.go` returning `total_inr_pool`, `market_maker_reserve`, and active user balances.
   - Include `TotalINRPool float64` in `dashboardData` (`internal/api/dashboard.go`) populated via `fiatM.GetTotalPoolReserve(ctx)`.
3. **UI Integration for INR Pool, Protocol Bonds/Shares & Converted Supply (`internal/api/dashboard_template.go`):**
   - Add an **INR Reserve Pool** metric card to the top stat-strip and Orderbook header.
   - Add **Protocol Total Staked Shares**, **Protocol Total Outstanding Bonds**, and **Total Supply Converted to Bonds & Shares** metric cards in `segStability`.
   - Update client-side JS (`refreshDashboardData`, `refreshStabilityUI`, `loadActiveWalletUI`) to dynamically update elements: `#statInrPool`, `#totalStakedSharesVal`, `#totalOutstandingBondsVal`, `#totalSupplyConvertedVal`, `#userStakedShares`, and `#userBondsHeld`.
4. **Write Co-located Unit Tests:**
   - Add tests in `stability_test.go`, `fiat_test.go`, and `dashboard_test.go` covering in-memory position persistence, fiat pool aggregation, and dashboard metric rendering.

## Files to modify
- `internal/stability/stability_impl.go`: Add in-memory position fallback maps for staking & bonds.
- `internal/stability/stability_test.go`: Add test verifying in-memory position persistence without Badger store.
- `internal/api/api.go`: Add `/api/v1/fiat/pool` handler and route.
- `internal/api/dashboard.go`: Add `TotalINRPool` to `dashboardData` struct and collect it during data assembly.
- `internal/api/dashboard_template.go`: Add INR Reserve Pool stat card, total bonds/shares cards, supply converted card, and update JS polling.
- `internal/api/dashboard_test.go`: Verify element IDs (`statInrPool`, `totalStakedSharesVal`, `totalOutstandingBondsVal`, `totalSupplyConvertedVal`) in dashboard HTML render test.

## Reuse
- `fiat.Manager.GetTotalPoolReserve(ctx context.Context)` in `internal/fiat/fiat_impl.go` for calculating total INR reserve pool.
- `stability.Engine.GetStatus(ctx context.Context)` in `internal/stability/stability_impl.go` for total staked shares and outstanding bonds.
- `token.AssetManager.TotalSupply(ctx context.Context)` in `internal/token/` for supply metrics.
- Existing chi routing & `respondJSON` helpers in `internal/api/api.go`.

## Steps
- [ ] **Step 1: Fix Stability In-Memory Position Persistence**
  - Add `stakingPositions` and `bondPositions` maps to `engineImpl` in `internal/stability/stability_impl.go`.
  - Initialize maps in `NewEngine`.
  - Update `loadStakingPosition`, `saveStakingPosition`, `loadBondPosition`, and `saveBondPosition` to read/write memory maps when `store == nil`.
- [ ] **Step 2: Expose Fiat Reserve Pool in API & Dashboard Model**
  - Add `TotalINRPool float64` to `dashboardData` in `internal/api/dashboard.go`.
  - Fetch `TotalINRPool` via `fiatM.GetTotalPoolReserve(ctx)` in `collectDashboardData`.
  - Add `handleFiatPool` endpoint at `GET /api/v1/fiat/pool` in `internal/api/api.go`.
- [ ] **Step 3: Update Dashboard Web UI Template & Client-Side JS**
  - Add INR Reserve Pool stat card in stat-strip and Orderbook depth header in `internal/api/dashboard_template.go`.
  - Add Protocol Total Shares, Total Bonds, and Supply Converted cards in `segStability` section of `internal/api/dashboard_template.go`.
  - Update client-side JS (`refreshDashboardData`, `refreshStabilityUI`, `loadActiveWalletUI`) to poll `/api/v1/fiat/pool` and update DOM IDs.
- [ ] **Step 4: Verify and Test End-to-End**
  - Add unit tests in `internal/stability/stability_test.go` and `internal/api/dashboard_test.go`.
  - Run `go test -v -race ./...` to ensure clean execution and race safety.
  - Verify formatting with `gofmt -s -w .` and linting with `go vet ./...`.

## Verification
- Run `go test -v -race ./...` across all packages.
- Verify `gofmt -s -w .` and `go vet ./...` pass with zero warnings.
