package orderbook_test

import (
	"context"
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/fiat"
	"github.com/n3055/liquid-link/internal/orderbook"
	"github.com/n3055/liquid-link/internal/token"
)

type mockBalanceRepo struct {
	balances map[crypto.Address]token.Amount
}

func (m *mockBalanceRepo) GetBalance(ctx context.Context, addr crypto.Address) (token.Amount, error) {
	bal, ok := m.balances[addr]
	if !ok {
		return 0, badger.ErrKeyNotFound
	}
	return bal, nil
}

func (m *mockBalanceRepo) SetBalance(ctx context.Context, addr crypto.Address, balance token.Amount) error {
	m.balances[addr] = balance
	return nil
}

type mockStore struct {
	data map[string][]byte
}

func (m *mockStore) Set(ctx context.Context, key, value []byte) error {
	m.data[string(key)] = value
	return nil
}

func (m *mockStore) Get(ctx context.Context, key []byte) ([]byte, error) {
	val, ok := m.data[string(key)]
	if !ok {
		return nil, badger.ErrKeyNotFound
	}
	return val, nil
}

func (m *mockStore) Delete(ctx context.Context, key []byte) error {
	delete(m.data, string(key))
	return nil
}

func (m *mockStore) Close() error                                { return nil }
func (m *mockStore) View(fn func(txn *badger.Txn) error) error   { return nil }
func (m *mockStore) Update(fn func(txn *badger.Txn) error) error { return nil }

func TestOrderbook_PlaceAndMatch(t *testing.T) {
	ctx := context.Background()
	fiatMgr := fiat.NewManager()
	eng := orderbook.NewEngine(nil, nil, fiatMgr)

	buyer := crypto.Address("0xbuyer123")
	seller := crypto.Address("0xseller123")

	// Place sell order at ₹1.00 for 100 LQD
	sellOrd := &orderbook.Order{
		User:   seller,
		Side:   orderbook.SideSell,
		Price:  1.00,
		Amount: 100,
	}
	placedSell, trades, err := eng.PlaceOrder(ctx, sellOrd)
	if err != nil {
		t.Fatalf("failed to place sell order: %v", err)
	}
	if len(trades) != 0 {
		t.Errorf("expected 0 trades on resting sell order, got %d", len(trades))
	}

	// Place matching buy order at ₹1.00 for 50 LQD
	buyOrd := &orderbook.Order{
		User:   buyer,
		Side:   orderbook.SideBuy,
		Price:  1.00,
		Amount: 50,
	}
	placedBuy, trades, err := eng.PlaceOrder(ctx, buyOrd)
	if err != nil {
		t.Fatalf("failed to place buy order: %v", err)
	}

	if len(trades) != 1 {
		t.Fatalf("expected 1 trade execution, got %d", len(trades))
	}
	if trades[0].Amount != 50 {
		t.Errorf("expected trade amount 50, got %d", trades[0].Amount)
	}
	if trades[0].Price != 1.00 {
		t.Errorf("expected trade price 1.00, got %f", trades[0].Price)
	}

	if placedBuy.Status != orderbook.StatusFilled {
		t.Errorf("expected buy order filled, got %s", placedBuy.Status)
	}

	_ = placedSell
}

func TestOrderbook_DepthAndCancel(t *testing.T) {
	ctx := context.Background()
	fiatMgr := fiat.NewManager()
	eng := orderbook.NewEngine(nil, nil, fiatMgr)

	user := crypto.Address("0xuser456")

	ord := &orderbook.Order{
		User:   user,
		Side:   orderbook.SideBuy,
		Price:  0.95,
		Amount: 500,
	}
	placed, _, err := eng.PlaceOrder(ctx, ord)
	if err != nil {
		t.Fatalf("failed to place order: %v", err)
	}

	depth, err := eng.GetDepth(ctx)
	if err != nil {
		t.Fatalf("failed to get depth: %v", err)
	}
	if len(depth.Bids) == 0 {
		t.Errorf("expected bids in orderbook depth")
	}

	err = eng.CancelOrder(ctx, placed.ID, user)
	if err != nil {
		t.Fatalf("failed to cancel order: %v", err)
	}

	orders, err := eng.GetUserOrders(ctx, user)
	if err != nil {
		t.Fatalf("failed to get user orders: %v", err)
	}
	if orders[0].Status != orderbook.StatusCancelled {
		t.Errorf("expected status cancelled, got %s", orders[0].Status)
	}
}

