package token

import (
	"context"
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/n3055/liquid-link/internal/crypto"
)

// mockBalanceRepository implements BalanceRepository in memory.
type mockBalanceRepository struct {
	balances map[crypto.Address]Amount
}

func newMockBalanceRepository() *mockBalanceRepository {
	return &mockBalanceRepository{
		balances: make(map[crypto.Address]Amount),
	}
}

func (r *mockBalanceRepository) GetBalance(ctx context.Context, addr crypto.Address) (Amount, error) {
	bal, exists := r.balances[addr]
	if !exists {
		return 0, badger.ErrKeyNotFound
	}
	return bal, nil
}

func (r *mockBalanceRepository) SetBalance(ctx context.Context, addr crypto.Address, balance Amount) error {
	r.balances[addr] = balance
	return nil
}

// mockStore implements storage.Store in memory using a simple map.
type mockStore struct {
	data map[string][]byte
}

func newMockStore() *mockStore {
	return &mockStore{
		data: make(map[string][]byte),
	}
}

func (s *mockStore) Set(ctx context.Context, key, value []byte) error {
	s.data[string(key)] = value
	return nil
}

func (s *mockStore) Get(ctx context.Context, key []byte) ([]byte, error) {
	val, exists := s.data[string(key)]
	if !exists {
		return nil, badger.ErrKeyNotFound
	}
	return val, nil
}

func (s *mockStore) Delete(ctx context.Context, key []byte) error {
	delete(s.data, string(key))
	return nil
}

func (s *mockStore) Close() error {
	return nil
}

func (s *mockStore) View(fn func(txn *badger.Txn) error) error {
	return nil
}

func (s *mockStore) Update(fn func(txn *badger.Txn) error) error {
	return nil
}

func TestAssetManagerMintBurnCreditDebit(t *testing.T) {
	repo := newMockBalanceRepository()
	store := newMockStore()
	manager := NewAssetManager(repo, store)
	ctx := context.Background()

	alice := crypto.Address("0xalice")
	bob := crypto.Address("0xbob")

	// 1. Initial supply and balances should be 0
	supply, err := manager.TotalSupply(ctx)
	if err != nil {
		t.Fatalf("failed to get total supply: %v", err)
	}
	if supply != 0 {
		t.Errorf("expected initial supply 0, got %d", supply)
	}

	bal, err := manager.BalanceOf(ctx, alice)
	if err != nil {
		t.Fatalf("failed to check Alice balance: %v", err)
	}
	if bal != 0 {
		t.Errorf("expected initial balance 0, got %d", bal)
	}

	// 2. Mint tokens to Alice
	err = manager.Mint(ctx, alice, 1000)
	if err != nil {
		t.Fatalf("failed to mint: %v", err)
	}

	// Verify balance and total supply
	bal, _ = manager.BalanceOf(ctx, alice)
	if bal != 1000 {
		t.Errorf("expected Alice balance 1000, got %d", bal)
	}

	supply, _ = manager.TotalSupply(ctx)
	if supply != 1000 {
		t.Errorf("expected total supply 1000, got %d", supply)
	}

	// 3. Credit Bob
	err = manager.Credit(ctx, bob, 500)
	if err != nil {
		t.Fatalf("failed to credit: %v", err)
	}
	bal, _ = manager.BalanceOf(ctx, bob)
	if bal != 500 {
		t.Errorf("expected Bob balance 500, got %d", bal)
	}
	// Credit shouldn't affect supply directly (it represents direct state adjustment)
	supply, _ = manager.TotalSupply(ctx)
	if supply != 1000 {
		t.Errorf("expected supply 1000 after Bob credit, got %d", supply)
	}

	// 4. Debit Bob
	err = manager.Debit(ctx, bob, 200)
	if err != nil {
		t.Fatalf("failed to debit: %v", err)
	}
	bal, _ = manager.BalanceOf(ctx, bob)
	if bal != 300 {
		t.Errorf("expected Bob balance 300 after debit, got %d", bal)
	}

	// 5. Burn tokens from Alice
	err = manager.Burn(ctx, alice, 400)
	if err != nil {
		t.Fatalf("failed to burn: %v", err)
	}

	bal, _ = manager.BalanceOf(ctx, alice)
	if bal != 600 {
		t.Errorf("expected Alice balance 600 after burn, got %d", bal)
	}

	supply, _ = manager.TotalSupply(ctx)
	if supply != 600 {
		t.Errorf("expected supply 600 after burn, got %d", supply)
	}

	// 6. Test invalid operations
	// Insufficient balance on debit
	err = manager.Debit(ctx, bob, 500)
	if err == nil {
		t.Error("expected error debiting Bob beyond balance, got nil")
	}

	// Insufficient balance on burn
	err = manager.Burn(ctx, alice, 1000)
	if err == nil {
		t.Error("expected error burning Alice tokens beyond balance, got nil")
	}

	// Zero amounts
	err = manager.Mint(ctx, alice, 0)
	if err == nil {
		t.Error("expected error minting zero tokens, got nil")
	}

	err = manager.Burn(ctx, alice, 0)
	if err == nil {
		t.Error("expected error burning zero tokens, got nil")
	}
}
