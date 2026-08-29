package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/n3055/liquid-link/internal/blockchain"
	"github.com/n3055/liquid-link/internal/config"
	"github.com/n3055/liquid-link/internal/consensus"
	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/ledger"
	"github.com/n3055/liquid-link/internal/pcn"
	"github.com/n3055/liquid-link/internal/peer"
	"github.com/n3055/liquid-link/internal/receipt"
	"github.com/n3055/liquid-link/internal/token"
	"github.com/n3055/liquid-link/internal/wallet"
)

// Server implements HTTP routing and API handlers.
type Server struct {
	cfg             *config.NodeConfig
	router          chi.Router
	ledger          ledger.Ledger
	blockchain      blockchain.Blockchain
	consensus       consensus.Engine
	pcn             pcn.Manager
	bridge          receipt.Bridge
	wallet          wallet.Wallet
	peers           peer.Registry
	assetMgr        token.AssetManager
	pcnRegistry     pcn.HolderRegistry
	receiptRegistry receipt.Registry
}

// NewServer creates a new API Server instance with registered routes and injected dependencies.
func NewServer(
	cfg *config.NodeConfig,
	ledger ledger.Ledger,
	bc blockchain.Blockchain,
	cons consensus.Engine,
	pcnMgr pcn.Manager,
	bridge receipt.Bridge,
	w wallet.Wallet,
	peerReg peer.Registry,
	assetMgr token.AssetManager,
	pcnRegistry pcn.HolderRegistry,
	receiptReg receipt.Registry,
) *Server {
	r := chi.NewRouter()

	s := &Server{
		cfg:             cfg,
		router:          r,
		ledger:          ledger,
		blockchain:      bc,
		consensus:       cons,
		pcn:             pcnMgr,
		bridge:          bridge,
		wallet:          w,
		peers:           peerReg,
		assetMgr:        assetMgr,
		pcnRegistry:     pcnRegistry,
		receiptRegistry: receiptReg,
	}

	s.setupRoutes()

	return s
}

// Handler returns the HTTP handler for the server.
func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) setupRoutes() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	// Web UI Dashboard
	s.router.Get("/", s.handleDashboard)
	s.router.Get("/ui", s.handleDashboard)

	// API v1 Namespace
	s.router.Route("/api/v1", func(r chi.Router) {
		r.Get("/status", s.handleStatus)

		// Wallet actions
		r.Route("/wallet", func(wr chi.Router) {
			wr.Get("/info", s.handleWalletInfo)
			wr.Get("/balance", s.handleBalance)
			wr.Post("/transfer", s.handleTransfer)
			wr.Post("/mint", s.handleMint)
		})

		// PCN / Channels actions
		r.Route("/channels", func(cr chi.Router) {
			cr.Get("/list", s.handleListChannels)
			cr.Get("/{id}", s.handleGetChannel)
			cr.Post("/open", s.handleOpenChannel)
			cr.Post("/update", s.handleUpdateChannel)
			cr.Post("/settle", s.handleSettleChannel)
		})

		// Bridge / Settlement actions
		r.Route("/bridge", func(br chi.Router) {
			br.Post("/fiat-request", s.handleBridgeRequest)
			br.Post("/verify-receipt", s.handleVerifyReceipt)
			br.Get("/utr/{utr}", s.handleGetUTRMapping)
		})

		// Blockchain & Consensus
		r.Route("/blockchain", func(bcr chi.Router) {
			bcr.Get("/height", s.handleBlockchainHeight)
			bcr.Get("/blocks", s.handleListBlocks)
			bcr.Get("/block/{query}", s.handleGetBlock)
		})

		// Peer Network Topology
		r.Route("/peers", func(pr chi.Router) {
			pr.Get("/", s.handleListPeers)
			pr.Post("/register", s.handleRegisterPeer)
		})
	})
}

// Helpers
func respondJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, statusCode int, message string) {
	respondJSON(w, statusCode, map[string]string{"error": message})
}

