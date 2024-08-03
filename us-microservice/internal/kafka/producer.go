package kafka

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	internalConfig "urlSh/internal/config"
	"urlSh/internal/domain/models"

	"github.com/IBM/sarama"
)

type Producer struct {
	log *slog.Logger
	prd sarama.AsyncProducer
	cfg *internalConfig.Config
	ch  chan models.Url
}

func NewProducer(logger *slog.Logger, cfg *internalConfig.Config, ch chan models.Url) Producer {
	// TODO kafka configuration
	sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	config := sarama.NewConfig()
	config.Version = sarama.DefaultVersion
	config.ClientID = "us-microservice-1"
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Flush.Frequency = 500 * time.Millisecond
	config.Producer.Partitioner = sarama.NewRoundRobinPartitioner

	producer, err := sarama.NewAsyncProducer(strings.Split(cfg.Brokers, ","), config)
	if err != nil {
		logger.Error("Failed to start Sarama producer:", slog.String("err", err.Error()))
	}

	return Producer{
		log: logger,
		prd: producer,
		cfg: cfg,
		ch:  ch,
	}
}

func (p *Producer) RunProducing(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case message := <-p.ch:
			p.log.Info("trying to produce message", message)
			jsonData, err := json.Marshal(message)
			if err != nil {
				p.log.Error("Failed to marshal message:", slog.String("err", err.Error()))
				continue
			}
			p.prd.Input() <- &sarama.ProducerMessage{
				Topic: p.cfg.Topic,
				Value: sarama.ByteEncoder(jsonData),
			}
			p.log.Info("Message Produced", slog.String("url", message.UrlText))
		}
	}
}
