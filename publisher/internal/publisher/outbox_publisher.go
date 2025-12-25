package publisher

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/yash/transaction-system/shared/types"
	"go.uber.org/zap"
)

// OutboxPublisher publishes outbox events to Kafka
type OutboxPublisher struct {
	db          *sql.DB
	writer      *kafka.Writer
	logger      *zap.Logger
	batchSize   int
	pollInterval time.Duration
}

// NewOutboxPublisher creates a new outbox publisher
func NewOutboxPublisher(
	db *sql.DB,
	kafkaBrokers string,
	topic string,
	batchSize int,
	pollInterval time.Duration,
	logger *zap.Logger,
) *OutboxPublisher {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(kafkaBrokers),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
		Async:    false, // Synchronous for reliability
		RequiredAcks: kafka.RequireAll, // Wait for all replicas
		WriteTimeout: 10 * time.Second,
	}

	return &OutboxPublisher{
		db:          db,
		writer:      writer,
		logger:      logger,
		batchSize:   batchSize,
		pollInterval: pollInterval,
	}
}

// Start starts the publisher loop
func (p *OutboxPublisher) Start(ctx context.Context) error {
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	p.logger.Info("Outbox publisher started",
		zap.Int("batch_size", p.batchSize),
		zap.Duration("poll_interval", p.pollInterval),
	)

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("Outbox publisher stopping...")
			return nil
		case <-ticker.C:
			if err := p.publishBatch(ctx); err != nil {
				p.logger.Error("Failed to publish batch", zap.Error(err))
				// Continue - will retry on next tick
			}
		}
	}
}

// publishBatch publishes a batch of pending outbox events
func (p *OutboxPublisher) publishBatch(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Fetch pending events
	query := `
		SELECT id, aggregate_type, aggregate_id, event_type, payload, created_at
		FROM outbox_events
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`

	rows, err := p.db.QueryContext(ctx, query, p.batchSize)
	if err != nil {
		return fmt.Errorf("failed to query outbox events: %w", err)
	}
	defer rows.Close()

	var events []outboxEvent
	for rows.Next() {
		var event outboxEvent
		err := rows.Scan(
			&event.ID, &event.AggregateType, &event.AggregateID,
			&event.EventType, &event.Payload, &event.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	if len(events) == 0 {
		return nil // No events to publish
	}

	p.logger.Debug("Publishing batch", zap.Int("count", len(events)))

	// Publish each event
	for _, event := range events {
		if err := p.publishEvent(ctx, event); err != nil {
			// Update error in DB but continue with other events
			p.updateEventError(ctx, event.ID, err.Error())
			continue
		}

		// Mark as published
		if err := p.markAsPublished(ctx, event.ID); err != nil {
			p.logger.Error("Failed to mark event as published",
				zap.String("event_id", event.ID.String()),
				zap.Error(err),
			)
		}
	}

	return nil
}

// publishEvent publishes a single event to Kafka
func (p *OutboxPublisher) publishEvent(ctx context.Context, event outboxEvent) error {
	// Create event envelope
	envelope := types.EventEnvelope{
		EventID:        uuid.New(),
		EventType:      event.EventType,
		OccurredAt:     event.CreatedAt,
		TraceID:        "", // Will be set by tracing middleware if available
		IdempotencyKey: "", // Will be extracted from payload if needed
		AggregateID:    event.AggregateID,
		Payload:        event.Payload,
	}

	// Extract idempotency key from payload if it's a transaction event
	if event.EventType == "transaction.created" {
		var payload types.TransactionCreatedPayload
		if err := json.Unmarshal(event.Payload, &payload); err == nil {
			envelope.IdempotencyKey = payload.IdempotencyKey
		}
	}

	envelopeBytes, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("failed to marshal envelope: %w", err)
	}

	// Publish to Kafka with transaction ID as key for partitioning
	message := kafka.Message{
		Key:   []byte(event.AggregateID.String()),
		Value: envelopeBytes,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(event.EventType)},
			{Key: "aggregate_id", Value: []byte(event.AggregateID.String())},
		},
	}

	if err := p.writer.WriteMessages(ctx, message); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	p.logger.Info("Event published",
		zap.String("event_id", envelope.EventID.String()),
		zap.String("event_type", event.EventType),
		zap.String("aggregate_id", event.AggregateID.String()),
	)

	return nil
}

// markAsPublished marks an outbox event as published
func (p *OutboxPublisher) markAsPublished(ctx context.Context, eventID uuid.UUID) error {
	query := `
		UPDATE outbox_events
		SET status = 'PUBLISHED', published_at = NOW()
		WHERE id = $1
	`

	_, err := p.db.ExecContext(ctx, query, eventID)
	return err
}

// updateEventError updates the error for an event
func (p *OutboxPublisher) updateEventError(ctx context.Context, eventID uuid.UUID, errorMsg string) error {
	query := `
		UPDATE outbox_events
		SET publish_attempts = publish_attempts + 1, last_error = $1
		WHERE id = $2
	`

	_, err := p.db.ExecContext(ctx, query, errorMsg, eventID)
	return err
}

// Close closes the publisher
func (p *OutboxPublisher) Close() error {
	return p.writer.Close()
}

type outboxEvent struct {
	ID           uuid.UUID
	AggregateType string
	AggregateID  uuid.UUID
	EventType    string
	Payload      json.RawMessage
	CreatedAt    time.Time
}




