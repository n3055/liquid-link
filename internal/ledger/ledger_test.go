package ledger

import (
	"context"
	"testing"

	"github.com/n3055/liquid-link/internal/config"
	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/storage"
)

func TestLedgerStateAndTransitions(t *testing.T) {
	// Setup storage
	tmpDir := t.TempDir()
	cfg := &config.StorageConfig{Path: tmpDir, ValueLogFileSize: 1024 * 1024}
	store, err := storage.NewBadgerStore(cfg)
	if err != nil {
		t.Fatalf("failed to open Badger store: %v", err)
	}
	defer store.Close()

	repo := NewStateRepository(store)
	l := NewLedger(repo)
	ctx := context.Background()

	alice := crypto.Address("0xalice")
	bob := crypto.Address("0xbob")

	// 1. Initial balance should be 0
	aliceState, err := l.GetAccountState(ctx, alice)
	if err != nil {
		t.Fatalf("failed to get alice state: %v", err)
	}
	if aliceState.Balance != 0 || aliceState.Nonce != 0 {
		t.Errorf("expected balance 0, nonce 0; got balance %d, nonce %d", aliceState.Balance, aliceState.Nonce)
	}

	// 2. Mint tokens to Alice using SystemAddress
	err = l.ApplyTransaction(ctx, SystemAddress, alice, 1000, 0)
	if err != nil {
		t.Fatalf("failed to apply mint transaction: %v", err)
	}

	// 3. Verify Alice state is updated in cache but not yet in repo
	aliceState, err = l.GetAccountState(ctx, alice)
	if err != nil {
		t.Fatalf("failed to get alice state: %v", err)
	}
	if aliceState.Balance != 1000 {
		t.Errorf("expected alice balance 1000, got %d", aliceState.Balance)
	}

	repoState, err := repo.GetAccountState(ctx, alice)
	if err == nil {
		t.Errorf("expected error getting Alice from repo before commit, got state: %v", repoState)
	}

	// 4. Perform a transaction: Alice -> Bob (500 tokens, nonce 1)
	err = l.ApplyTransaction(ctx, alice, bob, 400, 1)
	if err != nil {
		t.Fatalf("failed to apply transaction Alice -> Bob: %v", err)
	}

	// Check Alice and Bob updated states in ledger
	aliceState, _ = l.GetAccountState(ctx, alice)
	bobState, _ := l.GetAccountState(ctx, bob)

	if aliceState.Balance != 600 || aliceState.Nonce != 1 {
		t.Errorf("expected Alice balance 600, nonce 1; got balance %d, nonce %d", aliceState.Balance, aliceState.Nonce)
	}
	if bobState.Balance != 400 || bobState.Nonce != 0 {
		t.Errorf("expected Bob balance 400, nonce 0; got balance %d, nonce %d", bobState.Balance, bobState.Nonce)
	}

	// 5. Test validation failures
	// Insufficient balance
	err = l.ApplyTransaction(ctx, alice, bob, 700, 2)
	if err == nil {
		t.Error("expected error for insufficient balance transaction, got nil")
	}

	// Nonce mismatch
	err = l.ApplyTransaction(ctx, alice, bob, 100, 3) // expecting nonce 2, got 3
	if err == nil {
		t.Error("expected error for nonce mismatch transaction, got nil")
	}

	// 6. Test Rollback
	err = l.RollbackState(ctx)
	if err != nil {
		t.Fatalf("rollback failed: %v", err)
	}

	// State should be back to 0 because we never committed
	aliceState, _ = l.GetAccountState(ctx, alice)
	if aliceState.Balance != 0 {
		t.Errorf("expected Alice balance 0 after rollback, got %d", aliceState.Balance)
	}

	// 7. Apply again and Commit
	_ = l.ApplyTransaction(ctx, SystemAddress, alice, 1000, 0)
	_ = l.ApplyTransaction(ctx, alice, bob, 400, 1)
	err = l.CommitState(ctx, 1)
	if err != nil {
		t.Fatalf("commit failed: %v", err)
	}

	// 8. Verify persisted in repo
	persistedAlice, err := repo.GetAccountState(ctx, alice)
	if err != nil {
		t.Fatalf("failed to load Alice from repo: %v", err)
	}
	if persistedAlice.Balance != 600 || persistedAlice.Nonce != 1 {
		t.Errorf("expected persisted Alice balance 600, nonce 1; got %d, %d", persistedAlice.Balance, persistedAlice.Nonce)
	}

	persistedBob, err := repo.GetAccountState(ctx, bob)
	if err != nil {
		t.Fatalf("failed to load Bob from repo: %v", err)
	}
	if persistedBob.Balance != 400 || persistedBob.Nonce != 0 {
		t.Errorf("expected persisted Bob balance 400, nonce 0; got %d, %d", persistedBob.Balance, persistedBob.Nonce)
	}
}
