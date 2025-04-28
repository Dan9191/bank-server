package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/bank-service/internal/services"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type AccountHandler struct {
	accountService services.AccountService
	logger         *logrus.Logger
}

func NewAccountHandler(accountService services.AccountService, logger *logrus.Logger) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
		logger:         logger,
	}
}

func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		h.logger.Error("user_id not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	account, err := h.accountService.CreateAccount(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to create account: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := struct {
		ID        int64   `json:"id"`
		UserID    int64   `json:"user_id"`
		Balance   float64 `json:"balance"`
		Currency  string  `json:"currency"`
		CreatedAt string  `json:"created_at"`
	}{
		ID:        account.ID,
		UserID:    account.UserID,
		Balance:   account.Balance,
		Currency:  account.Currency,
		CreatedAt: account.CreatedAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode response: ", err)
	}
}

func (h *AccountHandler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		h.logger.Error("user_id not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accounts, err := h.accountService.GetAccounts(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get accounts: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := make([]struct {
		ID        int64   `json:"id"`
		UserID    int64   `json:"user_id"`
		Balance   float64 `json:"balance"`
		Currency  string  `json:"currency"`
		CreatedAt string  `json:"created_at"`
	}, len(accounts))
	for i, account := range accounts {
		resp[i] = struct {
			ID        int64   `json:"id"`
			UserID    int64   `json:"user_id"`
			Balance   float64 `json:"balance"`
			Currency  string  `json:"currency"`
			CreatedAt string  `json:"created_at"`
		}{
			ID:        account.ID,
			UserID:    account.UserID,
			Balance:   account.Balance,
			Currency:  account.Currency,
			CreatedAt: account.CreatedAt.Format(time.RFC3339),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode response: ", err)
	}
}

func (h *AccountHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		h.logger.Error("user_id not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	accountID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.Error("Invalid account ID: ", err)
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request: ", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	accounts, err := h.accountService.GetAccounts(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get accounts: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var accountExists bool
	for _, account := range accounts {
		if account.ID == accountID {
			accountExists = true
			break
		}
	}
	if !accountExists {
		h.logger.Error("Account not found or unauthorized")
		http.Error(w, "Account not found or unauthorized", http.StatusForbidden)
		return
	}

	if err := h.accountService.Deposit(r.Context(), accountID, req.Amount); err != nil {
		h.logger.Error("Failed to deposit: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AccountHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		h.logger.Error("user_id not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	accountID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.Error("Invalid account ID: ", err)
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request: ", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	accounts, err := h.accountService.GetAccounts(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get accounts: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var accountExists bool
	for _, account := range accounts {
		if account.ID == accountID {
			accountExists = true
			break
		}
	}
	if !accountExists {
		h.logger.Error("Account not found or unauthorized")
		http.Error(w, "Account not found or unauthorized", http.StatusForbidden)
		return
	}

	if err := h.accountService.Withdraw(r.Context(), accountID, req.Amount); err != nil {
		h.logger.Error("Failed to withdraw: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AccountHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		h.logger.Error("user_id not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		FromAccountID int64   `json:"from_account_id"`
		ToAccountID   int64   `json:"to_account_id"`
		Amount        float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request: ", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	accounts, err := h.accountService.GetAccounts(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get accounts: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var fromAccountExists bool
	for _, account := range accounts {
		if account.ID == req.FromAccountID {
			fromAccountExists = true
			break
		}
	}
	if !fromAccountExists {
		h.logger.Error("Source account not found or unauthorized")
		http.Error(w, "Source account not found or unauthorized", http.StatusForbidden)
		return
	}

	if err := h.accountService.Transfer(r.Context(), req.FromAccountID, req.ToAccountID, req.Amount); err != nil {
		h.logger.Error("Failed to transfer: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AccountHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		h.logger.Error("user_id not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	accountID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.Error("Invalid account ID: ", err)
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	transactions, err := h.accountService.GetTransactions(r.Context(), accountID, userID)
	if err != nil {
		h.logger.Error("Failed to get transactions: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := make([]struct {
		ID          int64   `json:"id"`
		AccountID   int64   `json:"account_id"`
		Amount      float64 `json:"amount"`
		Type        string  `json:"type"`
		Description string  `json:"description"`
		CreatedAt   string  `json:"created_at"`
	}, len(transactions))
	for i, transaction := range transactions {
		resp[i] = struct {
			ID          int64   `json:"id"`
			AccountID   int64   `json:"account_id"`
			Amount      float64 `json:"amount"`
			Type        string  `json:"type"`
			Description string  `json:"description"`
			CreatedAt   string  `json:"created_at"`
		}{
			ID:          transaction.ID,
			AccountID:   transaction.AccountID,
			Amount:      transaction.Amount,
			Type:        transaction.Type,
			Description: transaction.Description,
			CreatedAt:   transaction.CreatedAt.Format(time.RFC3339),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode response: ", err)
	}
}
