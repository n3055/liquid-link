package token

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/dgraph-io/badger/v4"
	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/storage"
)

// assetManagerImpl implements the AssetManager interface.
type assetManagerImpl struct {
	repo  BalanceRepository
	store storage.Store
}

// NewAssetManager creates a new AssetManager instance.
func NewAssetManager(repo BalanceRepository, store storage.Store) AssetManager {
	return &assetManagerImpl{
		repo:  repo,
		store: store,
	}
}

// Mint creates new tokens out of thin air and credits them to the specified address.
func (m *assetManagerImpl) Mint(ctx context.Context, to crypto.Address, amount Amount) error {
	if amount == 0 {
		return errors.New("mint amount must be greater than zero")
	}

	// 1. Credit the receiver
	if err := m.Credit(ctx, to, amount); err != nil {
		return fmt.Errorf("failed to credit receiver: %w", err)
	}

	// 2. Update Total Supply
	supply, err := m.TotalSupply(ctx)
	if err != nil {
		return err
	}

	newSupply := supply + amount
	if err := m.setTotalSupply(ctx, newSupply); err != nil {
		return fmt.Errorf("failed to update total supply: %w", err)
	}

	return nil
}

// Burn permanently removes tokens from circulation.
func (m *assetManagerImpl) Burn(ctx context.Context, from crypto.Address, amount Amount) error {
	if amount == 0 {
		return errors.New("burn amount must be greater than zero")
	}

	// 1. Debit the sender
	if err := m.Debit(ctx, from, amount); err != nil {
		return fmt.Errorf("failed to debit sender: %w", err)
	}

	// 2. Update Total Supply
	supply, err := m.TotalSupply(ctx)
	if err != nil {
		return err
	}

	if supply < amount {
		return errors.New("burn amount exceeds total supply (invalid state)")
	}

	newSupply := supply - amount
	if err := m.setTotalSupply(ctx, newSupply); err != nil {
		return fmt.Errorf("failed to update total supply: %w", err)
	}

	return nil
}

// BalanceOf returns the current balance of a specific address.
func (m *assetManagerImpl) BalanceOf(ctx context.Context, addr crypto.Address) (Amount, error) {
	bal, err := m.repo.GetBalance(ctx, addr)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) || os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}

	return bal, nil
}

// Credit increases the balance of an address by the specified amount.
func (m *assetManagerImpl) Credit(ctx context.Context, addr crypto.Address, amount Amount) error {
	if amount == 0 {
		return errors.New("credit amount must be greater than zero")
	}

	bal, err := m.repo.GetBalance(ctx, addr)
	if err != nil && !errors.Is(err, badger.ErrKeyNotFound) && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load balance: %w", err)
	}

	newBal := bal + amount
	if err := m.repo.SetBalance(ctx, addr, newBal); err != nil {
		return fmt.Errorf("failed to set balance: %w", err)
	}

	return nil
}

// Debit decreases the balance of an address by the specified amount.
func (m *assetManagerImpl) Debit(ctx context.Context, addr crypto.Address, amount Amount) error {
	if amount == 0 {
		return errors.New("debit amount must be greater than zero")
	}

	bal, err := m.repo.GetBalance(ctx, addr)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) || os.IsNotExist(err) {
			return fmt.Errorf("account not found: %s", addr)
		}
		return fmt.Errorf("failed to load balance: %w", err)
	}

	if bal < amount {
		return fmt.Errorf("insufficient balance: account has %d, required %d", bal, amount)
	}

	newBal := bal - amount
	if err := m.repo.SetBalance(ctx, addr, newBal); err != nil {
		return fmt.Errorf("failed to set balance: %w", err)
	}

	return nil
}

// TotalSupply returns the total circulating supply of the token.
func (m *assetManagerImpl) TotalSupply(ctx context.Context) (Amount, error) {
	key := []byte("token:total_supply")
	val, err := m.store.Get(ctx, key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) || os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get total supply: %w", err)
	}

	var supply Amount
	if err := json.Unmarshal(val, &supply); err != nil {
		return 0, fmt.Errorf("failed to unmarshal total supply: %w", err)
	}

	return supply, nil
}

// setTotalSupply persists the total supply key inside BadgerDB.
func (m *assetManagerImpl) setTotalSupply(ctx context.Context, supply Amount) error {
	key := []byte("token:total_supply")
	val, err := json.Marshal(supply)
	if err != nil {
		return fmt.Errorf("failed to marshal supply: %w", err)
	}

	return m.store.Set(ctx, key, val)
}
