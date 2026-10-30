package pcn_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/n3055/liquid-link/internal/config"
	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/ledger"
	"github.com/n3055/liquid-link/internal/pcn"
	"github.com/n3055/liquid-link/internal/storage"
	"github.com/n3055/liquid-link/internal/token"
)

func setupTestPCNManager(t *testing.T) (pcn.Manager, *crypto.KeyPair, func()) {
	tmpDir, err := os.MkdirTemp("", "pcn-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	cfg := &config.StorageConfig{Path: tmpDir, ValueLogFileSize: 1024 * 1024}
	store, err := storage.NewBadgerStore(cfg)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create badger store: %v", err)
	}

	stateRepo := ledger.NewStateRepository(store)
	led := ledger.NewLedger(stateRepo)
	balanceAdapter := ledger.NewLedgerBalanceAdapter(led.(ledger.StateRepository))
	assetMgr := token.NewAssetManager(balanceAdapter, store)

	gen := crypto.NewGenerator()
	kp, _ := gen.GenerateKeyPair()
	signer, _ := crypto.NewSigner(kp.PrivateKey)
	nodeAddr := gen.DeriveAddress(kp.PublicKey)
	_ = assetMgr.Mint(context.Background(), nodeAddr, 5000)

	registry := pcn.NewHolderRegistry(store)
	pcnMgr := pcn.NewManager(nodeAddr, signer, led, registry, assetMgr)

	cleanup := func() {
		_ = store.Close()
		os.RemoveAll(tmpDir)
	}

	return pcnMgr, kp, cleanup
}

func TestPCN_ChannelLifecycleAndChallenge(t *testing.T) {
	ctx := context.Background()
	mgr, kpA, cleanup := setupTestPCNManager(t)
	defer cleanup()

	gen := crypto.NewGenerator()
	partyA := gen.DeriveAddress(kpA.PublicKey)
	if partyA == "" {
		t.Fatalf("expected non-empty partyA address")
	}

	partyB := crypto.Address("0xnode_party_b")

	// 1. Open Channel
	ch, err := mgr.OpenChannel(ctx, partyB, 1000)
	if err != nil {
		t.Fatalf("failed to open channel: %v", err)
	}

	if ch.Share.BalancePartyA != 1000 || ch.Share.BalancePartyB != 0 {
		t.Errorf("unexpected initial balance split: A=%d, B=%d", ch.Share.BalancePartyA, ch.Share.BalancePartyB)
	}

	// 2. Off-chain State Updates
	chUpdated, err := mgr.UpdateChannelState(ctx, ch.ID, 300, true)
	if err != nil {
		t.Fatalf("failed to update channel state: %v", err)
	}
	if chUpdated.Sequence != 2 {
		t.Errorf("expected sequence 2, got %d", chUpdated.Sequence)
	}

	// 3. Initiate Settlement (Dispute Timelock Starts)
	if err := mgr.InitiateSettlement(ctx, ch.ID); err != nil {
		t.Fatalf("failed to initiate settlement: %v", err)
	}

	// 4. Challenge with Higher Sequence State (S=3)
	higherState := *chUpdated
	higherState.Sequence = 3
	higherState.Share.BalancePartyA = 600
	higherState.Share.BalancePartyB = 400

	if err := mgr.ChallengeSettlement(ctx, ch.ID, &higherState); err != nil {
		t.Fatalf("failed to challenge settlement: %v", err)
	}

	// 5. Finalize Settlement
	if err := mgr.FinalizeSettlement(ctx, ch.ID); err != nil {
		t.Fatalf("failed to finalize settlement: %v", err)
	}
}

