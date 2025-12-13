package http

import (
	"context"
	"encoding/json"
	"errors"
	stdhttp "net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"fin-track-app/internal/domain"
	"fin-track-app/internal/swagger"
)

type TransactionService interface {
	CreateTransaction(ctx context.Context, tx domain.Transaction) (domain.Transaction, error)
	ListTransactions(ctx context.Context, userID string) ([]domain.Transaction, error)
	UpdateTransaction(ctx context.Context, tx domain.Transaction) (domain.Transaction, error)
	DeleteTransaction(ctx context.Context, userID string, transactionID int64) error
}

type Server struct {
	service TransactionService
	router  *chi.Mux
	server  *stdhttp.Server
}

func NewServer(service TransactionService) *Server {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	s := &Server{
		service: service,
		router:  router,
	}

	s.routes()
	return s
}

func (s *Server) routes() {
	s.router.Get("/healthz", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	s.router.Route("/swagger", func(r chi.Router) {
		r.Get("/", swagger.UIHandler("/swagger/spec"))
		r.Get("/spec", swagger.SpecHandler(swagger.FinAPISpec()))
	})

	s.router.Route("/v1/users/{userID}", func(r chi.Router) {
		r.Post("/transactions", s.handleCreateTransaction)
		r.Get("/transactions", s.handleListTransactions)
		r.Put("/transactions/{transactionID}", s.handleUpdateTransaction)
		r.Patch("/transactions/{transactionID}", s.handleUpdateTransaction)
		r.Delete("/transactions/{transactionID}", s.handleDeleteTransaction)
	})
}

func (s *Server) Start(ctx context.Context, addr string) error {
	s.server = &stdhttp.Server{
		Addr:    addr,
		Handler: s.router,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		return s.Stop(context.Background())
	case err := <-errCh:
		if errors.Is(err, stdhttp.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func (s *Server) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

type transactionRequest struct {
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
	Type     string  `json:"type"`
}

func (s *Server) handleCreateTransaction(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	userID := chi.URLParam(r, "userID")
	var req transactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, stdhttp.StatusBadRequest, "invalid payload")
		return
	}

	txType := domain.TransactionType(req.Type)
	if txType != domain.TransactionTypeIncome && txType != domain.TransactionTypeExpense {
		httpError(w, stdhttp.StatusBadRequest, "type must be income or expense")
		return
	}

	tx := domain.Transaction{
		UserID:   userID,
		Amount:   req.Amount,
		Category: req.Category,
		Type:     txType,
	}

	created, err := s.service.CreateTransaction(r.Context(), tx)
	if err != nil {
		httpError(w, stdhttp.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, stdhttp.StatusCreated, created)
}

func (s *Server) handleListTransactions(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	userID := chi.URLParam(r, "userID")

	items, err := s.service.ListTransactions(r.Context(), userID)
	if err != nil {
		httpError(w, stdhttp.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, stdhttp.StatusOK, items)
}

func (s *Server) handleUpdateTransaction(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	userID := chi.URLParam(r, "userID")
	transactionID, err := parseTransactionID(r)
	if err != nil {
		httpError(w, stdhttp.StatusBadRequest, "invalid transaction id")
		return
	}

	var req transactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, stdhttp.StatusBadRequest, "invalid payload")
		return
	}

	txType := domain.TransactionType(req.Type)
	if txType != domain.TransactionTypeIncome && txType != domain.TransactionTypeExpense {
		httpError(w, stdhttp.StatusBadRequest, "type must be income or expense")
		return
	}

	updated, err := s.service.UpdateTransaction(r.Context(), domain.Transaction{
		ID:       transactionID,
		UserID:   userID,
		Amount:   req.Amount,
		Category: req.Category,
		Type:     txType,
	})
	if err != nil {
		if errors.Is(err, domain.ErrTransactionNotFound) {
			httpError(w, stdhttp.StatusNotFound, err.Error())
			return
		}
		httpError(w, stdhttp.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, stdhttp.StatusOK, updated)
}

func (s *Server) handleDeleteTransaction(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	userID := chi.URLParam(r, "userID")
	transactionID, err := parseTransactionID(r)
	if err != nil {
		httpError(w, stdhttp.StatusBadRequest, "invalid transaction id")
		return
	}

	if err := s.service.DeleteTransaction(r.Context(), userID, transactionID); err != nil {
		if errors.Is(err, domain.ErrTransactionNotFound) {
			httpError(w, stdhttp.StatusNotFound, err.Error())
			return
		}
		httpError(w, stdhttp.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(stdhttp.StatusNoContent)
}

func httpError(w stdhttp.ResponseWriter, code int, message string) {
	writeJSON(w, code, map[string]string{"error": message})
}

func writeJSON(w stdhttp.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func parseTransactionID(r *stdhttp.Request) (int64, error) {
	raw := chi.URLParam(r, "transactionID")
	return strconv.ParseInt(raw, 10, 64)
}
