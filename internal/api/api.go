package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/n3055/liquid-link/internal/blockchain"
	"github.com/n3055/liquid-link/internal/config"
	"github.com/n3055/liquid-link/internal/consensus"
	"github.com/n3055/liquid-link/internal/crypto"
	"github.com/n3055/liquid-link/internal/ledger"
	"github.com/n3055/liquid-link/internal/orderbook"
	"github.com/n3055/liquid-link/internal/pcn"
	"github.com/n3055/liquid-link/internal/peer"
	"github.com/n3055/liquid-link/internal/receipt"
	"github.com/n3055/liquid-link/internal/stability"
	"github.com/n3055/liquid-link/internal/token"
	"github.com/n3055/liquid-link/internal/wallet"
)

// Server implements HTTP routing and API handlers.
type Server struct {
	mu              sync.Mutex
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
	stabilityEng    stability.Engine
	orderbookEng    orderbook.Engine
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
			wr.Get("/holdings", s.handleWalletHoldings)
			wr.Post("/create", s.handleWalletCreate)
			wr.Post("/transfer", s.handleTransfer)
			wr.Post("/mint", s.handleMint)
			wr.Post("/credit-sync", s.handleCreditSync)
		})

		// Transaction broadcast
		r.Route("/tx", func(tr chi.Router) {
			tr.Post("/broadcast", s.handleBroadcastTx)
		})

		// Orderbook actions
		r.Route("/orderbook", func(obr chi.Router) {
			obr.Get("/depth", s.handleOrderbookDepth)
			obr.Get("/trades", s.handleOrderbookTrades)
			obr.Get("/user-orders", s.handleUserOrders)
			obr.Post("/order", s.handlePlaceOrder)
			obr.Post("/cancel", s.handleCancelOrder)
		})

		// Seigniorage Stability actions
		r.Route("/stability", func(sr chi.Router) {
			sr.Get("/status", s.handleStabilityStatus)
			sr.Get("/history", s.handleStabilityHistory)
			sr.Post("/stake", s.handleStakeShares)
			sr.Post("/unstake", s.handleUnstakeShares)
			sr.Post("/buy-bonds", s.handleBuyBonds)
			sr.Post("/redeem-bonds", s.handleRedeemBonds)
		})

		// PCN / Channels actions
		r.Route("/channels", func(cr chi.Router) {
			cr.Get("/list", s.handleListChannels)
			cr.Get("/{id}", s.handleGetChannel)
			cr.Post("/open", s.handleOpenChannel)
			cr.Post("/update", s.handleUpdateChannel)
			cr.Post("/settle", s.handleSettleChannel)
			cr.Post("/challenge", s.handleChallengeChannel)
		})

		// Bridge / Settlement actions
		r.Route("/bridge", func(br chi.Router) {
			br.Get("/list", s.handleListUTRMappings)
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

	// Trigger asynchronous inter-node balance synchronization across peer network
	go s.syncPeerCredit(crypto.Address(req.Receiver), req.Amount)

	respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"status":   "success",
		"sender":   string(sender),
		"receiver": req.Receiver,
		"amount":   req.Amount,
	})
}

type broadcastTxReq struct {
	Sender    string       `json:"sender"`
	Receiver  string       `json:"receiver"`
	Amount    token.Amount `json:"amount"`
	Nonce     uint64       `json:"nonce"`
	PublicKey string       `json:"public_key"`
	Signature string       `json:"signature"`
}

