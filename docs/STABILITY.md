# Dual-Token Algorithmic Seigniorage Peg Stability Specification

## 1. Overview

LiquidLink implements an algorithmic seigniorage stability mechanism designed to target a 1:1 peg between the **LQD stablecoin** and the **Indian Rupee (₹1.00 INR)**. The protocol integrates an on-chain limit orderbook matching engine with a dual-token stability model comprising **LQD-Shares (LQS)** and **LQD-Bonds (LQB)**.

---

## 2. Token Roles & Economic Dynamics

| Asset Class | Symbol | Target Value | Primary Utility |
| --- | --- | --- | --- |
| **Liquid Stablecoin** | `LQD` | ₹1.00 INR | Primary medium of exchange & PCN collateral |
| **Seigniorage Shares** | `LQS` | Variable | Governance & seigniorage dividend equity |
| **Seigniorage Bonds** | `LQB` | Discounted LQD | Debt instrument issued during supply contraction |

---

## 3. Peg State Machine & Rebalancing Mechanisms

```
                     ┌──────────────────────────────┐
                     │    Market Price (P_LQD)      │
                     └──────────────┬───────────────┘
                                    │
           ┌────────────────────────┼────────────────────────┐
           │                        │                        │
           ▼                        ▼                        ▼
  P_LQD > ₹1.005           ₹0.995 ≤ P_LQD ≤ ₹1.005       P_LQD < ₹0.995
┌────────────────────┐   ┌───────────────────────┐   ┌────────────────────┐
│  EXPANSION STATE   │   │     PEGGED STATE      │   │ CONTRACTION STATE  │
│ Mint LQD & Payout  │   │  Equilibrium Target   │   │ Sell Discounted    │
│  Staking Dividends │   │  No Protocol Action   │   │ Bonds (1.15x LQD)  │
└────────────────────┘   └───────────────────────┘   └────────────────────┘
```

### 3.1 Expansion Phase ($P_{\text{LQD}} > \text{₹}1.005$)
- **Trigger**: Demand exceeds supply, causing LQD market clearing price to rise above ₹1.00.
- **Action**: Seigniorage engine mints additional LQD token supply.
- **Distribution Order**:
  1. Redeem outstanding LQD-Bonds (paying 1.05x face value premium to bondholders).
  2. Distribute remaining minted LQD as APY yield dividends to users staking LQD-Shares in the protocol pool:
     $$\text{APY}_{\text{expansion}} = \text{BaseAPY} + (P_{\text{LQD}} - 1.00) \times 100\%$$

### 3.2 Contraction Phase ($P_{\text{LQD}} < \text{₹}0.995$)
- **Trigger**: Supply exceeds demand, causing LQD price to fall below ₹1.00.
- **Action**: Protocol issues discounted LQD-Bonds to contract circulating supply.
- **Bond Discount**: Users purchase LQD-Bonds by burning 1.0 LQD to acquire 1.15 LQD-Bonds.
- **Redemption**: LQD-Bonds become redeemable 1:1 for LQD + interest as soon as the market price recovers above ₹1.00.

---

## 4. Limit Orderbook Matching Engine

The orderbook matches limit buy (bids) and sell (asks) orders to establish clearing prices:

- **Order Schema**:
  ```go
  type Order struct {
      ID        string         `json:"id"`
      User      crypto.Address `json:"user"`
      Side      OrderSide      `json:"side"`  // SideBuy or SideSell
      Price     float64        `json:"price"` // Price per LQD in INR
      Amount    token.Amount   `json:"amount"`
      Filled    token.Amount   `json:"filled"`
      Status    OrderStatus    `json:"status"`
  }
  ```
- **Matching Priority**: Bids sorted price-descending (highest first); Asks sorted price-ascending (cheapest first); secondary sort by creation timestamp.
- **Settlement Safety**: `Debit` operation verifies seller account balance before executing credit. If seller balance is insufficient, the unbacked sell order is marked `StatusCancelled` and skipped to prevent unbacked token generation.