func TestPCN_ApplyOfflineStateStatement(t *testing.T) {
	ctx := context.Background()
	mgr, kpA, cleanup := setupTestPCNManager(t)
	defer cleanup()

	gen := crypto.NewGenerator()
	partyA := gen.DeriveAddress(kpA.PublicKey)
	signerA, _ := crypto.NewSigner(kpA.PrivateKey)

	kpB, _ := gen.GenerateKeyPair()
	signerB, _ := crypto.NewSigner(kpB.PrivateKey)
	partyB := gen.DeriveAddress(kpB.PublicKey)

	// 1. Open Channel (1000 collateral, Party A = 1000, Party B = 0)
	ch, err := mgr.OpenChannel(ctx, partyB, 1000)
	if err != nil {
		t.Fatalf("failed to open channel: %v", err)
	}

	// 2. Reject unauthorized payee signature (Party B tries to deduct Party A's balance)
	forgedHash := sha256.Sum256([]byte(fmt.Sprintf("%s|%d|%d|%d", ch.ID, 2, 700, 300)))
	forgedSigB, _ := signerB.Sign(forgedHash[:])
	forgedStmt := &pcn.OfflineStateStatement{
		Type:      pcn.TypePCNOffline,
		ChannelID: ch.ID,
		Sequence:  2,
		PartyA:    partyA,
		PartyB:    partyB,
		BalanceA:  700,
		BalanceB:  300,
		Sender:    partyB,
		Signature: forgedSigB,
		PublicKey: []byte(kpB.PublicKey),
	}
	if _, err := mgr.ApplyOfflineStateStatement(ctx, forgedStmt); err == nil {
		t.Fatalf("expected error when payee attempts to authorize payer's balance reduction, got nil")
	}

	// 3. Valid statement signed by payer Party A (Sequence 2, 700 / 300)
	hashA := sha256.Sum256([]byte(fmt.Sprintf("%s|%d|%d|%d", ch.ID, 2, 700, 300)))
	sigA, err := signerA.Sign(hashA[:])
	if err != nil {
		t.Fatalf("failed to sign offline state: %v", err)
	}

	validStmt := &pcn.OfflineStateStatement{
		Type:      pcn.TypePCNOffline,
		ChannelID: ch.ID,
		Sequence:  2,
		PartyA:    partyA,
		PartyB:    partyB,
		BalanceA:  700,
		BalanceB:  300,
		Sender:    partyA,
		Signature: sigA,
		PublicKey: []byte(kpA.PublicKey),
	}

	updatedCh, err := mgr.ApplyOfflineStateStatement(ctx, validStmt)
	if err != nil {
		t.Fatalf("failed to apply valid offline statement: %v", err)
	}
	if updatedCh.Sequence != 2 {
		t.Errorf("expected sequence 2, got %d", updatedCh.Sequence)
	}
	if updatedCh.Share.BalancePartyA != 700 || updatedCh.Share.BalancePartyB != 300 {
		t.Errorf("expected balances 700/300, got %d/%d", updatedCh.Share.BalancePartyA, updatedCh.Share.BalancePartyB)
	}

	// 4. Valid subsequent statement where Party B pays Party A (Sequence 3, 800 / 200, signed by Party B)
	hashB := sha256.Sum256([]byte(fmt.Sprintf("%s|%d|%d|%d", ch.ID, 3, 800, 200)))
	sigB, _ := signerB.Sign(hashB[:])
	validStmtB := &pcn.OfflineStateStatement{
		Type:      pcn.TypePCNOffline,
		ChannelID: ch.ID,
		Sequence:  3,
		PartyA:    partyA,
		PartyB:    partyB,
		BalanceA:  800,
		BalanceB:  200,
		Sender:    partyB,
		Signature: sigB,
		PublicKey: []byte(kpB.PublicKey),
	}
	updatedCh2, err := mgr.ApplyOfflineStateStatement(ctx, validStmtB)
	if err != nil {
		t.Fatalf("failed to apply valid statement signed by Party B: %v", err)
	}
	if updatedCh2.Sequence != 3 || updatedCh2.Share.BalancePartyA != 800 || updatedCh2.Share.BalancePartyB != 200 {
		t.Errorf("unexpected balances after Party B payment: %d/%d (seq %d)", updatedCh2.Share.BalancePartyA, updatedCh2.Share.BalancePartyB, updatedCh2.Sequence)
	}

	// 5. Reject out-of-order sequence (Sequence 3 <= Current 3)
	staleStmt := &pcn.OfflineStateStatement{
		ChannelID: ch.ID,
		Sequence:  3,
		BalanceA:  800,
		BalanceB:  200,
		Signature: sigB,
		PublicKey: []byte(kpB.PublicKey),
	}
	if _, err := mgr.ApplyOfflineStateStatement(ctx, staleStmt); err == nil {
		t.Errorf("expected error applying stale sequence, got nil")
	}

	// 6. Reject balance invariant violation (Total 1100 != 1000)
	invalidBalStmt := &pcn.OfflineStateStatement{
		ChannelID: ch.ID,
		Sequence:  4,
		BalanceA:  800,
		BalanceB:  300,
		Signature: sigB,
		PublicKey: []byte(kpB.PublicKey),
	}
	if _, err := mgr.ApplyOfflineStateStatement(ctx, invalidBalStmt); err == nil {
		t.Errorf("expected error for balance conservation violation, got nil")
	}

	// 7. Reject integer overflow in balance calculation (MaxUint64 + 1001 wraps to 1000)
	overflowStmt := &pcn.OfflineStateStatement{
		ChannelID: ch.ID,
		Sequence:  4,
		BalanceA:  math.MaxUint64,
		BalanceB:  1001,
		Signature: sigB,
		PublicKey: []byte(kpB.PublicKey),
	}
	if _, err := mgr.ApplyOfflineStateStatement(ctx, overflowStmt); err == nil {
		t.Errorf("expected error for balance overflow attempt, got nil")
	}

	// 8. Reject missing signature / public key
	missingSigStmt := &pcn.OfflineStateStatement{
		ChannelID: ch.ID,
		Sequence:  4,
		BalanceA:  600,
		BalanceB:  400,
	}
	if _, err := mgr.ApplyOfflineStateStatement(ctx, missingSigStmt); err == nil {
		t.Errorf("expected error for missing signature, got nil")
	}

	// 9. Reject unauthorized third-party signer (keypair C not in channel)
	kpC, _ := gen.GenerateKeyPair()
	signerC, _ := crypto.NewSigner(kpC.PrivateKey)
	hashC := sha256.Sum256([]byte(fmt.Sprintf("%s|%d|%d|%d", ch.ID, 4, 600, 400)))
	sigC, _ := signerC.Sign(hashC[:])
	unauthStmt := &pcn.OfflineStateStatement{
		ChannelID: ch.ID,
		Sequence:  4,
		BalanceA:  600,
		BalanceB:  400,
		Signature: sigC,
		PublicKey: []byte(kpC.PublicKey),
	}
	if _, err := mgr.ApplyOfflineStateStatement(ctx, unauthStmt); err == nil {
		t.Errorf("expected error for unauthorized third-party signer, got nil")
	}

	// 10. Reject invalid cryptographic signature
	badSigStmt := &pcn.OfflineStateStatement{
		ChannelID: ch.ID,
		Sequence:  4,
		BalanceA:  850,
		BalanceB:  150,
		Signature: []byte("invalid-fake-signature-bytes-64-byte-placeholder-for-test"),
		PublicKey: []byte(kpB.PublicKey),
	}
	if _, err := mgr.ApplyOfflineStateStatement(ctx, badSigStmt); err == nil {
		t.Errorf("expected error for invalid signature, got nil")
	}
}
