package orderbook

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/stability"
	"github.com/n3055/liquid-link/internal/token"
)

type engineImpl struct {
	mu           sync.RWMutex
	assetMgr     token.AssetManager
	stabilityEng stability.Engine
	orders       map[string]*Order
	userOrders   map[crypto.Address][]*Order
	trades       []*Trade
	lastPrice    float64
	idCounter    uint64
}

// [Orderbook] NewEngine initializes a new limit orderbook matching engine.
func NewEngine(assetMgr token.AssetManager, stabilityEng stability.Engine) Engine {
	eng := &engineImpl{
		assetMgr:     assetMgr,
		stabilityEng: stabilityEng,
		orders:       make(map[string]*Order),
		userOrders:   make(map[crypto.Address][]*Order),
		trades:       make([]*Trade, 0),
		lastPrice:    1.00, // Initial peg baseline: ₹1.00
	}

	// Initialize default liquidity orderbook depth
	eng.seedInitialLiquidity()

	return eng
}

func (e *engineImpl) seedInitialLiquidity() {
	mmUser := crypto.Address("0xmarket_maker_node")
	initialOrders := []*Order{
		{User: mmUser, Side: SideBuy, Price: 0.99, Amount: 10000},
		{User: mmUser, Side: SideBuy, Price: 0.98, Amount: 15000},
		{User: mmUser, Side: SideBuy, Price: 0.97, Amount: 20000},
		{User: mmUser, Side: SideSell, Price: 1.01, Amount: 10000},
		{User: mmUser, Side: SideSell, Price: 1.02, Amount: 15000},
		{User: mmUser, Side: SideSell, Price: 1.03, Amount: 20000},
	}

	ctx := context.Background()
	for _, ord := range initialOrders {
		_, _, _ = e.PlaceOrder(ctx, ord)
	}
}

func (e *engineImpl) PlaceOrder(ctx context.Context, order *Order) (*Order, []*Trade, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if order.Price <= 0 {
		return nil, nil, ErrInvalidPrice
	}
	if order.Amount == 0 {
		return nil, nil, ErrInvalidAmount
	}

	e.idCounter++
	if order.ID == "" {
		order.ID = fmt.Sprintf("ord-%d", e.idCounter)
	}
	order.Status = StatusOpen
	order.Filled = 0
	order.CreatedAt = time.Now()

	executedTrades := make([]*Trade, 0)

	// Perform limit order matching
	if order.Side == SideBuy {
		executedTrades = e.matchBuyOrder(ctx, order)
	} else {
		executedTrades = e.matchSellOrder(ctx, order)
	}

	// Save order state
	e.orders[order.ID] = order
	e.userOrders[order.User] = append(e.userOrders[order.User], order)

	return order, executedTrades, nil
}

func (e *engineImpl) matchBuyOrder(ctx context.Context, buyOrd *Order) []*Trade {
	trades := make([]*Trade, 0)

	// Gather open sell orders with price <= buyOrd.Price
	sellOrders := make([]*Order, 0)
	for _, ord := range e.orders {
		if ord.Side == SideSell && (ord.Status == StatusOpen || ord.Status == StatusPartial) && ord.Price <= buyOrd.Price {
			sellOrders = append(sellOrders, ord)
		}
	}

	// Sort sell orders by price ascending (cheapest first), then creation time
	sort.Slice(sellOrders, func(i, j int) bool {
		if sellOrders[i].Price == sellOrders[j].Price {
			return sellOrders[i].CreatedAt.Before(sellOrders[j].CreatedAt)
		}
		return sellOrders[i].Price < sellOrders[j].Price
	})

	for _, sellOrd := range sellOrders {
		if buyOrd.Filled >= buyOrd.Amount {
			break
		}

		buyRemaining := buyOrd.Amount - buyOrd.Filled
		sellRemaining := sellOrd.Amount - sellOrd.Filled
		matchQty := buyRemaining
		if sellRemaining < matchQty {
			matchQty = sellRemaining
		}

		if matchQty == 0 {
			continue
		}

		// Execute trade at seller's (maker) price
		execPrice := sellOrd.Price

		// Execute asset manager settlement if available
		if e.assetMgr != nil {
			if err := e.assetMgr.Debit(ctx, sellOrd.User, matchQty); err != nil {
				// Seller has insufficient token balance; cancel/skip unbacked sell order
				sellOrd.Status = StatusCancelled
				continue
			}
			if err := e.assetMgr.Credit(ctx, buyOrd.User, matchQty); err != nil {
				// Rollback debit on credit failure
				_ = e.assetMgr.Credit(ctx, sellOrd.User, matchQty)
				continue
			}
		}

		buyOrd.Filled += matchQty
		sellOrd.Filled += matchQty

		if buyOrd.Filled >= buyOrd.Amount {
			buyOrd.Status = StatusFilled
		} else {
			buyOrd.Status = StatusPartial
		}

		if sellOrd.Filled >= sellOrd.Amount {
			sellOrd.Status = StatusFilled
		} else {
			sellOrd.Status = StatusPartial
		}

		e.idCounter++
		trade := &Trade{
			ID:          fmt.Sprintf("trd-%d", e.idCounter),
			BuyOrderID:  buyOrd.ID,
			SellOrderID: sellOrd.ID,
			Buyer:       buyOrd.User,
			Seller:      sellOrd.User,
			Price:       execPrice,
			Amount:      matchQty,
			Timestamp:   time.Now(),
		}

		trades = append(trades, trade)
		e.trades = append(e.trades, trade)
		e.lastPrice = execPrice

		if e.stabilityEng != nil {
			e.stabilityEng.SetMarketPrice(execPrice)
		}
	}

	return trades
}

