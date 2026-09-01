package stability

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/storage"
	"github.com/n3055/liquid-link/internal/token"
)

type engineImpl struct {
	mu           sync.RWMutex
	store        storage.Store
	assetMgr     token.AssetManager
	currentPrice float64
	epoch        uint64
	priceHistory []PricePoint
	totalStaked  token.Amount
	totalBonds   token.Amount
}

// NewEngine creates and initializes a new Seigniorage Stability Engine.
func NewEngine(store storage.Store, assetMgr token.AssetManager) Engine {
	eng := &engineImpl{
		store:        store,
		assetMgr:     assetMgr,
		currentPrice: 1.02, // Initial price: Slight expansion (₹1.02)
		epoch:        1,
		totalStaked:  50000,
		totalBonds:   12000,
	}

	// Generate initial price history for charting
	now := time.Now().Unix()
	basePrices := []float64{0.97, 0.98, 0.99, 1.01, 1.03, 1.02, 0.99, 1.00, 1.04, 1.02}
	for i, p := range basePrices {
		eng.priceHistory = append(eng.priceHistory, PricePoint{
			Timestamp:   now - int64((len(basePrices)-i)*300),
			Price:       p,
			Volume:      uint64(15000 + (i * 2500)),
			TotalSupply: token.Amount(1000000 + (i * 12000)),
			TotalBonds:  token.Amount(12000 - (i * 500)),
			TotalShares: token.Amount(50000 + (i * 2000)),
		})
	}

	return eng
}

func (e *engineImpl) GetStatus(ctx context.Context) (*Status, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	state := e.determineState(e.currentPrice)
	apy := e.calculateAPY(e.currentPrice, e.totalStaked)

	return &Status{
		Price:                 e.currentPrice,
		TargetPeg:             1.00,
		State:                 state,
		TotalStakedShares:     e.totalStaked,
		TotalOutstandingBonds: e.totalBonds,
		CurrentAPY:            apy,
		Epoch:                 e.epoch,
	}, nil
}

func (e *engineImpl) GetPriceHistory(ctx context.Context) ([]PricePoint, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	history := make([]PricePoint, len(e.priceHistory))
	copy(history, e.priceHistory)
	return history, nil
}

func (e *engineImpl) StakeShares(ctx context.Context, user crypto.Address, amount token.Amount) (*StakingPosition, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if amount == 0 {
		return nil, fmt.Errorf("positive staking amount required")
	}

	// Debit LQD from user account
	if e.assetMgr != nil {
		if err := e.assetMgr.Debit(ctx, user, amount); err != nil {
			return nil, fmt.Errorf("failed to debit LQD for staking: %w", err)
		}
	}

	pos, err := e.loadStakingPosition(ctx, user)
	if err != nil {
		pos = &StakingPosition{Address: user, LastEpoch: e.epoch}
	}

	pos.StakedShares += amount
	e.totalStaked += amount

	if err := e.saveStakingPosition(ctx, pos); err != nil {
		return nil, err
	}

	return pos, nil
}

func (e *engineImpl) UnstakeShares(ctx context.Context, user crypto.Address, amount token.Amount) (*StakingPosition, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	pos, err := e.loadStakingPosition(ctx, user)
	if err != nil || pos.StakedShares < amount {
		return nil, ErrInsufficientShares
	}

	pos.StakedShares -= amount
	e.totalStaked -= amount

	// Return staked LQD + accrued yield to user
	payout := amount + pos.AccruedYield
	pos.AccruedYield = 0

	if e.assetMgr != nil {
		if err := e.assetMgr.Credit(ctx, user, payout); err != nil {
			return nil, fmt.Errorf("failed to credit unstaked payout: %w", err)
		}
	}

	if err := e.saveStakingPosition(ctx, pos); err != nil {
		return nil, err
	}

	return pos, nil
}

func (e *engineImpl) GetStakingPosition(ctx context.Context, user crypto.Address) (*StakingPosition, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.loadStakingPosition(ctx, user)
}

func (e *engineImpl) BuyBonds(ctx context.Context, user crypto.Address, lqdAmount token.Amount) (*BondPosition, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if lqdAmount == 0 {
		return nil, fmt.Errorf("positive amount required to buy bonds")
	}

	// Check if in contraction state
	state := e.determineState(e.currentPrice)
	if state != StateContraction && e.currentPrice >= 1.00 {
		return nil, ErrNotContraction
	}

	// Bond discount logic: 1 LQD buys 1.15 LQD-Bonds during contraction
	discountMultiplier := 1.15
	bondAmount := token.Amount(math.Round(float64(lqdAmount) * discountMultiplier))

	// Burn/Debit LQD from user
	if e.assetMgr != nil {
		if err := e.assetMgr.Debit(ctx, user, lqdAmount); err != nil {
			return nil, fmt.Errorf("failed to debit LQD for bond purchase: %w", err)
		}
	}

	bondPos, err := e.loadBondPosition(ctx, user)
	if err != nil {
		bondPos = &BondPosition{Address: user, PurchasePrice: e.currentPrice}
	}

	bondPos.BondsHeld += bondAmount
	bondPos.RedemptionVal += bondAmount
	e.totalBonds += bondAmount

	if err := e.saveBondPosition(ctx, bondPos); err != nil {
		return nil, err
	}

	return bondPos, nil
}

