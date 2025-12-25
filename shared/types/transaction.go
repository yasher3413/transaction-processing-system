package types

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeDebit  TransactionType = "DEBIT"
	TransactionTypeCredit TransactionType = "CREDIT"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending    TransactionStatus = "PENDING"
	TransactionStatusProcessing TransactionStatus = "PROCESSING"
	TransactionStatusProcessed  TransactionStatus = "PROCESSED"
	TransactionStatusFailed     TransactionStatus = "FAILED"
)

// AccountStatus represents the status of an account
type AccountStatus string

const (
	AccountStatusActive    AccountStatus = "ACTIVE"
	AccountStatusSuspended AccountStatus = "SUSPENDED"
)

// Account represents a financial account
type Account struct {
	ID          uuid.UUID    `json:"id"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Currency    string       `json:"currency"`
	BalanceCents int64      `json:"balance_cents"`
	Status      AccountStatus `json:"status"`
}

// Transaction represents a financial transaction
type Transaction struct {
	ID             uuid.UUID        `json:"id"`
	AccountID      uuid.UUID        `json:"account_id"`
	AmountCents    int64            `json:"amount_cents"`
	Currency       string           `json:"currency"`
	Type           TransactionType  `json:"type"`
	Status         TransactionStatus `json:"status"`
	IdempotencyKey string           `json:"idempotency_key"`
	FailureReason  *string          `json:"failure_reason,omitempty"`
	Metadata       json.RawMessage  `json:"metadata,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// CreateTransactionRequest represents a request to create a transaction
type CreateTransactionRequest struct {
	AccountID      uuid.UUID       `json:"account_id"`
	AmountCents    int64           `json:"amount_cents"`
	Currency       string          `json:"currency"`
	Type           TransactionType `json:"type"`
	IdempotencyKey string          `json:"idempotency_key"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
}

// CreateAccountRequest represents a request to create an account
type CreateAccountRequest struct {
	Currency string `json:"currency"`
}

// EventEnvelope represents a message envelope for event streaming
type EventEnvelope struct {
	EventID        uuid.UUID       `json:"event_id"`
	EventType      string          `json:"event_type"`
	OccurredAt     time.Time       `json:"occurred_at"`
	TraceID        string          `json:"trace_id"`
	IdempotencyKey string          `json:"idempotency_key"`
	AggregateID    uuid.UUID       `json:"aggregate_id"`
	Payload        json.RawMessage `json:"payload"`
}

// TransactionCreatedPayload represents the payload for transaction.created event
type TransactionCreatedPayload struct {
	TransactionID  uuid.UUID       `json:"transaction_id"`
	AccountID      uuid.UUID       `json:"account_id"`
	AmountCents    int64           `json:"amount_cents"`
	Currency       string          `json:"currency"`
	Type           TransactionType `json:"type"`
	IdempotencyKey string          `json:"idempotency_key"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
}

// TransactionProcessedPayload represents the payload for transaction.processed event
type TransactionProcessedPayload struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	AccountID     uuid.UUID `json:"account_id"`
	NewBalance    int64     `json:"new_balance"`
}

// TransactionFailedPayload represents the payload for transaction.failed event
type TransactionFailedPayload struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	AccountID     uuid.UUID `json:"account_id"`
	FailureReason string    `json:"failure_reason"`
}