// Handler implementations

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	height := uint64(0)
	if s.blockchain != nil {
		h, err := s.blockchain.Height(ctx)
		if err == nil {
			height = h
		}
	}

	supply := token.Amount(0)
	if s.assetMgr != nil {
		sup, err := s.assetMgr.TotalSupply(ctx)
		if err == nil {
			supply = sup
		}
	}

	peerCount := 0
	if s.peers != nil {
		pList, err := s.peers.ListPeers(ctx)
		if err == nil {
			peerCount = len(pList)
		}
	}

	bal := token.Amount(0)
	addr := ""
	if s.wallet != nil {
		addr = string(s.wallet.Address())
		if s.assetMgr != nil {
			b, err := s.assetMgr.BalanceOf(ctx, s.wallet.Address())
			if err == nil {
				bal = b
			}
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":       "online",
		"node_id":      s.cfg.ID,
		"port":         s.cfg.Port,
		"address":      addr,
		"balance":      bal,
		"total_supply": supply,
		"block_height": height,
		"peer_count":   peerCount,
	})
}

func (s *Server) handleWalletInfo(w http.ResponseWriter, r *http.Request) {
	if s.wallet == nil {
		respondError(w, http.StatusInternalServerError, "wallet not configured")
		return
	}

	ctx := r.Context()
	bal := token.Amount(0)
	if s.assetMgr != nil {
		b, err := s.assetMgr.BalanceOf(ctx, s.wallet.Address())
		if err == nil {
			bal = b
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"address": string(s.wallet.Address()),
		"balance": bal,
		"node_id": s.cfg.ID,
	})
}

func (s *Server) handleBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	targetAddr := r.URL.Query().Get("address")
	if targetAddr == "" {
		if s.wallet == nil {
			respondError(w, http.StatusBadRequest, "missing address query parameter")
			return
		}
		targetAddr = string(s.wallet.Address())
	}

	bal := token.Amount(0)
	if s.assetMgr != nil {
		b, err := s.assetMgr.BalanceOf(ctx, crypto.Address(targetAddr))
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		bal = b
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"address": targetAddr,
		"balance": bal,
	})
}

type transferReq struct {
	Sender   string       `json:"sender"`
	Receiver string       `json:"receiver"`
	Amount   token.Amount `json:"amount"`
}

func (s *Server) handleTransfer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req transferReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Receiver == "" || req.Amount == 0 {
		respondError(w, http.StatusBadRequest, "receiver and positive amount required")
		return
	}

	if s.assetMgr == nil || s.wallet == nil {
		respondError(w, http.StatusInternalServerError, "asset manager or wallet uninitialized")
		return
	}

	sender := s.wallet.Address()
	if req.Sender != "" {
		sender = crypto.Address(req.Sender)
	}
	if err := s.assetMgr.Debit(ctx, sender, req.Amount); err != nil {
		respondError(w, http.StatusBadRequest, "transfer failed: "+err.Error())
		return
	}

	if err := s.assetMgr.Credit(ctx, crypto.Address(req.Receiver), req.Amount); err != nil {
		// Rollback debit on failure
		_ = s.assetMgr.Credit(ctx, sender, req.Amount)
		respondError(w, http.StatusInternalServerError, "transfer credit failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"status":   "success",
		"sender":   string(sender),
		"receiver": req.Receiver,
		"amount":   req.Amount,
	})
}

type mintReq struct {
	Address string       `json:"address"`
	Amount  token.Amount `json:"amount"`
}

func (s *Server) handleMint(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req mintReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	targetAddr := req.Address
	if targetAddr == "" && s.wallet != nil {
		targetAddr = string(s.wallet.Address())
	}

	if req.Amount == 0 {
		respondError(w, http.StatusBadRequest, "amount must be greater than zero")
		return
	}

	if s.assetMgr == nil {
		respondError(w, http.StatusInternalServerError, "asset manager uninitialized")
		return
	}

	if err := s.assetMgr.Mint(ctx, crypto.Address(targetAddr), req.Amount); err != nil {
		respondError(w, http.StatusInternalServerError, "minting failed: "+err.Error())
		return
	}

	newBal, _ := s.assetMgr.BalanceOf(ctx, crypto.Address(targetAddr))

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":      "minted",
		"address":     targetAddr,
		"amount":      req.Amount,
		"new_balance": newBal,
	})
}

