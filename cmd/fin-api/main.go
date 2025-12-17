package main

import (
	"context"
	"log"

	"fin-track-app/internal/config"
	"fin-track-app/internal/database"
	"fin-track-app/internal/finapi/bootstrap"
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

	shardManager, err := database.NewIntShardManager(ctx, cfg.Postgres)
	if err != nil {
		log.Fatalf("create shard manager: %v", err)
	}
	defer shardManager.Close()

	producer, err := kafka.NewProducer(cfg.Kafka.Brokers, cfg.App.KafkaTopic)
	if err != nil {
		log.Fatalf("create kafka producer: %v", err)
	}
	defer producer.Close()

	repo := repository.NewTransactionRepository(shardManager)
	svc := service.NewTransactionService(repo, producer)
	httpServer := finapihttp.NewServer(svc)
	grpcServer := finapigrpc.NewServer(svc)

	bootstrap.RunApp(ctx, cancel, httpServer, grpcServer, cfg)
}
