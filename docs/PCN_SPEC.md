# Payment Channel Network (PCN) & Dispute Challenge Protocol Specification

## 1. Overview

The LiquidLink Payment Channel Network (PCN) provides zero-latency off-chain state channel transfers secured by bilateral co-signatures, collateral lockups, and an optimistic on-chain dispute challenge window.

---

## 2. State Channel Data Schema

A payment channel between Party A and Party B is defined by the following schema:

```go
type ChannelState string

const (
	StateOpen    ChannelState = "open"
	StateClosing ChannelState = "closing"
	StateClosed  ChannelState = "closed"
)

type Share struct {
	PartyA        crypto.Address `json:"party_a"`
	PartyB        crypto.Address `json:"party_b"`
	BalancePartyA token.Amount   `json:"balance_party_a"`
	BalancePartyB token.Amount   `json:"balance_party_b"`
}

type Bond struct {
	TxID             string       `json:"tx_id"`
	LockedCollateral token.Amount `json:"locked_collateral"`
	Expiry           time.Time    `json:"expiry"`
}

type Channel struct {
	ID        string       `json:"id"`
	State     ChannelState `json:"state"`
	Share     Share        `json:"share"`
	Bond      Bond         `json:"bond"`
	Sequence  uint64       `json:"sequence"`
	Signature []byte       `json:"signature"`
}
```

---

## 3. Channel Lifecycle & State Transitions

```
    ┌──────────┐
    │  Closed  │
    └────┬─────┘
         │ OpenChannel (Lock collateral)
         ▼
    ┌──────────┐      UpdateChannelState (S → S+1)
    │   Open   │ ◄──────────────────────────────────┐
    └────┬─────┘ ───────────────────────────────────┘
         │ InitiateSettlement (Start 24h dispute window)
         ▼
    ┌──────────┐
    │ Closing  │
    └────┬─────┘
         │ FinalizeSettlement (Distribute balances)
         ▼
    ┌──────────┐
    │  Closed  │
    └──────────┘
```

### 3.1 Opening a Channel (`OpenChannel`)
1. Initiator specifies counterparty address and initial deposit amount $D$.
2. Initiator's account balance is debited by $D$ via `AssetManager.Debit`.
3. Unique channel ID is derived deterministically:
   $$\text{ChannelID} = \text{SHA256}(\text{PartyA} \parallel \text{PartyB} \parallel D \parallel \text{Timestamp})[:16]$$
4. Initial state is set to `StateOpen` with Sequence $S=1$, $\text{BalancePartyA} = D$, $\text{BalancePartyB} = 0$.

### 3.2 Updating Off-Chain Balance State (`UpdateChannelState`)
1. Sender requests off-chain transfer of amount $a$.
2. Verifies channel is in `StateOpen` and sender has sufficient local share balance.
3. Adjusts balances:
   - If Party A sending to Party B: $\text{BalancePartyA}' = \text{BalancePartyA} - a$, $\text{BalancePartyB}' = \text{BalancePartyB} + a$.
   - If Party B sending to Party A: $\text{BalancePartyB}' = \text{BalancePartyB} - a$, $\text{BalancePartyA}' = \text{BalancePartyA} + a$.
4. Increments sequence number $S \to S+1$.
5. Recomputes state payload hash:
   $$H_{\text{state}} = \text{SHA256}(\text{ChannelID} \parallel S \parallel \text{BalancePartyA}' \parallel \text{BalancePartyB}')$$
6. Signs payload hash using sender Ed25519 private key.

### 3.3 Dispute Challenge Protocol (`InitiateSettlement` & `FinalizeSettlement`)
1. **Initiation**: Either participant calls `InitiateSettlement`. Channel state transitions from `StateOpen` to `StateClosing`, initiating a 24-hour challenge timelock $T_{\text{dispute}}$.
2. **Challenge Window ($T_{\text{dispute}} = 24\text{ hours}$)**: During the challenge window, either party may submit a higher sequence co-signed state $S' > S$.
3. **Punishment / Resolution**:
   - If a valid higher sequence state $S'$ is presented, the settlement updates to state $S'$.
   - If a party attempted to settle using an outdated state $S_{\text{old}} < S'$, the dishonest party's remaining share balance is slashed as a penalty.
4. **Finalization**: Upon expiration of $T_{\text{dispute}}$, `FinalizeSettlement` releases locked collateral by crediting $\text{BalancePartyA}$ to Party A and $\text{BalancePartyB}$ to Party B, transitioning state to `StateClosed`.