func (s *Server) handleListChannels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if s.pcnRegistry == nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{"channels": []interface{}{}})
		return
	}

	channels, err := s.pcnRegistry.ListChannels(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if channels == nil {
		channels = []*pcn.Channel{}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"channels": channels,
	})
}

func (s *Server) handleGetChannel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	channelID := chi.URLParam(r, "id")
	if channelID == "" {
		respondError(w, http.StatusBadRequest, "missing channel id parameter")
		return
	}

	if s.pcnRegistry == nil {
		respondError(w, http.StatusNotFound, "pcn registry uninitialized")
		return
	}

	ch, err := s.pcnRegistry.GetChannel(ctx, channelID)
	if err != nil {
		if errors.Is(err, pcn.ErrChannelNotFound) {
			respondError(w, http.StatusNotFound, "channel not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, ch)
}

type openChannelReq struct {
	Counterparty string       `json:"counterparty"`
	Amount       token.Amount `json:"amount"`
}

func (s *Server) handleOpenChannel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req openChannelReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Counterparty == "" || req.Amount == 0 {
		respondError(w, http.StatusBadRequest, "counterparty and deposit amount required")
		return
	}

	if s.pcn == nil {
		respondError(w, http.StatusInternalServerError, "pcn manager uninitialized")
		return
	}

	ch, err := s.pcn.OpenChannel(ctx, crypto.Address(req.Counterparty), req.Amount)
	if err != nil {
		respondError(w, http.StatusBadRequest, "failed to open channel: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, ch)
}

type updateChannelReq struct {
	ChannelID string       `json:"channel_id"`
	Amount    token.Amount `json:"amount"`
	Sending   bool         `json:"sending"`
}

func (s *Server) handleUpdateChannel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req updateChannelReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ChannelID == "" || req.Amount == 0 {
		respondError(w, http.StatusBadRequest, "channel_id and non-zero amount required")
		return
	}

	if s.pcn == nil {
		respondError(w, http.StatusInternalServerError, "pcn manager uninitialized")
		return
	}

	ch, err := s.pcn.UpdateChannelState(ctx, req.ChannelID, req.Amount, req.Sending)
	if err != nil {
		respondError(w, http.StatusBadRequest, "failed to update channel: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, ch)
}

type settleChannelReq struct {
	ChannelID string `json:"channel_id"`
	Action    string `json:"action"` // "initiate" or "finalize"
}

func (s *Server) handleSettleChannel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req settleChannelReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ChannelID == "" {
		respondError(w, http.StatusBadRequest, "channel_id required")
		return
	}

	if s.pcn == nil {
		respondError(w, http.StatusInternalServerError, "pcn manager uninitialized")
		return
	}

	if req.Action == "finalize" {
		if err := s.pcn.FinalizeSettlement(ctx, req.ChannelID); err != nil {
			respondError(w, http.StatusBadRequest, "failed to finalize channel settlement: "+err.Error())
			return
		}
	} else {
		if err := s.pcn.InitiateSettlement(ctx, req.ChannelID); err != nil {
			respondError(w, http.StatusBadRequest, "failed to initiate channel settlement: "+err.Error())
			return
		}
	}

	ch, _ := s.pcnRegistry.GetChannel(ctx, req.ChannelID)
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "settlement_updated",
		"channel": ch,
	})
}

type bridgeReq struct {
	Receiver string       `json:"receiver"`
	Amount   token.Amount `json:"amount"`
}

func (s *Server) handleBridgeRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req bridgeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Receiver == "" || req.Amount == 0 {
		respondError(w, http.StatusBadRequest, "receiver address and positive amount required")
		return
	}

	if s.bridge == nil {
		respondError(w, http.StatusInternalServerError, "bridge service uninitialized")
		return
	}

	mapping, err := s.bridge.InitiateBridgeRequest(ctx, crypto.Address(req.Receiver), req.Amount)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bridge request failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"status":  "bridge_request_created",
		"mapping": mapping,
	})
}

