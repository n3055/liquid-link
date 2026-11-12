package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/orderbook"
	"github.com/n3055/liquid-link/internal/pcn"
	"github.com/n3055/liquid-link/internal/receipt"
	"github.com/n3055/liquid-link/internal/stability"
	"github.com/n3055/liquid-link/internal/token"
)

// [Dashboard] ActivityItem encapsulates unified event log items for the Activity UI feed.
type ActivityItem struct {
	Type      string `json:"type"`
	Title     string `json:"title"`
	Subtitle  string `json:"subtitle"`
	Detail    string `json:"detail"`
	Timestamp string `json:"timestamp"`
}

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
	ActivityLog      []*ActivityItem
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

// [Dashboard] collectDashboardData gathers live system statistics across all node sub-services,
// scoped to the target user wallet if provided.
func (s *Server) collectDashboardData(ctx context.Context, targetUser ...string) *dashboardData {
	d := &dashboardData{
		NodeID: s.cfg.ID,
		Port:   s.cfg.Port,
	}

	activeUser := ""
	if len(targetUser) > 0 && targetUser[0] != "" {
		activeUser = targetUser[0]
	} else if s.wallet != nil {
		activeUser = string(s.wallet.Address())
	}
	d.Address = activeUser

	userCryptoAddr := crypto.Address(activeUser)

	if s.assetMgr != nil {
		if activeUser != "" {
			if b, err := s.assetMgr.BalanceOf(ctx, userCryptoAddr); err == nil {
				d.Balance = b
			}
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
			remoteCount := 0
			for _, p := range peers {
				if p.ID != s.cfg.ID {
					remoteCount++
				}
			}
			d.PeerCount = remoteCount
		}
	}

	if s.pcnRegistry != nil {
		if chs, err := s.pcnRegistry.ListChannels(ctx); err == nil {
			if activeUser != "" {
				var filtered []*pcn.Channel
				for _, ch := range chs {
					if ch != nil && (string(ch.Share.PartyA) == activeUser || string(ch.Share.PartyB) == activeUser) {
						filtered = append(filtered, ch)
					}
				}
				d.Channels = filtered
			} else {
				d.Channels = chs
			}
		}
	}

	if s.receiptRegistry != nil {
		if mappings, err := s.receiptRegistry.ListUTRMapping(ctx); err == nil {
			if activeUser != "" {
				var filtered []*receipt.UTRMapping
				for _, m := range mappings {
					if m != nil && string(m.Address) == activeUser {
						filtered = append(filtered, m)
					}
				}
				d.UTRMappings = filtered
			} else {
				d.UTRMappings = mappings
			}
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

	if activeUser != "" {
		if acc, err := fiatM.GetAccount(ctx, userCryptoAddr); err == nil && acc != nil {
			d.INRBalance = acc.Available
			d.INRReserved = acc.Reserved
		}
		if stPos, err := stab.GetStakingPosition(ctx, userCryptoAddr); err == nil && stPos != nil {
			d.StakedShares = stPos.StakedShares
		}
		if bdPos, err := stab.GetBondPosition(ctx, userCryptoAddr); err == nil && bdPos != nil {
			d.BondsHeld = bdPos.BondsHeld
		}
	}

	// [Fiat Bridge] Populate Unified Activity Feed combining UTRs, Channels, and Blocks
	for _, m := range d.UTRMappings {
		d.ActivityLog = append(d.ActivityLog, &ActivityItem{
			Type:      "bridge",
			Title:     "Fiat Bridge " + m.UTR,
			Subtitle:  "Tx " + m.TransactionID,
			Detail:    m.Status,
			Timestamp: m.CreatedAt.Format("02 Jan 15:04"),
		})
	}
	for _, ch := range d.Channels {
		d.ActivityLog = append(d.ActivityLog, &ActivityItem{
			Type:      "channel",
			Title:     "State Channel " + ch.ID,
			Subtitle:  fmt.Sprintf("Seq %d", ch.Sequence),
			Detail:    string(ch.State),
			Timestamp: "Active",
		})
	}

	return d
}

// [Dashboard] handleDashboard processes HTTP requests to render the Web UI control panel.
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	userAddr := r.URL.Query().Get("address")
	d := s.collectDashboardData(r.Context(), userAddr)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := renderDashboard(w, d); err != nil {
		respondError(w, http.StatusInternalServerError, "render error")
	}
}
