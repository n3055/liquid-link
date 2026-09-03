package api

import (
	"context"
	"net/http"

	"github.com/n3055/liquid-link/internal/orderbook"
	"github.com/n3055/liquid-link/internal/pcn"
	"github.com/n3055/liquid-link/internal/receipt"
	"github.com/n3055/liquid-link/internal/stability"
	"github.com/n3055/liquid-link/internal/token"
)

// [Dashboard] dashboardData encapsulates all view-model metrics presented on the node web dashboard.
type dashboardData struct {
	NodeID           string
	Port             int
	Address          string
	Balance          token.Amount
	TotalSupply      token.Amount
	BlockHeight      uint64
	PeerCount        int
	Channels         []*pcn.Channel
	UTRMappings      []*receipt.UTRMapping
	RecentBlocks     []interface{}
	Tip              interface{}
	OrderbookDepth   *orderbook.Depth
	OrderbookTrades  []*orderbook.Trade
	StabilityStatus  *stability.Status
	StabilityHistory []stability.PricePoint
	StakedShares     token.Amount
	BondsHeld        token.Amount
	INRBalance       float64
	INRReserved      float64
	TotalINRPool     float64
}

// [Dashboard] collectDashboardData gathers live system statistics across all node sub-services.
func (s *Server) collectDashboardData(ctx context.Context) *dashboardData {
	d := &dashboardData{
		NodeID: s.cfg.ID,
		Port:   s.cfg.Port,
	}

	if s.wallet != nil {
		d.Address = string(s.wallet.Address())
	}

	if s.assetMgr != nil {
		if b, err := s.assetMgr.BalanceOf(ctx, s.wallet.Address()); err == nil {
			d.Balance = b
		}
		if t, err := s.assetMgr.TotalSupply(ctx); err == nil {
			d.TotalSupply = t
		}
	}

	if s.blockchain != nil {
		if h, err := s.blockchain.Height(ctx); err == nil {
			d.BlockHeight = h
		}
		if tip, err := s.blockchain.CurrentBlock(ctx); err == nil {
			d.Tip = tip
		}
		if d.BlockHeight > 0 {
			start := int64(d.BlockHeight) - 10
			if start < 0 {
				start = 0
			}
			for i := start; i < int64(d.BlockHeight); i++ {
				blk, err := s.blockchain.GetBlockByHeight(ctx, uint64(i))
				if err == nil && blk != nil {
					d.RecentBlocks = append(d.RecentBlocks, blk)
				}
			}
		}
	}

	if s.peers != nil {
		if peers, err := s.peers.ListPeers(ctx); err == nil {
			d.PeerCount = len(peers)
		}
	}

	if s.pcnRegistry != nil {
		if chs, err := s.pcnRegistry.ListChannels(ctx); err == nil {
			d.Channels = chs
		}
	}

	if s.receiptRegistry != nil {
		if mappings, err := s.receiptRegistry.ListUTRMapping(ctx); err == nil {
			d.UTRMappings = mappings
		}
	}

	ob := s.getOrderbookEngine()
	if depth, err := ob.GetDepth(ctx); err == nil {
		d.OrderbookDepth = depth
	}
	if trades, err := ob.GetTrades(ctx, 15); err == nil {
		d.OrderbookTrades = trades
	}

	stab := s.getStabilityEngine()
	if status, err := stab.GetStatus(ctx); err == nil {
		d.StabilityStatus = status
	}
	if history, err := stab.GetPriceHistory(ctx); err == nil {
		d.StabilityHistory = history
	}

	fiatM := s.getFiatManager()
	if poolRes, err := fiatM.GetTotalPoolReserve(ctx); err == nil {
		d.TotalINRPool = poolRes
	}

	if s.wallet != nil {
		userAddr := s.wallet.Address()
		if acc, err := fiatM.GetAccount(ctx, userAddr); err == nil && acc != nil {
			d.INRBalance = acc.Available
			d.INRReserved = acc.Reserved
		}
		if stPos, err := stab.GetStakingPosition(ctx, userAddr); err == nil && stPos != nil {
			d.StakedShares = stPos.StakedShares
		}
		if bdPos, err := stab.GetBondPosition(ctx, userAddr); err == nil && bdPos != nil {
			d.BondsHeld = bdPos.BondsHeld
		}
	}

	return d
}

// [Dashboard] handleDashboard processes HTTP requests to render the Web UI control panel.
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	d := s.collectDashboardData(r.Context())

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := renderDashboard(w, d); err != nil {
		respondError(w, http.StatusInternalServerError, "render error")
	}
}
