package fiat_test

import (
	"context"
	"testing"

	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/fiat"
)

func TestFiatManagerOperations(t *testing.T) {
	mgr := fiat.NewManager()
	ctx := context.Background()

	userA := crypto.Address("0xuser_a")
	userB := crypto.Address("0xuser_b")

	// 1. Initial balance check
	accA, err := mgr.GetAccount(ctx, userA)
	if err != nil {
		t.Fatalf("unexpected error getting account: %v", err)
	}
	if accA.Available != fiat.DefaultInitialINRBalance {
		t.Errorf("expected initial available %f, got %f", fiat.DefaultInitialINRBalance, accA.Available)
	}

	// 2. Reserve INR
	if err := mgr.ReserveINR(ctx, userA, 2000.0); err != nil {
		t.Fatalf("failed to reserve INR: %v", err)
	}
	accA, _ = mgr.GetAccount(ctx, userA)
	if accA.Available != 8000.0 || accA.Reserved != 2000.0 {
		t.Errorf("expected available 8000 & reserved 2000, got avail %f res %f", accA.Available, accA.Reserved)
	}

	// 3. Insufficient Reserve Failure
	if err := mgr.ReserveINR(ctx, userA, 10000.0); err != fiat.ErrInsufficientINRBalance {
		t.Errorf("expected ErrInsufficientINRBalance, got %v", err)
	}

	// 4. Release INR
	if err := mgr.ReleaseINR(ctx, userA, 500.0); err != nil {
		t.Fatalf("failed to release INR: %v", err)
	}
	accA, _ = mgr.GetAccount(ctx, userA)
	if accA.Available != 8500.0 || accA.Reserved != 1500.0 {
		t.Errorf("expected available 8500 & reserved 1500, got avail %f res %f", accA.Available, accA.Reserved)
	}

	// 5. Transfer INR (Trade Settlement with price surplus return)
	// Buyer ordered at reserved 1500, trade executed at 1400 (execAmount=1400, reservedDeduction=1500)
	if err := mgr.TransferINR(ctx, userA, userB, 1400.0, 1500.0); err != nil {
		t.Fatalf("failed to transfer INR: %v", err)
	}

	accA, _ = mgr.GetAccount(ctx, userA)
	accB, _ := mgr.GetAccount(ctx, userB)

	// Buyer reserved 1500 used up, surplus 100 returned to available -> 8500 + 100 = 8600 available
	if accA.Reserved != 0.0 || accA.Available != 8600.0 {
		t.Errorf("expected buyer avail 8600 & reserved 0, got avail %f res %f", accA.Available, accA.Reserved)
	}
	// Seller credited 1400 -> 10000 + 1400 = 11400
	if accB.Available != 11400.0 {
		t.Errorf("expected seller avail 11400, got %f", accB.Available)
	}

	// 6. Pool total
	pool, err := mgr.GetTotalPoolReserve(ctx)
	if err != nil {
		t.Fatalf("failed to get pool reserve: %v", err)
	}
	expectedPool := 8600.0 + 11400.0
	if pool != expectedPool {
		t.Errorf("expected pool total %f, got %f", expectedPool, pool)
	}
}
