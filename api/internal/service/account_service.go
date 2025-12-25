package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yash/transaction-system/shared/types"
	"go.uber.org/zap"
)

// UpdateAccountBalanceMetric is defined in metrics.go

// AccountService handles account operations
type AccountService struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewAccountService creates a new account service
func NewAccountService(db *sql.DB, logger *zap.Logger) *AccountService {
	return &AccountService{
		db:     db,
		logger: logger,
	}
}

// CreateAccount creates a new account
func (s *AccountService) CreateAccount(ctx context.Context, req types.CreateAccountRequest) (*types.Account, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	id := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO accounts (id, currency, balance_cents, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at, currency, balance_cents, status
	`

	var account types.Account
	err := s.db.QueryRowContext(ctx, query,
		id, req.Currency, 0, types.AccountStatusActive, now, now,
	).Scan(
		&account.ID, &account.CreatedAt, &account.UpdatedAt,
		&account.Currency, &account.BalanceCents, &account.Status,
	)

	if err != nil {
		s.logger.Error("Failed to create account", zap.Error(err))
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	UpdateAccountBalanceMetric(account.ID.String(), account.Currency, account.BalanceCents)
	s.logger.Info("Account created", zap.String("account_id", account.ID.String()))
	return &account, nil
}

// GetAccount retrieves an account by ID
func (s *AccountService) GetAccount(ctx context.Context, accountID uuid.UUID) (*types.Account, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT id, created_at, updated_at, currency, balance_cents, status
		FROM accounts
		WHERE id = $1
	`

	var account types.Account
	err := s.db.QueryRowContext(ctx, query, accountID).Scan(
		&account.ID, &account.CreatedAt, &account.UpdatedAt,
		&account.Currency, &account.BalanceCents, &account.Status,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("account not found: %w", err)
		}
		s.logger.Error("Failed to get account", zap.Error(err), zap.String("account_id", accountID.String()))
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	UpdateAccountBalanceMetric(account.ID.String(), account.Currency, account.BalanceCents)
	return &account, nil
}
