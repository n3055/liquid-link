package routing_test

import (
	"context"
	"testing"
	"time"

	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/routing"
)

func TestRouting_FindRouteAndHTLC(t *testing.T) {
	ctx := context.Background()
	r := routing.NewRouter()

	nodeA := crypto.Address("0xnode_a")
	nodeC := crypto.Address("0xnode_c")

	// 1. Find route from A to C
	route, err := r.FindRoute(ctx, nodeA, nodeC, 1000)
	if err != nil {
		t.Fatalf("failed to find route: %v", err)
	}

	if len(route.Hops) == 0 {
		t.Errorf("expected non-empty hops route")
	}

	if route.Cost == 0 {
		t.Errorf("expected positive routing cost fee")
	}

	// 2. Generate HTLC
	htlc, err := routing.GenerateHTLC(1000, 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate HTLC: %v", err)
	}

	if len(htlc.HashLock) == 0 || len(htlc.PreImage) == 0 {
		t.Errorf("expected non-empty hashlock and pre-image")
	}
}