type verifyReceiptReq struct {
	UTR       string `json:"utr"`
	ReceiptID string `json:"receipt_id"`
}

func (s *Server) handleVerifyReceipt(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req verifyReceiptReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UTR == "" {
		respondError(w, http.StatusBadRequest, "utr required")
		return
	}

	if s.bridge == nil {
		respondError(w, http.StatusInternalServerError, "bridge service uninitialized")
		return
	}

	rec := &receipt.Receipt{
		ID:  req.ReceiptID,
		UTR: req.UTR,
	}
	if rec.ID == "" {
		rec.ID = "rec-" + req.UTR
	}

	if err := s.bridge.SettleBridgeRequest(ctx, req.UTR, rec); err != nil {
		respondError(w, http.StatusBadRequest, "receipt verification failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "receipt_verified",
		"utr":    req.UTR,
	})
}

func (s *Server) handleGetUTRMapping(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	utr := chi.URLParam(r, "utr")
	if utr == "" {
		respondError(w, http.StatusBadRequest, "missing utr parameter")
		return
	}

	if s.receiptRegistry == nil {
		respondError(w, http.StatusInternalServerError, "receipt registry uninitialized")
		return
	}

	mapping, err := s.receiptRegistry.GetUTRMapping(ctx, utr)
	if err != nil {
		respondError(w, http.StatusNotFound, "utr mapping not found")
		return
	}

	respondJSON(w, http.StatusOK, mapping)
}

func (s *Server) handleBlockchainHeight(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if s.blockchain == nil {
		respondError(w, http.StatusInternalServerError, "blockchain uninitialized")
		return
	}

	h, err := s.blockchain.Height(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tip, _ := s.blockchain.CurrentBlock(ctx)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"height": h,
		"tip":    tip,
	})
}

func (s *Server) handleListBlocks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if s.blockchain == nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{"blocks": []interface{}{}})
		return
	}

	h, err := s.blockchain.Height(ctx)
	if err != nil || h == 0 {
		respondJSON(w, http.StatusOK, map[string]interface{}{"blocks": []interface{}{}})
		return
	}

	var blocks []interface{}
	for i := int64(h) - 1; i >= 0 && len(blocks) < 20; i-- {
		blk, err := s.blockchain.GetBlockByHeight(ctx, uint64(i))
		if err == nil && blk != nil {
			blocks = append(blocks, blk)
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"height": h,
		"blocks": blocks,
	})
}

func (s *Server) handleGetBlock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := chi.URLParam(r, "query")
	if query == "" {
		respondError(w, http.StatusBadRequest, "query height or hash parameter required")
		return
	}

	if s.blockchain == nil {
		respondError(w, http.StatusInternalServerError, "blockchain uninitialized")
		return
	}

	if height, err := strconv.ParseUint(query, 10, 64); err == nil {
		blk, err := s.blockchain.GetBlockByHeight(ctx, height)
		if err != nil {
			respondError(w, http.StatusNotFound, "block height not found")
			return
		}
		respondJSON(w, http.StatusOK, blk)
		return
	}

	blk, err := s.blockchain.GetBlockByHash(ctx, strings.TrimPrefix(query, "0x"))
	if err != nil {
		respondError(w, http.StatusNotFound, "block hash not found")
		return
	}
	respondJSON(w, http.StatusOK, blk)
}

func (s *Server) handleListPeers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if s.peers == nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{"peers": []interface{}{}})
		return
	}

	pList, err := s.peers.ListPeers(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if pList == nil {
		pList = []*peer.Info{}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"peers": pList,
	})
}

func (s *Server) handleRegisterPeer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var p peer.Info
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respondError(w, http.StatusBadRequest, "invalid peer info payload")
		return
	}

	if p.ID == "" || p.Endpoint == "" {
		respondError(w, http.StatusBadRequest, "peer id and endpoint required")
		return
	}

	if s.peers == nil {
		respondError(w, http.StatusInternalServerError, "peer registry uninitialized")
		return
	}

	p.Connected = true
	if err := s.peers.RegisterPeer(ctx, &p); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to register peer: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "peer_registered",
		"peer":   p,
	})
}
