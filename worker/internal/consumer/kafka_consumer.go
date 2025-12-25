package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/yash/transaction-system/shared/types"
	"github.com/yash/transaction-system/worker/internal/processor"
	"go.uber.org/zap"
)

// KafkaConsumer consumes messages from Kafka
type KafkaConsumer struct {
	reader       *kafka.Reader
	processor    *processor.TransactionProcessor
	logger       *zap.Logger
	maxRetries   int
	retryBackoff time.Duration
	dlqWriter    *kafka.Writer
}

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(
	brokers string,
	topic string,
	consumerGroup string,
	dlqTopic string,
	maxRetries int,
	retryBackoff time.Duration,
	processor *processor.TransactionProcessor,
	logger *zap.Logger,
) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokers},
		Topic:    topic,
		GroupID:  consumerGroup,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		MaxWait:  1 * time.Second,
	})

	dlqWriter := &kafka.Writer{
		Addr:         kafka.TCP(brokers),
		Topic:        dlqTopic,
		Balancer:     &kafka.LeastBytes{},
		Async:        false,
		RequiredAcks: kafka.RequireAll,
		WriteTimeout: 10 * time.Second,
	}

	return &KafkaConsumer{
		reader:       reader,
		processor:    processor,
		logger:       logger,
		maxRetries:   maxRetries,
		retryBackoff: retryBackoff,
		dlqWriter:    dlqWriter,
	}
}

// Start starts consuming messages
func (c *KafkaConsumer) Start(ctx context.Context) error {
	c.logger.Info("Kafka consumer started",
		zap.String("topic", c.reader.Config().Topic),
		zap.String("group", c.reader.Config().GroupID),
	)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Consumer stopping...")
			return nil
		default:
			if err := c.processMessage(ctx); err != nil {
				c.logger.Error("Failed to process message", zap.Error(err))
				// Continue processing other messages
			}
		}
	}
}

// processMessage processes a single message
func (c *KafkaConsumer) processMessage(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Fetch message
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch message: %w", err)
	}

	// Parse envelope
	var envelope types.EventEnvelope
	if err := json.Unmarshal(msg.Value, &envelope); err != nil {
		c.logger.Error("Failed to unmarshal message",
			zap.Error(err),
			zap.ByteString("value", msg.Value),
		)
		// Commit offset even for invalid messages (avoid infinite loop)
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.logger.Error("Failed to commit invalid message", zap.Error(err))
		}
		return err
	}

	c.logger.Debug("Processing message",
		zap.String("event_id", envelope.EventID.String()),
		zap.String("event_type", envelope.EventType),
		zap.Int("partition", msg.Partition),
		zap.Int64("offset", msg.Offset),
	)

	// Process with retries
	var lastErr error
	shouldRetry := true

	for attempt := 0; attempt < c.maxRetries && shouldRetry; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt) * c.retryBackoff
			processor.RetryCounter.WithLabelValues(envelope.EventType).Inc()
			c.logger.Info("Retrying message",
				zap.String("event_id", envelope.EventID.String()),
				zap.Int("attempt", attempt+1),
				zap.Duration("backoff", backoff),
			)
			time.Sleep(backoff)
		}

		shouldRetry, lastErr = c.processor.ProcessTransactionCreated(ctx, envelope)

		if !shouldRetry {
			// Success or non-retryable error
			break
		}
	}

	// If still failed after retries, send to DLQ
	if shouldRetry && lastErr != nil {
		c.logger.Error("Message failed after max retries, sending to DLQ",
			zap.String("event_id", envelope.EventID.String()),
			zap.Int("attempts", c.maxRetries),
			zap.Error(lastErr),
		)

		processor.DLQMessagesTotal.Inc()
		if err := c.sendToDLQ(ctx, msg, envelope, lastErr); err != nil {
			c.logger.Error("Failed to send to DLQ", zap.Error(err))
			// Don't commit - will retry
			return err
		}
	}

	// Commit offset
	if err := c.reader.CommitMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to commit message: %w", err)
	}

	return nil
}

// sendToDLQ sends a failed message to the dead-letter queue
func (c *KafkaConsumer) sendToDLQ(ctx context.Context, originalMsg kafka.Message, envelope types.EventEnvelope, err error) error {
	dlqMessage := kafka.Message{
		Key:   originalMsg.Key,
		Value: originalMsg.Value,
		Headers: append(originalMsg.Headers,
			kafka.Header{Key: "dlq_reason", Value: []byte(err.Error())},
			kafka.Header{Key: "original_partition", Value: []byte(fmt.Sprintf("%d", originalMsg.Partition))},
			kafka.Header{Key: "original_offset", Value: []byte(fmt.Sprintf("%d", originalMsg.Offset))},
		),
	}

	if err := c.dlqWriter.WriteMessages(ctx, dlqMessage); err != nil {
		return fmt.Errorf("failed to write to DLQ: %w", err)
	}

	c.logger.Info("Message sent to DLQ",
		zap.String("event_id", envelope.EventID.String()),
		zap.String("reason", err.Error()),
	)

	return nil
}

// Close closes the consumer
func (c *KafkaConsumer) Close() error {
	if err := c.reader.Close(); err != nil {
		return err
	}
	return c.dlqWriter.Close()
}
