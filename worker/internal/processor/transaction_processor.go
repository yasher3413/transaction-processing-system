package processor

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yash/transaction-system/shared/types"
	"go.uber.org/zap"
)

// TransactionProcessor processes transaction events
type TransactionProcessor struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewTransactionProcessor creates a new transaction processor
func NewTransactionProcessor(db *sql.DB, logger *zap.Logger) *TransactionProcessor {
	return &TransactionProcessor{
		db:     db,
		logger: logger,
	}
}

// ProcessTransactionCreated processes a transaction.created event
// Returns: (shouldRetry bool, error)
func (p *TransactionProcessor) ProcessTransactionCreated(ctx context.Context, envelope types.EventEnvelope) (bool, error) {
	start := time.Now()
	defer func() {
		workerProcessingDuration.WithLabelValues(envelope.EventType).Observe(time.Since(start).Seconds())
	}()

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Parse payload
	var payload types.TransactionCreatedPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return false, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Check idempotency: has this event been processed?
	var processedEventID uuid.UUID
	checkQuery := `SELECT event_id FROM processed_events WHERE event_id = $1`
	err := p.db.QueryRowContext(ctx, checkQuery, envelope.EventID).Scan(&processedEventID)

	if err == nil {
		// Already processed - idempotent no-op
		eventsConsumedTotal.WithLabelValues(envelope.EventType, "duplicate").Inc()
		p.logger.Info("Event already processed (idempotent)",
			zap.String("event_id", envelope.EventID.String()),
			zap.String("transaction_id", payload.TransactionID.String()),
		)
		return false, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return true, fmt.Errorf("failed to check idempotency: %w", err)
	}

	// Start transaction for atomic processing
	tx, err := p.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return true, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			p.logger.Error("Failed to rollback transaction", zap.Error(err))
		}
	}()

	// Insert into processed_events first (idempotency check)
	insertProcessedQuery := `
		INSERT INTO processed_events (event_id, transaction_id)
		VALUES ($1, $2)
		ON CONFLICT (event_id) DO NOTHING
	`
	result, err := tx.ExecContext(ctx, insertProcessedQuery, envelope.EventID, payload.TransactionID)
	if err != nil {
		return true, fmt.Errorf("failed to insert processed event: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Another worker processed it - no-op
		tx.Commit()
		p.logger.Info("Event processed by another worker (idempotent)",
			zap.String("event_id", envelope.EventID.String()),
		)
		return false, nil
	}

	// Update transaction status to PROCESSING
	updateStatusQuery := `
		UPDATE transactions
		SET status = 'PROCESSING', updated_at = NOW()
		WHERE id = $1 AND status = 'PENDING'
	`
	_, err = tx.ExecContext(ctx, updateStatusQuery, payload.TransactionID)
	if err != nil {
		return true, fmt.Errorf("failed to update transaction status: %w", err)
	}

	// Lock account row and update balance
	var currentBalance int64
	var accountCurrency string
	lockAccountQuery := `
		SELECT balance_cents, currency
		FROM accounts
		WHERE id = $1
		FOR UPDATE
	`
	err = tx.QueryRowContext(ctx, lockAccountQuery, payload.AccountID).Scan(&currentBalance, &accountCurrency)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("account not found")
		}
		return true, fmt.Errorf("failed to lock account: %w", err)
	}

	// Validate currency match
	if accountCurrency != payload.Currency {
		return false, fmt.Errorf("currency mismatch: account=%s, transaction=%s", accountCurrency, payload.Currency)
	}

	// Calculate new balance
	var newBalance int64
	if payload.Type == types.TransactionTypeCredit {
		newBalance = currentBalance + payload.AmountCents
	} else { // DEBIT
		newBalance = currentBalance - payload.AmountCents
	}

	// Validate debit doesn't go negative (business rule)
	if newBalance < 0 && payload.Type == types.TransactionTypeDebit {
		// Mark transaction as failed
		failureReason := fmt.Sprintf("insufficient balance: current=%d, debit=%d", currentBalance, payload.AmountCents)
		failQuery := `
			UPDATE transactions
			SET status = 'FAILED', failure_reason = $1, updated_at = NOW()
			WHERE id = $2
		`
		_, err = tx.ExecContext(ctx, failQuery, failureReason, payload.TransactionID)
		if err != nil {
			return true, fmt.Errorf("failed to mark transaction as failed: %w", err)
		}

		eventsConsumedTotal.WithLabelValues(envelope.EventType, "failed").Inc()
		tx.Commit()
		return false, fmt.Errorf("insufficient balance: %s", failureReason)
	}

	// Update account balance
	updateBalanceQuery := `
		UPDATE accounts
		SET balance_cents = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err = tx.ExecContext(ctx, updateBalanceQuery, newBalance, payload.AccountID)
	if err != nil {
		return true, fmt.Errorf("failed to update balance: %w", err)
	}

	// Mark transaction as PROCESSED
	markProcessedQuery := `
		UPDATE transactions
		SET status = 'PROCESSED', updated_at = NOW()
		WHERE id = $1
	`
	_, err = tx.ExecContext(ctx, markProcessedQuery, payload.TransactionID)
	if err != nil {
		return true, fmt.Errorf("failed to mark transaction as processed: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return true, fmt.Errorf("failed to commit transaction: %w", err)
	}

	eventsConsumedTotal.WithLabelValues(envelope.EventType, "success").Inc()
	p.logger.Info("Transaction processed successfully",
		zap.String("transaction_id", payload.TransactionID.String()),
		zap.String("account_id", payload.AccountID.String()),
		zap.Int64("old_balance", currentBalance),
		zap.Int64("new_balance", newBalance),
		zap.String("type", string(payload.Type)),
	)

	return false, nil
}
