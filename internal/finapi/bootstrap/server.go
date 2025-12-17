package bootstrap

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"fin-track-app/internal/config"
	grpcserver "fin-track-app/internal/finapi/grpc"
	httpserver "fin-track-app/internal/finapi/http"
)

func RunApp(
	ctx context.Context,
	cancel context.CancelFunc,
	httpServer *httpserver.Server,
	grpcServer *grpcserver.Server,
	cfg *config.Config,
) {
	httpAddr := fmt.Sprintf("%s:%d", cfg.FinAPI.HTTPHost, cfg.FinAPI.HTTPPort)
	grpcAddr := fmt.Sprintf("%s:%d", cfg.FinAPI.GRPCHost, cfg.FinAPI.GRPCPort)

	errCh := make(chan error, 2)

	go func() {
		log.Printf("fin-api HTTP listening on %s", httpAddr)
		errCh <- httpServer.Start(ctx, httpAddr)
	}()

	go func() {
		log.Printf("fin-api gRPC listening on %s", grpcAddr)
		errCh <- grpcServer.Start(grpcAddr)
	}()

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		cancel()
	}()

	select {
	case err := <-errCh:
		log.Fatalf("server error: %v", err)
	case <-ctx.Done():
		log.Println("shutting down gracefully...")
	}
}