// [API] handleBroadcastTx receives pre-signed client transaction payloads, validates cryptographic signatures, and executes transfers.
func (s *Server) handleBroadcastTx(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req broadcastTxReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid transaction payload")
		return
	}

	if req.Sender == "" || req.Receiver == "" || req.Amount == 0 {
		respondError(w, http.StatusBadRequest, "sender, receiver, and positive amount required")
		return
	}

	if s.assetMgr == nil {
		respondError(w, http.StatusInternalServerError, "asset manager uninitialized")
		return
	}

	// Verify cryptographic signature if provided
	if req.PublicKey != "" && req.Signature != "" {
		pubKeyBytes, err := hex.DecodeString(req.PublicKey)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid public key format")
			return
		}
		sigBytes, err := hex.DecodeString(req.Signature)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid signature format")
			return
		}

		payload := fmt.Sprintf("%s|%s|%d|%d", req.Sender, req.Receiver, req.Amount, req.Nonce)
		hash := sha256.Sum256([]byte(payload))

		verifier := crypto.NewVerifier()
		if !verifier.Verify(hash[:], sigBytes, crypto.PublicKey(pubKeyBytes)) {
			respondError(w, http.StatusBadRequest, "invalid transaction signature")
			return
		}
	}

	senderAddr := crypto.Address(req.Sender)
	if err := s.assetMgr.Debit(ctx, senderAddr, req.Amount); err != nil {
		respondError(w, http.StatusBadRequest, "transfer failed: "+err.Error())
		return
	}

	if err := s.assetMgr.Credit(ctx, crypto.Address(req.Receiver), req.Amount); err != nil {
		_ = s.assetMgr.Credit(ctx, senderAddr, req.Amount)
		respondError(w, http.StatusInternalServerError, "transfer credit failed: "+err.Error())
		return
	}

	txIDHash := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%d|%d", req.Sender, req.Receiver, req.Amount, req.Nonce)))
	txID := "tx-" + hex.EncodeToString(txIDHash[:8])

	go s.syncPeerCredit(crypto.Address(req.Receiver), req.Amount)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":         "broadcast_success",
		"transaction_id": txID,
		"sender":         req.Sender,
		"receiver":       req.Receiver,
		"amount":         req.Amount,
	})
}

// [Ledger] syncPeerCredit broadcasts a credit update for an address to all peer endpoints over HTTP REST.
func (s *Server) syncPeerCredit(addr crypto.Address, amount token.Amount) {
	defer func() {
		_ = recover()
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	endpoints := make(map[string]bool)

	if s.peers != nil {
		if peers, err := s.peers.ListPeers(ctx); err == nil {
			for _, p := range peers {
				if p.ID != s.cfg.ID && p.Endpoint != "" {
					endpoints[p.Endpoint] = true
				}
			}
		}
	}

	// Always fallback to standard local node ports if peer list is unpopulated
	defaultPorts := []int{8080, 8081}
	for _, port := range defaultPorts {
		if port != s.cfg.Port {
			endpoints[fmt.Sprintf("http://localhost:%d", port)] = true
		}
	}

	payload, err := json.Marshal(map[string]interface{}{
		"address": string(addr),
		"amount":  amount,
	})
	if err != nil {
		return
	}

	client := &http.Client{Timeout: 2 * time.Second}
	for ep := range endpoints {
		url := fmt.Sprintf("%s/api/v1/wallet/credit-sync", strings.TrimSuffix(ep, "/"))
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err == nil {
				_ = resp.Body.Close()
			}
		}
	}
}

func (s *Server) handleCreditSync(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req mintReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Address == "" || req.Amount == 0 {
		respondError(w, http.StatusBadRequest, "address and positive amount required")
		return
	}

	if s.assetMgr == nil {
		respondError(w, http.StatusInternalServerError, "asset manager uninitialized")
		return
	}

	if err := s.assetMgr.Credit(ctx, crypto.Address(req.Address), req.Amount); err != nil {
		respondError(w, http.StatusInternalServerError, "credit sync failed: "+err.Error())
		return
	}

	newBal, _ := s.assetMgr.BalanceOf(ctx, crypto.Address(req.Address))

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":      "credited",
		"address":     req.Address,
		"amount":      req.Amount,
		"new_balance": newBal,
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

type challengeChannelReq struct {
	ChannelID string `json:"channel_id"`
	Sequence  uint64 `json:"sequence"`
}

func (s *Server) handleChallengeChannel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req challengeChannelReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ChannelID == "" {
		respondError(w, http.StatusBadRequest, "channel_id required")
		return
	}

	if s.pcn == nil || s.pcnRegistry == nil {
		respondError(w, http.StatusInternalServerError, "pcn uninitialized")
		return
	}

	ch, err := s.pcnRegistry.GetChannel(ctx, req.ChannelID)
	if err != nil {
		respondError(w, http.StatusNotFound, "channel not found")
		return
	}

	higherSeq := *ch
	higherSeq.Sequence = req.Sequence
	if higherSeq.Sequence <= ch.Sequence {
		higherSeq.Sequence = ch.Sequence + 1
	}

	if err := s.pcn.ChallengeSettlement(ctx, req.ChannelID, &higherSeq); err != nil {
		respondError(w, http.StatusBadRequest, "challenge failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "dispute_challenged",
		"channel": higherSeq,
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

// [API] handleListUTRMappings returns all fiat bridge UTR mappings in the system.
func (s *Server) handleListUTRMappings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if s.receiptRegistry == nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{"mappings": []interface{}{}})
		return
	}

	mappings, err := s.receiptRegistry.ListUTRMapping(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list UTR mappings: "+err.Error())
		return
	}

	if mappings == nil {
		mappings = []*receipt.UTRMapping{}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"mappings": mappings,
	})
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

