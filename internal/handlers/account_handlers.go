package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/bank-service/internal/services"
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
	// Извлекаем user_id из контекста (добавлен middleware)
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		h.logger.Error("user_id not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Создаем счет
	account, err := h.accountService.CreateAccount(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to create account: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Формируем ответ
	resp := struct {
		ID       int64   `json:"id"`
		UserID   int64   `json:"user_id"`
		Balance  float64 `json:"balance"`
		Currency string  `json:"currency"`
	}{account.ID, account.UserID, account.Balance, account.Currency}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode response: ", err)
	}
}

func (h *AccountHandler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	// Извлекаем user_id из контекста
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		h.logger.Error("user_id not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем счета
	accounts, err := h.accountService.GetAccounts(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get accounts: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Формируем ответ
	resp := make([]struct {
		ID       int64   `json:"id"`
		UserID   int64   `json:"user_id"`
		Balance  float64 `json:"balance"`
		Currency string  `json:"currency"`
	}, len(accounts))
	for i, account := range accounts {
		resp[i] = struct {
			ID       int64   `json:"id"`
			UserID   int64   `json:"user_id"`
			Balance  float64 `json:"balance"`
			Currency string  `json:"currency"`
		}{account.ID, account.UserID, account.Balance, account.Currency}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode response: ", err)
	}
}