func (e *engineImpl) RedeemBonds(ctx context.Context, user crypto.Address, bondAmount token.Amount) (*BondPosition, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	bondPos, err := e.loadBondPosition(ctx, user)
	if err != nil || bondPos.BondsHeld < bondAmount {
		return nil, ErrInsufficientBonds
	}

	state := e.determineState(e.currentPrice)
	if state != StateExpansion && e.currentPrice <= 1.00 {
		return nil, ErrNotExpansion
	}

	bondPos.BondsHeld -= bondAmount
	if e.totalBonds >= bondAmount {
		e.totalBonds -= bondAmount
	}

	// 1 LQD-Bond redeems for 1 LQD + 5% expansion bonus
	redemptionPayout := token.Amount(math.Round(float64(bondAmount) * 1.05))

	if e.assetMgr != nil {
		if err := e.assetMgr.Credit(ctx, user, redemptionPayout); err != nil {
			return nil, fmt.Errorf("failed to credit bond redemption payout: %w", err)
		}
	}

	if err := e.saveBondPosition(ctx, bondPos); err != nil {
		return nil, err
	}

	return bondPos, nil
}

func (e *engineImpl) GetBondPosition(ctx context.Context, user crypto.Address) (*BondPosition, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.loadBondPosition(ctx, user)
}

func (e *engineImpl) SetMarketPrice(price float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if price <= 0 {
		return
	}
	e.currentPrice = math.Round(price*100) / 100
	var supp token.Amount = 1000000
	if e.assetMgr != nil {
		if s, err := e.assetMgr.TotalSupply(context.Background()); err == nil {
			supp = s
		}
	}
	e.priceHistory = append(e.priceHistory, PricePoint{
		Timestamp:   time.Now().Unix(),
		Price:       e.currentPrice,
		Volume:      uint64(25000 + rand.Intn(5000)),
		TotalSupply: supp,
		TotalBonds:  e.totalBonds,
		TotalShares: e.totalStaked,
	})
	if len(e.priceHistory) > 30 {
		e.priceHistory = e.priceHistory[1:]
	}
}

func (e *engineImpl) ProcessEpoch(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.epoch++

	// Simulate organic price movement around ₹1.00 peg
	delta := (rand.Float64() - 0.48) * 0.03
	e.currentPrice = math.Round((e.currentPrice+delta)*100) / 100

	// Algorithmic Seigniorage Expansion / Contraction
	if e.currentPrice > 1.005 {
		// Expansion: Algorithmic minting to satisfy demand & distribute yield
		expansionAmount := token.Amount(math.Round((e.currentPrice - 1.00) * 10000))
		if e.assetMgr != nil && expansionAmount > 0 {
			_ = e.assetMgr.Mint(ctx, crypto.Address("0xseigniorage_treasury"), expansionAmount)
		}
	}

	// Add to price history
	var supp token.Amount = 1000000
	if e.assetMgr != nil {
		if s, err := e.assetMgr.TotalSupply(ctx); err == nil {
			supp = s
		}
	}
	e.priceHistory = append(e.priceHistory, PricePoint{
		Timestamp:   time.Now().Unix(),
		Price:       e.currentPrice,
		Volume:      uint64(20000 + rand.Intn(10000)),
		TotalSupply: supp,
		TotalBonds:  e.totalBonds,
		TotalShares: e.totalStaked,
	})
	if len(e.priceHistory) > 30 {
		e.priceHistory = e.priceHistory[1:]
	}

	return nil
}

func (e *engineImpl) StartBackgroundWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				_ = e.ProcessEpoch(ctx)
			}
		}
	}()
}

func (e *engineImpl) determineState(price float64) PegState {
	if price > 1.005 {
		return StateExpansion
	} else if price < 0.995 {
		return StateContraction
	}
	return StatePegged
}

func (e *engineImpl) calculateAPY(price float64, totalStaked token.Amount) float64 {
	baseAPY := 14.5
	if price > 1.00 {
		return baseAPY + ((price - 1.00) * 100.0) // Expansion bonus APY
	}
	return baseAPY
}

func (e *engineImpl) loadStakingPosition(ctx context.Context, user crypto.Address) (*StakingPosition, error) {
	if e.store == nil {
		return &StakingPosition{Address: user, LastEpoch: e.epoch}, nil
	}
	key := []byte(fmt.Sprintf("stability:stake:%s", user))
	data, err := e.store.Get(ctx, key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return &StakingPosition{Address: user, LastEpoch: e.epoch}, nil
		}
		return nil, err
	}
	var pos StakingPosition
	if err := json.Unmarshal(data, &pos); err != nil {
		return nil, err
	}
	return &pos, nil
}

func (e *engineImpl) saveStakingPosition(ctx context.Context, pos *StakingPosition) error {
	if e.store == nil {
		return nil
	}
	encoded, err := json.Marshal(pos)
	if err != nil {
		return err
	}
	key := []byte(fmt.Sprintf("stability:stake:%s", pos.Address))
	return e.store.Set(ctx, key, encoded)
}

func (e *engineImpl) loadBondPosition(ctx context.Context, user crypto.Address) (*BondPosition, error) {
	if e.store == nil {
		return &BondPosition{Address: user}, nil
	}
	key := []byte(fmt.Sprintf("stability:bond:%s", user))
	data, err := e.store.Get(ctx, key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return &BondPosition{Address: user}, nil
		}
		return nil, err
	}
	var pos BondPosition
	if err := json.Unmarshal(data, &pos); err != nil {
		return nil, err
	}
	return &pos, nil
}

func (e *engineImpl) saveBondPosition(ctx context.Context, pos *BondPosition) error {
	if e.store == nil {
		return nil
	}
	encoded, err := json.Marshal(pos)
	if err != nil {
		return err
	}
	key := []byte(fmt.Sprintf("stability:bond:%s", pos.Address))
	return e.store.Set(ctx, key, encoded)
}