func (s *Server) getStabilityEngine() stability.Engine {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stabilityEng == nil {
		s.stabilityEng = stability.NewEngine(nil, s.assetMgr)
	}
	return s.stabilityEng
}

func (s *Server) getOrderbookEngine() orderbook.Engine {
	stab := s.getStabilityEngine()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.orderbookEng == nil {
		s.orderbookEng = orderbook.NewEngine(s.assetMgr, stab)
	}
	return s.orderbookEng
}

// [API] handleWalletCreate generates a new wallet keypair, initializes balance, and returns credentials.
func (s *Server) handleWalletCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	newW, err := wallet.GenerateWallet()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate wallet: "+err.Error())
		return
	}

	addr := newW.Address()

	// Credit initial test balance to new wallet (1000 LQD)
	if s.assetMgr != nil {
		_ = s.assetMgr.Mint(ctx, addr, 1000)
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"status":  "created",
		"address": string(addr),
		"balance": 1000,
	})
}

// [API] handleWalletHoldings returns complete token balance, staked share holdings, and bond holdings for an address.
func (s *Server) handleWalletHoldings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	targetAddr := r.URL.Query().Get("address")
	if targetAddr == "" && s.wallet != nil {
		targetAddr = string(s.wallet.Address())
	}
	if targetAddr == "" {
		respondError(w, http.StatusBadRequest, "address parameter required")
		return
	}

	userAddr := crypto.Address(targetAddr)
	var balance token.Amount
	if s.assetMgr != nil {
		if b, err := s.assetMgr.BalanceOf(ctx, userAddr); err == nil {
			balance = b
		}
	}

	stabEng := s.getStabilityEngine()
	stakePos, _ := stabEng.GetStakingPosition(ctx, userAddr)
	bondPos, _ := stabEng.GetBondPosition(ctx, userAddr)

	var stakedShares token.Amount
	var accruedYield token.Amount
	if stakePos != nil {
		stakedShares = stakePos.StakedShares
		accruedYield = stakePos.AccruedYield
	}

	var bondsHeld token.Amount
	if bondPos != nil {
		bondsHeld = bondPos.BondsHeld
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"address":       targetAddr,
		"balance":       balance,
		"staked_shares": stakedShares,
		"accrued_yield": accruedYield,
		"bonds_held":    bondsHeld,
	})
}

// [Orderbook] handleOrderbookDepth returns the current orderbook depth (bids and asks).
func (s *Server) handleOrderbookDepth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	depth, err := s.getOrderbookEngine().GetDepth(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch orderbook depth: "+err.Error())
		return
	}
	respondJSON(w, http.StatusOK, depth)
}

// [Orderbook] handleOrderbookTrades returns recent executed trades.
func (s *Server) handleOrderbookTrades(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	trades, err := s.getOrderbookEngine().GetTrades(ctx, 30)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch trades: "+err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"trades": trades,
	})
}

// [Orderbook] handleUserOrders returns all orders for a specific user address.
func (s *Server) handleUserOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	addrStr := r.URL.Query().Get("address")
	if addrStr == "" && s.wallet != nil {
		addrStr = string(s.wallet.Address())
	}
	if addrStr == "" {
		respondError(w, http.StatusBadRequest, "address parameter required")
		return
	}

	orders, err := s.getOrderbookEngine().GetUserOrders(ctx, crypto.Address(addrStr))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch user orders: "+err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"orders": orders,
	})
}

type placeOrderReq struct {
	User   string       `json:"user"`
	Side   string       `json:"side"`
	Price  float64      `json:"price"`
	Amount token.Amount `json:"amount"`
}

