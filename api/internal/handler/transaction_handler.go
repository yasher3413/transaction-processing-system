package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/yash/transaction-system/api/internal/service"
	"github.com/yash/transaction-system/shared/types"
	"go.uber.org/zap"
)

// TransactionHandler handles transaction HTTP requests
type TransactionHandler struct {
	transactionService *service.TransactionService
	logger             *zap.Logger
}

// NewTransactionHandler creates a new transaction handler
func NewTransactionHandler(transactionService *service.TransactionService, logger *zap.Logger) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
		logger:             logger,
	}
}

// CreateTransaction handles POST /v1/transactions
func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req types.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate
	if req.AccountID == uuid.Nil {
		h.respondError(w, http.StatusBadRequest, "account_id is required", nil)
		return
	}
	if req.AmountCents <= 0 {
		h.respondError(w, http.StatusBadRequest, "amount_cents must be positive", nil)
		return
	}
	if req.Currency == "" {
		h.respondError(w, http.StatusBadRequest, "currency is required", nil)
		return
	}
	if req.IdempotencyKey == "" {
		h.respondError(w, http.StatusBadRequest, "idempotency_key is required", nil)
		return
	}
	if req.Type != types.TransactionTypeDebit && req.Type != types.TransactionTypeCredit {
		h.respondError(w, http.StatusBadRequest, "type must be DEBIT or CREDIT", nil)
		return
	}

	transaction, err := h.transactionService.CreateTransaction(r.Context(), req)
	if err != nil {
		if err.Error() == "account not found" {
			h.respondError(w, http.StatusNotFound, "Account not found", err)
			return
		}
		if err.Error() == "account is not active" {
			h.respondError(w, http.StatusBadRequest, "Account is not active", err)
			return
		}
		h.logger.Error("Failed to create transaction", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Failed to create transaction", err)
		return
	}

	h.respondJSON(w, http.StatusCreated, transaction)
}

// GetTransaction handles GET /v1/transactions/:id
func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	transactionIDStr := chi.URLParam(r, "id")
	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid transaction ID", err)
		return
	}

	transaction, err := h.transactionService.GetTransaction(r.Context(), transactionID)
	if err != nil {
		if err.Error() == "transaction not found" {
			h.respondError(w, http.StatusNotFound, "Transaction not found", err)
			return
		}
		h.logger.Error("Failed to get transaction", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Failed to get transaction", err)
		return
	}

	h.respondJSON(w, http.StatusOK, transaction)
}

// ListTransactions handles GET /v1/transactions
func (h *TransactionHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	accountIDStr := r.URL.Query().Get("account_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var accountID *uuid.UUID
	if accountIDStr != "" {
		if id, err := uuid.Parse(accountIDStr); err == nil {
			accountID = &id
		}
	}

	transactions, err := h.transactionService.ListTransactions(r.Context(), accountID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list transactions", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Failed to list transactions", err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"transactions": transactions,
		"limit":        limit,
		"offset":       offset,
	})
}

func (h *TransactionHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *TransactionHandler) respondError(w http.ResponseWriter, status int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	response := map[string]string{
		"error": message,
	}
	if err != nil {
		response["details"] = err.Error()
	}
	json.NewEncoder(w).Encode(response)
}