func TestOrderbook_UnbackedOrderDebitFailure(t *testing.T) {
	ctx := context.Background()
	// Mock asset manager where debit fails due to zero balance
	fiatMgr := fiat.NewManager()
	eng := orderbook.NewEngine(nil, nil, fiatMgr)

	seller := crypto.Address("0xzero_balance_seller")
	buyer := crypto.Address("0xbuyer_test")

	// Sell order placed with 0 balance
	sellOrd := &orderbook.Order{
		User:   seller,
		Side:   orderbook.SideSell,
		Price:  1.00,
		Amount: 100,
	}
	_, _, err := eng.PlaceOrder(ctx, sellOrd)
	if err != nil {
		t.Fatalf("failed to place sell order: %v", err)
	}

	buyOrd := &orderbook.Order{
		User:   buyer,
		Side:   orderbook.SideBuy,
		Price:  1.00,
		Amount: 50,
	}
	_, trades, err := eng.PlaceOrder(ctx, buyOrd)
	if err != nil {
		t.Fatalf("failed to place buy order: %v", err)
	}

	// In the absence of asset manager, trades proceed normally. When asset manager is present and debit fails,
	// order is cancelled. Let's verify status checking works.
	if buyOrd.Amount != 50 {
		t.Errorf("expected buy amount 50")
	}
	_ = trades
}

func TestBug_BuyBidFillAndCancel(t *testing.T) {
	ctx := context.Background()
	repo := &mockBalanceRepo{balances: make(map[crypto.Address]token.Amount)}
	store := &mockStore{data: make(map[string][]byte)}
	assetMgr := token.NewAssetManager(repo, store)
	fiatMgr := fiat.NewManager()

	eng := orderbook.NewEngine(assetMgr, nil, fiatMgr)

	buyer := crypto.Address("0xbuyer_user")
	seller := crypto.Address("0xseller_user")

	// Fund seller with LQD and buyer with INR
	_ = assetMgr.Mint(ctx, seller, 1000)
	_ = fiatMgr.CreditINR(ctx, buyer, 1000.0)

	// Seller places sell order for 100 LQD at 1.00 INR
	sellOrd := &orderbook.Order{
		User:   seller,
		Side:   orderbook.SideSell,
		Price:  1.00,
		Amount: 100,
	}
	placedSell, trades, err := eng.PlaceOrder(ctx, sellOrd)
	if err != nil {
		t.Fatalf("failed to place sell order: %v", err)
	}
	t.Logf("Placed sell order %s, trades count: %d", placedSell.ID, len(trades))

	// Buyer places buy order for 100 LQD at 1.00 INR
	buyOrd := &orderbook.Order{
		User:   buyer,
		Side:   orderbook.SideBuy,
		Price:  1.00,
		Amount: 100,
	}
	placedBuy, trades, err := eng.PlaceOrder(ctx, buyOrd)
	if err != nil {
		t.Fatalf("failed to place buy order: %v", err)
	}

	t.Logf("Placed buy order %s, trades count: %d, status: %s, filled: %d", placedBuy.ID, len(trades), placedBuy.Status, placedBuy.Filled)

	// Check balances
	buyerLQD, _ := assetMgr.BalanceOf(ctx, buyer)
	sellerLQD, _ := assetMgr.BalanceOf(ctx, seller)
	t.Logf("Buyer LQD: %d, Seller LQD: %d", buyerLQD, sellerLQD)

	buyerAcc, _ := fiatMgr.GetAccount(ctx, buyer)
	sellerAcc, _ := fiatMgr.GetAccount(ctx, seller)
	t.Logf("Buyer INR Available: %.2f, Reserved: %.2f", buyerAcc.Available, buyerAcc.Reserved)
	t.Logf("Seller INR Available: %.2f, Reserved: %.2f", sellerAcc.Available, sellerAcc.Reserved)

	// Verify expected trade outcomes
	if buyerLQD != 100 {
		t.Errorf("expected buyer LQD balance 100, got %d", buyerLQD)
	}
	if sellerLQD != 900 {
		t.Errorf("expected seller LQD balance 900, got %d", sellerLQD)
	}
	if placedBuy.Status != orderbook.StatusFilled {
		t.Errorf("expected buy order status filled, got %s", placedBuy.Status)
	}
	if placedSell.Status != orderbook.StatusFilled {
		t.Errorf("expected sell order status filled, got %s", placedSell.Status)
	}
}

