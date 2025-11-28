package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"

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
	httpAddr := fmt.Sprintf("%s:%d", cfg.FinAnalytics.HTTPHost, cfg.FinAnalytics.HTTPPort)

	errCh := make(chan error, 2)

	go func() {
		log.Printf("fin-analytics HTTP listening on %s", httpAddr)
		errCh <- analyticsHTTP.Start(ctx, httpAddr)
	}()

	go func() {
		log.Println("fin-analytics consumer started")
		errCh <- kafkaConsumer.Start(ctx)
	}()

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		cancel()
	}()

	if err := <-errCh; err != nil && err != context.Canceled && err != sarama.ErrClosedConsumerGroup {
		log.Fatalf("service error: %v", err)
	}
}