func (e *engineImpl) matchSellOrder(ctx context.Context, sellOrd *Order) []*Trade {
	trades := make([]*Trade, 0)

	// Gather open buy orders with price >= sellOrd.Price
	buyOrders := make([]*Order, 0)
	for _, ord := range e.orders {
		if ord.Side == SideBuy && (ord.Status == StatusOpen || ord.Status == StatusPartial) && ord.Price >= sellOrd.Price {
			buyOrders = append(buyOrders, ord)
		}
	}

	// Sort buy orders by price descending (highest bid first), then creation time
	sort.Slice(buyOrders, func(i, j int) bool {
		if buyOrders[i].Price == buyOrders[j].Price {
			return buyOrders[i].CreatedAt.Before(buyOrders[j].CreatedAt)
		}
		return buyOrders[i].Price > buyOrders[j].Price
	})

	for _, buyOrd := range buyOrders {
		if sellOrd.Filled >= sellOrd.Amount {
			break
		}

		sellRemaining := sellOrd.Amount - sellOrd.Filled
		buyRemaining := buyOrd.Amount - buyOrd.Filled
		matchQty := sellRemaining
		if buyRemaining < matchQty {
			matchQty = buyRemaining
		}

		if matchQty == 0 {
			continue
		}

		// Execute trade at buyer's (maker) price
		execPrice := buyOrd.Price

		if e.assetMgr != nil {
			if err := e.assetMgr.Debit(ctx, sellOrd.User, matchQty); err != nil {
				// Seller has insufficient token balance; cancel/skip unbacked sell order
				sellOrd.Status = StatusCancelled
				continue
			}
			if err := e.assetMgr.Credit(ctx, buyOrd.User, matchQty); err != nil {
				// Rollback debit on credit failure
				_ = e.assetMgr.Credit(ctx, sellOrd.User, matchQty)
				continue
			}
		}

		sellOrd.Filled += matchQty
		buyOrd.Filled += matchQty

		if sellOrd.Filled >= sellOrd.Amount {
			sellOrd.Status = StatusFilled
		} else {
			sellOrd.Status = StatusPartial
		}

		if buyOrd.Filled >= buyOrd.Amount {
			buyOrd.Status = StatusFilled
		} else {
			buyOrd.Status = StatusPartial
		}

		e.idCounter++
		trade := &Trade{
			ID:          fmt.Sprintf("trd-%d", e.idCounter),
			BuyOrderID:  buyOrd.ID,
			SellOrderID: sellOrd.ID,
			Buyer:       buyOrd.User,
			Seller:      sellOrd.User,
			Price:       execPrice,
			Amount:      matchQty,
			Timestamp:   time.Now(),
		}

		trades = append(trades, trade)
		e.trades = append(e.trades, trade)
		e.lastPrice = execPrice

		if e.stabilityEng != nil {
			e.stabilityEng.SetMarketPrice(execPrice)
		}
	}

	return trades
}

func (e *engineImpl) CancelOrder(ctx context.Context, orderID string, user crypto.Address) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	ord, exists := e.orders[orderID]
	if !exists {
		return ErrOrderNotFound
	}
	if ord.User != user {
		return ErrUnauthorized
	}
	if ord.Status == StatusFilled || ord.Status == StatusCancelled {
		return ErrOrderCompleted
	}

	ord.Status = StatusCancelled
	return nil
}

func (e *engineImpl) GetDepth(ctx context.Context) (*Depth, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	bidMap := make(map[float64]token.Amount)
	askMap := make(map[float64]token.Amount)

	for _, ord := range e.orders {
		if ord.Status != StatusOpen && ord.Status != StatusPartial {
			continue
		}
		remaining := ord.Amount - ord.Filled
		if remaining == 0 {
			continue
		}
		if ord.Side == SideBuy {
			bidMap[ord.Price] += remaining
		} else {
			askMap[ord.Price] += remaining
		}
	}

	bids := make([]PriceLevel, 0, len(bidMap))
	for price, qty := range bidMap {
		bids = append(bids, PriceLevel{Price: price, Quantity: qty})
	}
	sort.Slice(bids, func(i, j int) bool {
		return bids[i].Price > bids[j].Price
	})

	asks := make([]PriceLevel, 0, len(askMap))
	for price, qty := range askMap {
		asks = append(asks, PriceLevel{Price: price, Quantity: qty})
	}
	sort.Slice(asks, func(i, j int) bool {
		return asks[i].Price < asks[j].Price
	})

	return &Depth{Bids: bids, Asks: asks}, nil
}

func (e *engineImpl) GetTrades(ctx context.Context, limit int) ([]*Trade, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if limit <= 0 || limit > len(e.trades) {
		limit = len(e.trades)
	}

	start := len(e.trades) - limit
	result := make([]*Trade, limit)
	copy(result, e.trades[start:])

	// Return most recent trade first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result, nil
}

func (e *engineImpl) GetUserOrders(ctx context.Context, user crypto.Address) ([]*Order, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	ords := e.userOrders[user]
	result := make([]*Order, len(ords))
	copy(result, ords)
	return result, nil
}

func (e *engineImpl) GetMarketPrice(ctx context.Context) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.lastPrice
}
