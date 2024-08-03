package kafka

import (
	"analys/internal/config"
	"context"
	"errors"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/IBM/sarama"
)

type MessageHandler func(message *sarama.ConsumerMessage) error

type Consumer struct {
	handler MessageHandler
	log     *slog.Logger
}

func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func NewConsumer(log *slog.Logger, handler MessageHandler) *Consumer {
	return &Consumer{
		log:     log,
		handler: handler,
	}
}

func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				consumer.log.Info("message channel was closed")
			}
			err := consumer.handler(message)
			if err != nil {
				consumer.log.Error("error on consumer handling message:", slog.String("err", err.Error()))
			}
			session.MarkMessage(message, "")
		case <-session.Context().Done():
			session.Commit()
			return nil
		}
	}
}

func InitConsumerConfig() *sarama.Config {
	sarama.Logger = log.New(os.Stdout, "[sarama]", log.LstdFlags)
	config := sarama.NewConfig()
	config.Version = sarama.DefaultVersion
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	return config
}

var group = "1"

type SaveStatsService interface {
	SaveRecord(ctx context.Context, content []byte) error
}

func RunConsumer(ctx context.Context, wg *sync.WaitGroup, log *slog.Logger, cfg *config.Config, service SaveStatsService) (sarama.ConsumerGroup, error) {
	consumer := NewConsumer(log, func(message *sarama.ConsumerMessage) error {

		err := service.SaveRecord(ctx, message.Value)
		if err != nil {
			return err
		}
		log.Info("Message claimed:", slog.String("value", string(message.Value)), slog.String("time", message.Timestamp.String()), slog.String("topic", message.Topic))
		return nil
	})
	consumerGroup, err := sarama.NewConsumerGroup(strings.Split(cfg.Brokers, ","), group, InitConsumerConfig())
	if err != nil {
		panic(err)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := consumerGroup.Consume(ctx, strings.Split(cfg.Topic, ","), consumer); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
			}
			if ctx.Err() != nil {
				return
			}

		}
	}()
	return consumerGroup, nil
}
