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
	if err := mgr.ReserveINR(ctx, userA, 200.0); err != nil {
		t.Fatalf("failed to reserve INR: %v", err)
	}
	accA, _ = mgr.GetAccount(ctx, userA)
	if accA.Available != 800.0 || accA.Reserved != 200.0 {
		t.Errorf("expected available 800 & reserved 200, got avail %f res %f", accA.Available, accA.Reserved)
	}

	// 3. Insufficient Reserve Failure
	if err := mgr.ReserveINR(ctx, userA, 5000.0); err != fiat.ErrInsufficientINRBalance {
		t.Errorf("expected ErrInsufficientINRBalance, got %v", err)
	}

	// 4. Release INR
	if err := mgr.ReleaseINR(ctx, userA, 50.0); err != nil {
		t.Fatalf("failed to release INR: %v", err)
	}
	accA, _ = mgr.GetAccount(ctx, userA)
	if accA.Available != 850.0 || accA.Reserved != 150.0 {
		t.Errorf("expected available 850 & reserved 150, got avail %f res %f", accA.Available, accA.Reserved)
	}

	// 5. Transfer INR (Trade Settlement with price surplus return)
	// Buyer ordered at reserved 150, trade executed at 140 (execAmount=140, reservedDeduction=150)
	if err := mgr.TransferINR(ctx, userA, userB, 140.0, 150.0); err != nil {
		t.Fatalf("failed to transfer INR: %v", err)
	}

	accA, _ = mgr.GetAccount(ctx, userA)
	accB, _ := mgr.GetAccount(ctx, userB)

	// Buyer reserved 150 used up, surplus 10 returned to available -> 850 + 10 = 860 available
	if accA.Reserved != 0.0 || accA.Available != 860.0 {
		t.Errorf("expected buyer avail 860 & reserved 0, got avail %f res %f", accA.Available, accA.Reserved)
	}
	// Seller credited 140 -> 1000 + 140 = 1140
	if accB.Available != 1140.0 {
		t.Errorf("expected seller avail 1140, got %f", accB.Available)
	}

	// 6. Pool total
	pool, err := mgr.GetTotalPoolReserve(ctx)
	if err != nil {
		t.Fatalf("failed to get pool reserve: %v", err)
	}
	expectedPool := 860.0 + 1140.0
	if pool != expectedPool {
		t.Errorf("expected pool total %f, got %f", expectedPool, pool)
	}
}
