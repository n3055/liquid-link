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

func setupTestPCNManager(t *testing.T) (pcn.Manager, crypto.Address, func()) {
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

	nodeAddr := crypto.Address("0xnode_party_a")
	_ = assetMgr.Mint(context.Background(), nodeAddr, 5000)

	gen := crypto.NewGenerator()
	kp, _ := gen.GenerateKeyPair()
	signer, _ := crypto.NewSigner(kp.PrivateKey)

	registry := pcn.NewHolderRegistry(store)
	pcnMgr := pcn.NewManager(nodeAddr, signer, led, registry, assetMgr)

	cleanup := func() {
		_ = store.Close()
		os.RemoveAll(tmpDir)
	}

	return pcnMgr, nodeAddr, cleanup
}

func TestPCN_ChannelLifecycleAndChallenge(t *testing.T) {
	ctx := context.Background()
	mgr, partyA, cleanup := setupTestPCNManager(t)
	defer cleanup()

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
	mgr, partyA, cleanup := setupTestPCNManager(t)
	defer cleanup()

	gen := crypto.NewGenerator()
	kpB, _ := gen.GenerateKeyPair()
	signerB, _ := crypto.NewSigner(kpB.PrivateKey)
	partyB := gen.DeriveAddress(kpB.PublicKey)

	// 1. Open Channel (1000 collateral)
	ch, err := mgr.OpenChannel(ctx, partyB, 1000)
	if err != nil {
		t.Fatalf("failed to open channel: %v", err)
	}

	// Create valid offline state statement signed by party B (Sequence 2, 700 / 300)
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s|%d|%d|%d", ch.ID, 2, 700, 300)))
	sigB, err := signerB.Sign(hash[:])
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
		Sender:    partyB,
		Signature: sigB,
		PublicKey: []byte(kpB.PublicKey),
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

	// 2. Reject out-of-order sequence (Sequence 2 <= Current 2)
	staleStmt := &pcn.OfflineStateStatement{
		ChannelID: ch.ID,
		Sequence:  2,
		BalanceA:  700,
		BalanceB:  300,
		Signature: sigB,
		PublicKey: []byte(kpB.PublicKey),
	}
	if _, err := mgr.ApplyOfflineStateStatement(ctx, staleStmt); err == nil {
		t.Errorf("expected error applying stale sequence, got nil")
	}

	// 3. Reject balance invariant violation (Total 1100 != 1000)
	invalidBalStmt := &pcn.OfflineStateStatement{
		ChannelID: ch.ID,
		Sequence:  3,
		BalanceA:  800,
		BalanceB:  300,
		Signature: sigB,
		PublicKey: []byte(kpB.PublicKey),
	}
	if _, err := mgr.ApplyOfflineStateStatement(ctx, invalidBalStmt); err == nil {
		t.Errorf("expected error for balance conservation violation, got nil")
	}

	// 4. Reject integer overflow in balance calculation (MaxUint64 + 1001 wraps to 1000)
	overflowStmt := &pcn.OfflineStateStatement{
		ChannelID: ch.ID,
		Sequence:  3,
		BalanceA:  math.MaxUint64,
		BalanceB:  1001,
		Signature: sigB,
		PublicKey: []byte(kpB.PublicKey),
	}
	if _, err := mgr.ApplyOfflineStateStatement(ctx, overflowStmt); err == nil {
		t.Errorf("expected error for balance overflow attempt, got nil")
	}

	// 5. Reject missing signature / public key
	missingSigStmt := &pcn.OfflineStateStatement{
		ChannelID: ch.ID,
		Sequence:  3,
		BalanceA:  600,
		BalanceB:  400,
	}
	if _, err := mgr.ApplyOfflineStateStatement(ctx, missingSigStmt); err == nil {
		t.Errorf("expected error for missing signature, got nil")
	}

	// 6. Reject unauthorized third-party signer (keypair C not in channel)
	kpC, _ := gen.GenerateKeyPair()
	signerC, _ := crypto.NewSigner(kpC.PrivateKey)
	hashC := sha256.Sum256([]byte(fmt.Sprintf("%s|%d|%d|%d", ch.ID, 3, 600, 400)))
	sigC, _ := signerC.Sign(hashC[:])
	unauthStmt := &pcn.OfflineStateStatement{
		ChannelID: ch.ID,
		Sequence:  3,
		BalanceA:  600,
		BalanceB:  400,
		Signature: sigC,
		PublicKey: []byte(kpC.PublicKey),
	}
	if _, err := mgr.ApplyOfflineStateStatement(ctx, unauthStmt); err == nil {
		t.Errorf("expected error for unauthorized third-party signer, got nil")
	}

	// 7. Reject invalid cryptographic signature
	badSigStmt := &pcn.OfflineStateStatement{
		ChannelID: ch.ID,
		Sequence:  3,
		BalanceA:  600,
		BalanceB:  400,
		Signature: []byte("invalid-fake-signature-bytes-64-byte-placeholder-for-test"),
		PublicKey: []byte(kpB.PublicKey),
	}
	if _, err := mgr.ApplyOfflineStateStatement(ctx, badSigStmt); err == nil {
		t.Errorf("expected error for invalid signature, got nil")
	}
}
