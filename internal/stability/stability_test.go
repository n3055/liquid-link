package stability_test

import (
	"context"
	"os"
	"testing"

	"github.com/n3055/liquid-link/internal/config"
	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/ledger"
	"github.com/n3055/liquid-link/internal/stability"
	"github.com/n3055/liquid-link/internal/storage"
	"github.com/n3055/liquid-link/internal/token"
)

func setupTestEngine(t *testing.T) (stability.Engine, func()) {
	tmpDir, err := os.MkdirTemp("", "stability-test-*")
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

	eng := stability.NewEngine(store, assetMgr)

	cleanup := func() {
		_ = store.Close()
		os.RemoveAll(tmpDir)
	}

	return eng, cleanup
}

func TestStability_GetStatus(t *testing.T) {
	ctx := context.Background()
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	status, err := eng.GetStatus(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.TargetPeg != 1.00 {
		t.Errorf("expected target peg 1.00, got %f", status.TargetPeg)
	}
	if status.CurrentAPY <= 0 {
		t.Errorf("expected positive APY, got %f", status.CurrentAPY)
	}
}

func TestStability_PriceHistory(t *testing.T) {
	ctx := context.Background()
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	history, err := eng.GetPriceHistory(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(history) == 0 {
		t.Errorf("expected non-empty price history")
	}
}

func TestStability_StakingFlow(t *testing.T) {
	ctx := context.Background()
	tmpDir, err := os.MkdirTemp("", "stability-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	cfg := &config.StorageConfig{Path: tmpDir, ValueLogFileSize: 1024 * 1024}
	store, err := storage.NewBadgerStore(cfg)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create badger store: %v", err)
	}
	defer func() {
		_ = store.Close()
		os.RemoveAll(tmpDir)
	}()

	stateRepo := ledger.NewStateRepository(store)
	led := ledger.NewLedger(stateRepo)
	balanceAdapter := ledger.NewLedgerBalanceAdapter(led.(ledger.StateRepository))
	assetMgr := token.NewAssetManager(balanceAdapter, store)

	user := crypto.Address("0xtestuser123")
	_ = assetMgr.Mint(ctx, user, 1000)

	eng := stability.NewEngine(store, assetMgr)

	pos, err := eng.StakeShares(ctx, user, 500)
	if err != nil {
		t.Fatalf("failed to stake: %v", err)
	}

	if pos.StakedShares != 500 {
		t.Errorf("expected 500 staked shares, got %d", pos.StakedShares)
	}

	fetched, err := eng.GetStakingPosition(ctx, user)
	if err != nil {
		t.Fatalf("failed to fetch staking position: %v", err)
	}
	if fetched.StakedShares != 500 {
		t.Errorf("expected 500 fetched staked shares, got %d", fetched.StakedShares)
	}
}