// [Orderbook] handlePlaceOrder submits a new limit buy or sell order to the orderbook.
func (s *Server) handlePlaceOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req placeOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid order payload")
		return
	}

	userAddr := req.User
	if userAddr == "" && s.wallet != nil {
		userAddr = string(s.wallet.Address())
	}
	if userAddr == "" {
		respondError(w, http.StatusBadRequest, "user address required")
		return
	}

	ord := &orderbook.Order{
		User:   crypto.Address(userAddr),
		Side:   orderbook.OrderSide(strings.ToLower(req.Side)),
		Price:  req.Price,
		Amount: req.Amount,
	}

	placed, trades, err := s.getOrderbookEngine().PlaceOrder(ctx, ord)
	if err != nil {
		respondError(w, http.StatusBadRequest, "order placement failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"order":  placed,
		"trades": trades,
	})
}

type cancelOrderReq struct {
	OrderID string `json:"order_id"`
	User    string `json:"user"`
}

// [Orderbook] handleCancelOrder cancels an open order.
func (s *Server) handleCancelOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req cancelOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid cancel payload")
		return
	}

	userAddr := req.User
	if userAddr == "" && s.wallet != nil {
		userAddr = string(s.wallet.Address())
	}

	if err := s.getOrderbookEngine().CancelOrder(ctx, req.OrderID, crypto.Address(userAddr)); err != nil {
		respondError(w, http.StatusBadRequest, "failed to cancel order: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "cancelled",
		"order_id": req.OrderID,
	})
}

// [Stability] handleStabilityStatus returns the current stability status and seigniorage parameters.
func (s *Server) handleStabilityStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status, err := s.getStabilityEngine().GetStatus(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get stability status: "+err.Error())
		return
	}
	respondJSON(w, http.StatusOK, status)
}

// [Stability] handleStabilityHistory returns price history points for charting.
func (s *Server) handleStabilityHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	history, err := s.getStabilityEngine().GetPriceHistory(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get price history: "+err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"history": history,
	})
}

type stakeReq struct {
	User   string       `json:"user"`
	Amount token.Amount `json:"amount"`
}

// [Stability] handleStakeShares stakes LQD tokens into seigniorage shares.
func (s *Server) handleStakeShares(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req stakeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid stake payload")
		return
	}

	userAddr := req.User
	if userAddr == "" && s.wallet != nil {
		userAddr = string(s.wallet.Address())
	}

	pos, err := s.getStabilityEngine().StakeShares(ctx, crypto.Address(userAddr), req.Amount)
	if err != nil {
		respondError(w, http.StatusBadRequest, "stake failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, pos)
}

// [Stability] handleUnstakeShares unstakes shares and claims yield.
func (s *Server) handleUnstakeShares(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req stakeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid unstake payload")
		return
	}

	userAddr := req.User
	if userAddr == "" && s.wallet != nil {
		userAddr = string(s.wallet.Address())
	}

	pos, err := s.getStabilityEngine().UnstakeShares(ctx, crypto.Address(userAddr), req.Amount)
	if err != nil {
		respondError(w, http.StatusBadRequest, "unstake failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, pos)
}

type bondReq struct {
	User   string       `json:"user"`
	Amount token.Amount `json:"amount"`
}

// [Stability] handleBuyBonds purchases discounted LQD bonds during contraction.
func (s *Server) handleBuyBonds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req bondReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid bond payload")
		return
	}

	userAddr := req.User
	if userAddr == "" && s.wallet != nil {
		userAddr = string(s.wallet.Address())
	}

	pos, err := s.getStabilityEngine().BuyBonds(ctx, crypto.Address(userAddr), req.Amount)
	if err != nil {
		respondError(w, http.StatusBadRequest, "buy bonds failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, pos)
}

// [Stability] handleRedeemBonds redeems LQD bonds for LQD + bonus during expansion.
func (s *Server) handleRedeemBonds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req bondReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid redemption payload")
		return
	}

	userAddr := req.User
	if userAddr == "" && s.wallet != nil {
		userAddr = string(s.wallet.Address())
	}

	pos, err := s.getStabilityEngine().RedeemBonds(ctx, crypto.Address(userAddr), req.Amount)
	if err != nil {
		respondError(w, http.StatusBadRequest, "redeem bonds failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, pos)
}
