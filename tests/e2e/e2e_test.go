package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yash/transaction-system/shared/types"
)

const (
	apiBaseURL = "http://localhost:8080"
	apiKey     = "demo-api-key-12345"
)

func TestE2E_TransactionFlow(t *testing.T) {
	// Create account
	accountID := createAccount(t, "USD")
	require.NotEqual(t, uuid.Nil, accountID)

	// Create transaction
	idempotencyKey := uuid.New().String()
	transactionID := createTransaction(t, accountID, 10000, "USD", types.TransactionTypeCredit, idempotencyKey)
	require.NotEqual(t, uuid.Nil, transactionID)

	// Wait for processing
	waitForTransactionStatus(t, transactionID, types.TransactionStatusProcessed, 30*time.Second)

	// Verify account balance
	account := getAccount(t, accountID)
	assert.Equal(t, int64(10000), account.BalanceCents)

	// Test idempotency - same idempotency key should return same transaction
	transactionID2 := createTransaction(t, accountID, 10000, "USD", types.TransactionTypeCredit, idempotencyKey)
	assert.Equal(t, transactionID, transactionID2)

	// Verify balance unchanged
	account = getAccount(t, accountID)
	assert.Equal(t, int64(10000), account.BalanceCents)

	// Test debit
	idempotencyKey2 := uuid.New().String()
	transactionID3 := createTransaction(t, accountID, 5000, "USD", types.TransactionTypeDebit, idempotencyKey2)
	waitForTransactionStatus(t, transactionID3, types.TransactionStatusProcessed, 30*time.Second)

	account = getAccount(t, accountID)
	assert.Equal(t, int64(5000), account.BalanceCents)
}

func TestE2E_InsufficientBalance(t *testing.T) {
	// Create account
	accountID := createAccount(t, "USD")

	// Create transaction that will fail (insufficient balance)
	idempotencyKey := uuid.New().String()
	transactionID := createTransaction(t, accountID, 10000, "USD", types.TransactionTypeDebit, idempotencyKey)

	// Wait for failure
	waitForTransactionStatus(t, transactionID, types.TransactionStatusFailed, 30*time.Second)

	// Verify transaction has failure reason
	transaction := getTransaction(t, transactionID)
	assert.NotNil(t, transaction.FailureReason)
	assert.Contains(t, *transaction.FailureReason, "insufficient balance")
}

func createAccount(t *testing.T, currency string) uuid.UUID {
	req := types.CreateAccountRequest{
		Currency: currency,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(
		fmt.Sprintf("%s/v1/accounts", apiBaseURL),
		"application/json",
		bytes.NewBuffer(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var account types.Account
	err = json.NewDecoder(resp.Body).Decode(&account)
	require.NoError(t, err)

	return account.ID
}

func createTransaction(t *testing.T, accountID uuid.UUID, amountCents int64, currency string, txType types.TransactionType, idempotencyKey string) uuid.UUID {
	req := types.CreateTransactionRequest{
		AccountID:      accountID,
		AmountCents:    amountCents,
		Currency:       currency,
		Type:           txType,
		IdempotencyKey: idempotencyKey,
	}

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1/transactions", apiBaseURL), bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var transaction types.Transaction
	err = json.NewDecoder(resp.Body).Decode(&transaction)
	require.NoError(t, err)

	return transaction.ID
}

func getAccount(t *testing.T, accountID uuid.UUID) *types.Account {
	httpReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1/accounts/%s", apiBaseURL, accountID.String()), nil)
	httpReq.Header.Set("X-API-Key", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var account types.Account
	err = json.NewDecoder(resp.Body).Decode(&account)
	require.NoError(t, err)

	return &account
}

func getTransaction(t *testing.T, transactionID uuid.UUID) *types.Transaction {
	httpReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1/transactions/%s", apiBaseURL, transactionID.String()), nil)
	httpReq.Header.Set("X-API-Key", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var transaction types.Transaction
	err = json.NewDecoder(resp.Body).Decode(&transaction)
	require.NoError(t, err)

	return &transaction
}

func waitForTransactionStatus(t *testing.T, transactionID uuid.UUID, expectedStatus types.TransactionStatus, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		transaction := getTransaction(t, transactionID)
		if transaction.Status == expectedStatus {
			return
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("Transaction %s did not reach status %s within %v", transactionID, expectedStatus, timeout)
}
