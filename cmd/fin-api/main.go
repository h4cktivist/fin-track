package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"fin-track-app/internal/config"
	"fin-track-app/internal/database"
	finapigrpc "fin-track-app/internal/finapi/grpc"
	finapihttp "fin-track-app/internal/finapi/http"
	"fin-track-app/internal/finapi/repository"
	"fin-track-app/internal/finapi/service"
	"fin-track-app/internal/kafka"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.NewPostgresPool(ctx, cfg.Postgres)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer db.Close()

	producer, err := kafka.NewProducer(cfg.Kafka.Brokers, cfg.App.KafkaTopic)
	if err != nil {
		log.Fatalf("create kafka producer: %v", err)
	}
	defer producer.Close()

	repo := repository.NewTransactionRepository(db)
	svc := service.NewTransactionService(repo, producer)
	httpServer := finapihttp.NewServer(svc)
	grpcServer := finapigrpc.NewServer(svc)

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

	if err := <-errCh; err != nil {
		log.Fatalf("server error: %v", err)
	}
}
