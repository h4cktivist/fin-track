package main

import (
	"context"
	"fin-track-app/internal/finanalytics/bootstrap"
	"log"
	"time"

	"fin-track-app/internal/config"
	"fin-track-app/internal/database"
	"fin-track-app/internal/finanalytics/cache"
	"fin-track-app/internal/finanalytics/grpcclient"
	finanalyticshttp "fin-track-app/internal/finanalytics/http"
	"fin-track-app/internal/finanalytics/service"
	"fin-track-app/internal/kafka"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	redisClient, err := database.NewRedisClient(ctx, cfg.Redis)
	if err != nil {
		log.Fatalf("connect redis: %v", err)
	}
	defer redisClient.Close()

	cache := cache.New(redisClient, 15*time.Minute)

	grpcClient, err := grpcclient.New(cfg.FinAPI.GRPCTarget)
	if err != nil {
		log.Fatalf("grpc client: %v", err)
	}

	svc := service.New(cache, grpcClient)

	kafkaConsumer, err := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, []string{cfg.App.KafkaTopic}, svc.ProcessKafkaMessage)
	if err != nil {
		log.Fatalf("kafka consumer: %v", err)
	}
	defer kafkaConsumer.Close()

	analyticsHTTP := finanalyticshttp.NewServer(svc)

	bootstrap.RunAnalyticsApp(ctx, cancel, analyticsHTTP, kafkaConsumer, cfg)
}
