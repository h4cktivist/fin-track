package http

import (
	"context"
	"encoding/json"
	"errors"
	"fin-analytics/internal/domain"
	stdhttp "net/http"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"fin-analytics/internal/swagger"
)

type AnalyticsService interface {
	ProcessKafkaMessage(ctx context.Context, msg *sarama.ConsumerMessage) error
	GetStats(ctx context.Context, userID int) (domain.FinanceStats, error)
}

type Server struct {
	service AnalyticsService
	router  *chi.Mux
	server  *stdhttp.Server
}

func NewServer(service AnalyticsService) *Server {
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
		writeJSON(w, stdhttp.StatusOK, map[string]string{"status": "ok"})
	})

	s.router.Route("/swagger", func(r chi.Router) {
		r.Get("/", swagger.UIHandler("/swagger/spec"))
		r.Get("/spec", swagger.SpecHandler(swagger.FinAnalyticsSpec()))
	})

	s.router.Get("/v1/users/{userID}/stats", s.handleGetStats)
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

func (s *Server) handleGetStats(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		writeJSON(w, stdhttp.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	stats, err := s.service.GetStats(r.Context(), userID)
	if err != nil {
		writeJSON(w, stdhttp.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, stdhttp.StatusOK, stats)
}

func writeJSON(w stdhttp.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
