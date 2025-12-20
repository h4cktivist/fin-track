package bootstrap

import (
	"context"
	"fin-track-app/fin-api/internal/config"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	finanalyticshttp "fin-track-app/internal/finanalytics/http"
	"fin-track-app/internal/kafka"

	"github.com/IBM/sarama"
)

func RunAnalyticsApp(
	ctx context.Context,
	cancel context.CancelFunc,
	httpServer *finanalyticshttp.Server,
	kafkaConsumer *kafka.Consumer,
	cfg *config.Config,
) {
	httpAddr := fmt.Sprintf("%s:%d", cfg.FinAnalytics.HTTPHost, cfg.FinAnalytics.HTTPPort)

	errCh := make(chan error, 2)

	go func() {
		log.Printf("fin-analytics HTTP listening on %s", httpAddr)
		errCh <- httpServer.Start(ctx, httpAddr)
	}()

	go func() {
		log.Println("fin-analytics consumer started")
		errCh <- kafkaConsumer.Start(ctx)
	}()

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Println("shutdown signal received")
		cancel()
	}()

	select {
	case err := <-errCh:
		if err != context.Canceled && err != sarama.ErrClosedConsumerGroup {
			log.Fatalf("service error: %v", err)
		}
		log.Printf("server stopped: %v", err)
	case <-ctx.Done():
		log.Println("context cancelled, shutting down...")
	}
}
