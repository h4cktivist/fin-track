package http

import (
	"context"
	"errors"
	stdhttp "net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"fin-track-app/internal/domain"
	"fin-track-app/internal/swagger"
)

type TransactionService interface {
	CreateTransaction(ctx context.Context, tx domain.Transaction) (domain.Transaction, error)
	ListTransactions(ctx context.Context, userID int) ([]domain.Transaction, error)
	UpdateTransaction(ctx context.Context, tx domain.Transaction) (domain.Transaction, error)
	DeleteTransaction(ctx context.Context, userID int, transactionID int64) error
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
