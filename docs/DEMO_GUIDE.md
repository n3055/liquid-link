# LiquidLink Step-by-Step UI & Feature Demonstration Guide

This guide provides a step-by-step walkthrough for demonstrating all features of the LiquidLink platform using the interactive Web Dashboard and REST API.

---

## 🚀 Step 1: Launch Multi-Validator Network

1. Open a terminal and run Validator Node A:
   ```bash
   make run-a
   ```
2. Open a second terminal and run Validator Node B:
   ```bash
   make run-b
   ```
3. Open your browser and navigate to:
   - **Node A Dashboard**: [http://localhost:8080/](http://localhost:8080/)
   - **Node B Dashboard**: [http://localhost:8081/](http://localhost:8081/)

---

## 🎨 Step 2: Client-Side Web3 Wallet Creation

1. On **Node A Dashboard** ([http://localhost:8080/](http://localhost:8080/)), navigate to the top-right header control.
2. Click **Generate Wallet** (`POST /api/v1/wallet/create`).
3. Notice:
   - A unique Ed25519 public/private keypair is generated.
   - A 100% unique, pixel-art cartoon identicon avatar is rendered for the wallet.
   - An initial genesis balance of **1,000 LQD** is credited to the account.

---

## 💸 Step 3: Direct On-Chain Crypto Transfer

1. Copy the address of Node B from [http://localhost:8081/api/v1/status](http://localhost:8081/api/v1/status).
2. On Node A Dashboard, click **Quick Actions $\to$ Send**.
3. Enter Receiver Address (`0x...`) and Amount (`250 LQD`).
4. Click **Submit Transfer**.
5. Observe:
   - Transaction status returns `202 Accepted`.
   - Node A balance updates immediately.
   - Node B receives asynchronous credit sync (`POST /api/v1/wallet/credit-sync`) without inflating total circulating supply.

---

## ⚡ Step 4: Payment Channel Network (PCN) State Channels

1. Click **Channels $\to$ Open Channel**.
2. Enter Counterparty Address and Deposit Amount (`500 LQD`).
3. Click **Open Channel**.
4. Observe channel state transitions to `StateOpen` with Sequence $S=1$.
5. Execute off-chain sequence balance updates by clicking **Update State**. Observe sequence number increment to $S=2, S=3$ with bilateral co-signatures.
6. Initiate dispute settlement by clicking **Initiate Settle**. Notice channel enters `StateClosing` with a 24-hour challenge timelock window.

---

## 📈 Step 5: LQD/INR Limit Orderbook & Market Peg Stability

1. Click **Orderbook (INR)** tab.
2. View real-time orderbook depth (highest bids and lowest asks) and clearing price.
3. Submit a new Limit Buy or Sell order (`Price: 1.00 INR`, `Amount: 100 LQD`).
4. Switch to **Staking & Peg Stability** tab:
   - **Expansion ($P > \text{₹}1.00$)**: Stake LQD tokens in seigniorage pool to earn yield APY dividends.
   - **Contraction ($P < \text{₹}1.00$)**: Buy discounted LQD-Bonds (1.15x face value upon expansion redemption).

---

## 🏦 Step 6: Non-Custodial Fiat Bridge & UTR Payouts

1. Click **Fiat Bridge $\to$ Request**.
2. Enter Receiver Address and Amount (`150 LQD`). Click **Submit Bridge Request**.
3. Observe:
   - Crypto collateral is locked in escrow (`Debit`).
   - A unique bank reference number (`UTR-...`) is generated with status `pending`.
4. Submit operator receipt verification (`POST /api/v1/bridge/verify-receipt`).
5. Observe status transitions to `settled`, finalizing the fiat bridging transaction.