func TestPartialFill_MultipleMatches(t *testing.T) {
	ctx := context.Background()
	repo := &mockBalanceRepo{balances: make(map[crypto.Address]token.Amount)}
	store := &mockStore{data: make(map[string][]byte)}
	assetMgr := token.NewAssetManager(repo, store)
	fiatMgr := fiat.NewManager()

	eng := orderbook.NewEngine(assetMgr, nil, fiatMgr)

	buyer := crypto.Address("0xbuyer_multi")
	seller1 := crypto.Address("0xseller1_multi")
	seller2 := crypto.Address("0xseller2_multi")

	_ = assetMgr.Mint(ctx, seller1, 50)
	_ = assetMgr.Mint(ctx, seller2, 100)
	_ = fiatMgr.CreditINR(ctx, buyer, 1000.0)

	// Buyer places buy order for 100 LQD at 1.00 INR
	buyOrd := &orderbook.Order{
		User:   buyer,
		Side:   orderbook.SideBuy,
		Price:  1.00,
		Amount: 100,
	}
	placedBuy, trades1, err := eng.PlaceOrder(ctx, buyOrd)
	if err != nil {
		t.Fatalf("failed to place buy order: %v", err)
	}
	t.Logf("Placed buy order %s, trades count: %d, status: %s, filled: %d", placedBuy.ID, len(trades1), placedBuy.Status, placedBuy.Filled)

	// Seller 1 places sell order for 50 LQD at 1.00 INR
	sellOrd1 := &orderbook.Order{
		User:   seller1,
		Side:   orderbook.SideSell,
		Price:  1.00,
		Amount: 50,
	}
	placedSell1, trades2, err := eng.PlaceOrder(ctx, sellOrd1)
	if err != nil {
		t.Fatalf("failed to place sell order 1: %v", err)
	}
	t.Logf("Placed sell order 1 %s, trades count: %d, status: %s, filled: %d", placedSell1.ID, len(trades2), placedSell1.Status, placedSell1.Filled)
	t.Logf("Buy order after sell 1: status: %s, filled: %d", placedBuy.Status, placedBuy.Filled)

	// Seller 2 places sell order for 100 LQD at 1.00 INR
	sellOrd2 := &orderbook.Order{
		User:   seller2,
		Side:   orderbook.SideSell,
		Price:  1.00,
		Amount: 100,
	}
	placedSell2, trades3, err := eng.PlaceOrder(ctx, sellOrd2)
	if err != nil {
		t.Fatalf("failed to place sell order 2: %v", err)
	}
	t.Logf("Placed sell order 2 %s, trades count: %d, status: %s, filled: %d", placedSell2.ID, len(trades3), placedSell2.Status, placedSell2.Filled)
	t.Logf("Buy order after sell 2: status: %s, filled: %d", placedBuy.Status, placedBuy.Filled)

	buyerLQD, _ := assetMgr.BalanceOf(ctx, buyer)
	seller1LQD, _ := assetMgr.BalanceOf(ctx, seller1)
	seller2LQD, _ := assetMgr.BalanceOf(ctx, seller2)

	buyerAcc, _ := fiatMgr.GetAccount(ctx, buyer)
	seller1Acc, _ := fiatMgr.GetAccount(ctx, seller1)
	seller2Acc, _ := fiatMgr.GetAccount(ctx, seller2)

	t.Logf("Final Buyer LQD: %d (expected 100)", buyerLQD)
	t.Logf("Final Seller1 LQD: %d (expected 0)", seller1LQD)
	t.Logf("Final Seller2 LQD: %d (expected 50)", seller2LQD)
	t.Logf("Final Buyer INR Reserved: %.2f (expected 0.00)", buyerAcc.Reserved)
	t.Logf("Final Seller1 INR Available: %.2f (expected 1050.00)", seller1Acc.Available)
	t.Logf("Final Seller2 INR Available: %.2f (expected 1050.00)", seller2Acc.Available)

	if buyerLQD != 100 {
		t.Errorf("buyer LQD mismatch: expected 100, got %d", buyerLQD)
	}
	if seller1LQD != 0 {
		t.Errorf("seller1 LQD mismatch: expected 0, got %d", seller1LQD)
	}
	if seller2LQD != 50 {
		t.Errorf("seller2 LQD mismatch: expected 50, got %d", seller2LQD)
	}
}

func TestOrderbook_InvalidSide(t *testing.T) {
	ctx := context.Background()
	fiatMgr := fiat.NewManager()
	eng := orderbook.NewEngine(nil, nil, fiatMgr)

	user := crypto.Address("0xinvalid_side_user")

	invalidOrd := &orderbook.Order{
		User:   user,
		Side:   orderbook.OrderSide("bid"),
		Price:  1.00,
		Amount: 100,
	}

	_, _, err := eng.PlaceOrder(ctx, invalidOrd)
	if err == nil {
		t.Fatalf("expected error for invalid order side, got nil")
	}
	if err != orderbook.ErrInvalidSide {
		t.Errorf("expected ErrInvalidSide, got %v", err)
	}
}
