package service

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

// TransactionService handles transaction operations
type TransactionService struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewTransactionService creates a new transaction service
func NewTransactionService(db *sql.DB, logger *zap.Logger) *TransactionService {
	return &TransactionService{
		db:     db,
		logger: logger,
	}
}

// CreateTransaction creates a transaction with idempotency check and outbox write
func (s *TransactionService) CreateTransaction(ctx context.Context, req types.CreateTransactionRequest) (*types.Transaction, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Start transaction
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			s.logger.Error("Failed to rollback transaction", zap.Error(err))
		}
	}()

	// Check idempotency: if same (account_id, idempotency_key) exists, return it
	var existingTx types.Transaction
	var existingMetadata []byte
	checkQuery := `
		SELECT id, account_id, amount_cents, currency, type, status, idempotency_key, 
		       failure_reason, metadata, created_at, updated_at
		FROM transactions
		WHERE account_id = $1 AND idempotency_key = $2
		LIMIT 1
	`
	err = tx.QueryRowContext(ctx, checkQuery, req.AccountID, req.IdempotencyKey).Scan(
		&existingTx.ID, &existingTx.AccountID, &existingTx.AmountCents,
		&existingTx.Currency, &existingTx.Type, &existingTx.Status,
		&existingTx.IdempotencyKey, &existingTx.FailureReason,
		&existingMetadata, &existingTx.CreatedAt, &existingTx.UpdatedAt,
	)
	if err == nil {
		existingTx.Metadata = existingMetadata
	}

	if err == nil {
		// Idempotent request - return existing transaction
		tx.Commit()
		s.logger.Info("Idempotent transaction request",
			zap.String("transaction_id", existingTx.ID.String()),
			zap.String("idempotency_key", req.IdempotencyKey),
		)
		return &existingTx, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to check idempotency: %w", err)
	}

	// Validate account exists
	var accountStatus string
	accountQuery := `SELECT status FROM accounts WHERE id = $1`
	err = tx.QueryRowContext(ctx, accountQuery, req.AccountID).Scan(&accountStatus)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("failed to validate account: %w", err)
	}

	if accountStatus != string(types.AccountStatusActive) {
		return nil, fmt.Errorf("account is not active")
	}

	// Validate amount
	if req.AmountCents <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	// Create transaction
	txID := uuid.New()
	now := time.Now()

	// Handle metadata - convert empty to nil for JSONB
	var metadataValue interface{}
	if len(req.Metadata) == 0 {
		metadataValue = nil
	} else {
		metadataValue = req.Metadata
	}

	insertTxQuery := `
		INSERT INTO transactions (id, account_id, amount_cents, currency, type, status, 
		                          idempotency_key, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, account_id, amount_cents, currency, type, status, idempotency_key,
		          failure_reason, metadata, created_at, updated_at
	`

	var transaction types.Transaction
	var metadataBytes []byte
	err = tx.QueryRowContext(ctx, insertTxQuery,
		txID, req.AccountID, req.AmountCents, req.Currency, req.Type,
		types.TransactionStatusPending, req.IdempotencyKey, metadataValue, now, now,
	).Scan(
		&transaction.ID, &transaction.AccountID, &transaction.AmountCents,
		&transaction.Currency, &transaction.Type, &transaction.Status,
		&transaction.IdempotencyKey, &transaction.FailureReason,
		&metadataBytes, &transaction.CreatedAt, &transaction.UpdatedAt,
	)
	if err == nil {
		transaction.Metadata = metadataBytes
	}

	if err != nil {
		// Check if it's a unique constraint violation (race condition)
		if err.Error() == "pq: duplicate key value violates unique constraint \"transactions_account_id_idempotency_key_key\"" {
			// Another request created it - fetch and return
			var raceMetadata []byte
			err = tx.QueryRowContext(ctx, checkQuery, req.AccountID, req.IdempotencyKey).Scan(
				&transaction.ID, &transaction.AccountID, &transaction.AmountCents,
				&transaction.Currency, &transaction.Type, &transaction.Status,
				&transaction.IdempotencyKey, &transaction.FailureReason,
				&raceMetadata, &transaction.CreatedAt, &transaction.UpdatedAt,
			)
			if err == nil {
				transaction.Metadata = raceMetadata
				tx.Commit()
				return &transaction, nil
			}
		}
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Create outbox event
	payload := types.TransactionCreatedPayload{
		TransactionID:  transaction.ID,
		AccountID:      transaction.AccountID,
		AmountCents:    transaction.AmountCents,
		Currency:       transaction.Currency,
		Type:           transaction.Type,
		IdempotencyKey: transaction.IdempotencyKey,
		Metadata:       transaction.Metadata,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	outboxID := uuid.New()
	outboxQuery := `
		INSERT INTO outbox_events (id, aggregate_type, aggregate_id, event_type, payload, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = tx.ExecContext(ctx, outboxQuery,
		outboxID, "transaction", transaction.ID, "transaction.created",
		payloadBytes, "PENDING", now,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create outbox event: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("Transaction created with outbox event",
		zap.String("transaction_id", transaction.ID.String()),
		zap.String("idempotency_key", req.IdempotencyKey),
	)

	return &transaction, nil
}

// GetTransaction retrieves a transaction by ID
func (s *TransactionService) GetTransaction(ctx context.Context, transactionID uuid.UUID) (*types.Transaction, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT id, account_id, amount_cents, currency, type, status, idempotency_key,
		       failure_reason, metadata, created_at, updated_at
		FROM transactions
		WHERE id = $1
	`

	var transaction types.Transaction
	var metadataBytes []byte
	err := s.db.QueryRowContext(ctx, query, transactionID).Scan(
		&transaction.ID, &transaction.AccountID, &transaction.AmountCents,
		&transaction.Currency, &transaction.Type, &transaction.Status,
		&transaction.IdempotencyKey, &transaction.FailureReason,
		&metadataBytes, &transaction.CreatedAt, &transaction.UpdatedAt,
	)
	if err == nil {
		transaction.Metadata = metadataBytes
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("transaction not found")
		}
		s.logger.Error("Failed to get transaction", zap.Error(err), zap.String("transaction_id", transactionID.String()))
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &transaction, nil
}

// ListTransactions lists transactions with pagination
func (s *TransactionService) ListTransactions(ctx context.Context, accountID *uuid.UUID, limit, offset int) ([]types.Transaction, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var query string
	var args []interface{}

	if accountID != nil {
		query = `
			SELECT id, account_id, amount_cents, currency, type, status, idempotency_key,
			       failure_reason, metadata, created_at, updated_at
			FROM transactions
			WHERE account_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{accountID, limit, offset}
	} else {
		query = `
			SELECT id, account_id, amount_cents, currency, type, status, idempotency_key,
			       failure_reason, metadata, created_at, updated_at
			FROM transactions
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []types.Transaction
	for rows.Next() {
		var tx types.Transaction
		var metadataBytes []byte
		err := rows.Scan(
			&tx.ID, &tx.AccountID, &tx.AmountCents,
			&tx.Currency, &tx.Type, &tx.Status,
			&tx.IdempotencyKey, &tx.FailureReason,
			&metadataBytes, &tx.CreatedAt, &tx.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		tx.Metadata = metadataBytes
		transactions = append(transactions, tx)
	}

	return transactions, nil
}
