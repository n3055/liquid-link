package pcn_test

import (
	"context"
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
