package main

import (
	"context"
	"fin-api/config"
	"fin-api/internal/bootstrap"
	finapigrpc "fin-api/internal/grpc"
	finapihttp "fin-api/internal/http"
	"fin-api/internal/repository"
	"fin-api/internal/service"
	"log"

	"fin-api/internal/database"
	"fin-api/internal/kafka"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	bucketManager, err := database.NewBucketManager(ctx, cfg.Postgres)
	if err != nil {
		log.Fatalf("create shard manager: %v", err)
	}
	defer bucketManager.Close()

	producer, err := kafka.NewProducer(cfg.Kafka.Brokers, cfg.App.KafkaTopic)
	if err != nil {
		log.Fatalf("create kafka producer: %v", err)
	}
	defer producer.Close()

	repo := repository.NewTransactionRepository(bucketManager)
	svc := service.NewTransactionService(repo, producer)
	httpServer := finapihttp.NewServer(svc)
	grpcServer := finapigrpc.NewServer(svc)

	bootstrap.RunApp(ctx, cancel, httpServer, grpcServer, cfg)
}
