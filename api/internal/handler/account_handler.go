package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/yash/transaction-system/api/internal/service"
	"github.com/yash/transaction-system/shared/types"
	"go.uber.org/zap"
)

// AccountHandler handles account HTTP requests
type AccountHandler struct {
	accountService *service.AccountService
	logger         *zap.Logger
}

// NewAccountHandler creates a new account handler
func NewAccountHandler(accountService *service.AccountService, logger *zap.Logger) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
		logger:         logger,
	}
}

// CreateAccount handles POST /v1/accounts
func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req types.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate
	if req.Currency == "" {
		h.respondError(w, http.StatusBadRequest, "Currency is required", nil)
		return
	}

	account, err := h.accountService.CreateAccount(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create account", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Failed to create account", err)
		return
	}

	h.respondJSON(w, http.StatusCreated, account)
}

// GetAccount handles GET /v1/accounts/:id
func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	accountIDStr := chi.URLParam(r, "id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid account ID", err)
		return
	}

	account, err := h.accountService.GetAccount(r.Context(), accountID)
	if err != nil {
		if err.Error() == "account not found: sql: no rows in result set" {
			h.respondError(w, http.StatusNotFound, "Account not found", err)
			return
		}
		h.logger.Error("Failed to get account", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Failed to get account", err)
		return
	}

	h.respondJSON(w, http.StatusOK, account)
}

func (h *AccountHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *AccountHandler) respondError(w http.ResponseWriter, status int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	response := map[string]string{
		"error": message,
	}
	if err != nil {
		response["details"] = err.Error()
	}
	_ = json.NewEncoder(w).Encode(response)
}
