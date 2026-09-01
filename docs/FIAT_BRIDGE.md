# Non-Custodial Fiat Bridge & UTR Cryptographic Receipt Specification

## 1. Overview

The LiquidLink Fiat Bridge enables non-custodial bridging between on-chain crypto assets (LQD) and real-world fiat gateway payments (INR via UPI / IMPS). The protocol locks crypto assets in smart escrow, generates a Unique Transaction Reference (UTR) mapping, and releases or liquidates collateral upon cryptographic verification of bank settlement receipts.

---

## 2. Data Schema

### 2.1 UTR Mapping Record
```go
type UTRMapping struct {
	UTR           string    `json:"utr"`            // e.g. "UTRa1b2c3d4e5f6"
	TransactionID string    `json:"transaction_id"` // On-chain transaction reference
	Status        string    `json:"status"`         // "pending", "linked", "settled"
	CreatedAt     time.Time `json:"created_at"`
}
```

### 2.2 Bank Settlement Receipt
```go
type Receipt struct {
	ID        string         `json:"id"`
	UTR       string         `json:"utr"`
	Address   crypto.Address `json:"address"`
	Amount    token.Amount   `json:"amount"`
	Timestamp time.Time      `json:"timestamp"`
	Signature []byte         `json:"signature"` // Authorized gateway operator signature
}
```

---

## 3. Fiat Bridge Lifecycle & Sequence

```
User Address                     Bridge / Node                     Fiat Bank Gateway
    │                                  │                                   │
    │  1. InitiateBridgeRequest(LQD)   │                                   │
    ├─────────────────────────────────►│                                   │
    │                                  │  2. Debit LQD & Lock Escrow       │
    │                                  │  3. MapUTR("UTR-BANK-...")        │
    │                                  │                                   │
    │  4. Return UTRMapping (Pending)  │                                   │
    │◄─────────────────────────────────┤                                   │
    │                                  │                                   │
    │                                  │  5. Initiate Fiat Payout (INR)    │
    │                                  │──────────────────────────────────►│
    │                                  │                                   │
    │                                  │  6. Return Operator Receipt       │
    │                                  │◄──────────────────────────────────┤
    │                                  │                                   │
    │  7. SettleBridgeRequest(Receipt) │                                   │
    ├─────────────────────────────────►│                                   │
    │                                  │  8. Verify Signature & Status     │
    │                                  │  9. Transition Status -> "settled"│
    │  10. Settlement Confirmed        │                                   │
    │◄─────────────────────────────────┤                                   │
```

---

## 4. Security & Cryptographic Verification

1. **Collateral Lock**: Calling `InitiateBridgeRequest` immediately debits the user's crypto balance and locks funds in escrow. If UTR creation fails, debits are automatically credited back.
2. **Deterministic UTR Generation**: UTR identifiers are generated using SHA-256 hashes of the target address and timestamp:
   $$\text{UTR} = \text{"UTR"} + \text{SHA256}(\text{"UTR-BANK-"}\, \parallel \text{Address} \parallel \text{Timestamp})[:8]$$
3. **Receipt Validation (`VerifyReceipt`)**:
   - Computes payload hash:
     $$H_{\text{receipt}} = \text{SHA256}(\text{ID} \parallel \text{UTR} \parallel \text{Address} \parallel \text{Amount} \parallel \text{Timestamp})$$
   - Verifies Ed25519 signature against gateway operator public key ($K_{\text{operator}}$).
   - Only valid operator signatures can transition UTR status from `pending` to `settled`.
