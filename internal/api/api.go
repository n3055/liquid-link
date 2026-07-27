package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/n3055/liquid-link/internal/config"
	"github.com/n3055/liquid-link/internal/ledger"
	"github.com/n3055/liquid-link/internal/pcn"
	"github.com/n3055/liquid-link/internal/receipt"
	"github.com/n3055/liquid-link/internal/wallet"
)

// Server implements HTTP routing and handlers.
type Server struct {
	cfg    *config.NodeConfig
	router chi.Router
	ledger ledger.Ledger
	pcn    pcn.Manager
	bridge receipt.Bridge
	wallet wallet.Wallet
}

// NewServer creates a new API Server instance with registered routes.
func NewServer(
	cfg *config.NodeConfig,
	ledger ledger.Ledger,
	pcnMgr pcn.Manager,
	bridge receipt.Bridge,
	w wallet.Wallet,
) *Server {
	r := chi.NewRouter()

	s := &Server{
		cfg:    cfg,
		router: r,
		ledger: ledger,
		pcn:    pcnMgr,
		bridge: bridge,
		wallet: w,
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

	// API v1 Namespace
	s.router.Route("/api/v1", func(r chi.Router) {
		r.Get("/status", s.handleStatus)

		// Wallet actions
		r.Route("/wallet", func(wr chi.Router) {
			wr.Get("/balance", s.handleBalance)
			wr.Post("/transfer", s.handleTransfer)
		})

		// PCN / Channels actions
		r.Route("/channels", func(cr chi.Router) {
			cr.Post("/open", s.handleOpenChannel)
			cr.Post("/update", s.handleUpdateChannel)
			cr.Post("/settle", s.handleSettleChannel)
		})

		// Bridge / Settlement actions
		r.Route("/bridge", func(br chi.Router) {
			br.Post("/fiat-request", s.handleBridgeRequest)
			br.Post("/verify-receipt", s.handleVerifyReceipt)
		})
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"online","node_id":"` + s.cfg.ID + `"}`))
}

func (s *Server) handleBalance(w http.ResponseWriter, r *http.Request) {
	// TODO: Retrieve balance from ledger
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"balance": 10000}`))
}

func (s *Server) handleTransfer(w http.ResponseWriter, r *http.Request) {
	// TODO: Create and submit signed transaction
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"status":"transaction_pending"}`))
}

func (s *Server) handleOpenChannel(w http.ResponseWriter, r *http.Request) {
	// TODO: Open payment channel
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(`{"status":"channel_opening"}`))
}

func (s *Server) handleUpdateChannel(w http.ResponseWriter, r *http.Request) {
	// TODO: Co-sign and update state
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"channel_updated"}`))
}

func (s *Server) handleSettleChannel(w http.ResponseWriter, r *http.Request) {
	// TODO: Initiate settle process
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"settlement_initiated"}`))
}

func (s *Server) handleBridgeRequest(w http.ResponseWriter, r *http.Request) {
	// TODO: Bridge crypto to fiat
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"status":"bridge_request_created"}`))
}

func (s *Server) handleVerifyReceipt(w http.ResponseWriter, r *http.Request) {
	// TODO: Verify receipt from fiat gateway
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"receipt_verified"}`))
}
