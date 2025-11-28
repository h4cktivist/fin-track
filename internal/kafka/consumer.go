package kafka

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
)

type MessageHandler func(ctx context.Context, message *sarama.ConsumerMessage) error

type Consumer struct {
	group   sarama.ConsumerGroup
	topics  []string
	handler MessageHandler
}

func NewConsumer(brokers []string, groupID string, topics []string, handler MessageHandler) (*Consumer, error) {
	cfg := sarama.NewConfig()
	cfg.Consumer.Return.Errors = true
	cfg.Version = sarama.V3_5_0_0

	group, err := sarama.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		return nil, fmt.Errorf("create consumer group: %w", err)
	}

	return &Consumer{
		group:   group,
		topics:  topics,
		handler: handler,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	for {
		if err := c.group.Consume(ctx, c.topics, consumerGroupHandler{handler: c.handler}); err != nil {
			return fmt.Errorf("consume: %w", err)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (c *Consumer) Close() error {
	if c.group == nil {
		return nil
	}
	return c.group.Close()
}

type consumerGroupHandler struct {
	handler MessageHandler
}

func (consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		if err := h.handler(session.Context(), msg); err != nil {
			return err
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
