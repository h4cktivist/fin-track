package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"

	"fin-api/internal/domain"
)

type Producer struct {
	topic    string
	producer sarama.SyncProducer
}

func NewProducer(brokers []string, topic string) (*Producer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Retry.Max = 5
	cfg.Producer.Return.Successes = true

	p, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return nil, fmt.Errorf("create kafka producer: %w", err)
	}

	return &Producer{topic: topic, producer: p}, nil
}

func (p *Producer) PublishTransactions(ctx context.Context, msg domain.TransactionMessage) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal transaction message: %w", err)
	}

	kmsg := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.ByteEncoder(payload),
	}

	if _, _, err := p.producer.SendMessage(kmsg); err != nil {
		return fmt.Errorf("send kafka message: %w", err)
	}

	return nil
}

func (p *Producer) Close() error {
	if p.producer == nil {
		return nil
	}
	return p.producer.Close()
}
