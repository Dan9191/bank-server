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

type CreditHandler struct {
	creditService services.CreditService
	logger        *logrus.Logger
}

func NewCreditHandler(creditService services.CreditService, logger *logrus.Logger) *CreditHandler {
	return &CreditHandler{
		creditService: creditService,
		logger:        logger,
	}
}

func (h *CreditHandler) CreateCredit(w http.ResponseWriter, r *http.Request) {
	// Извлекаем user_id из контекста
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		h.logger.Error("user_id not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Декодируем тело запроса
	var req struct {
		Amount       float64 `json:"amount"`
		InterestRate float64 `json:"interest_rate"`
		TermMonths   int     `json:"term_months"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request: ", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Создаём кредит
	credit, err := h.creditService.CreateCredit(r.Context(), userID, req.Amount, req.InterestRate, req.TermMonths)
	if err != nil {
		h.logger.Error("Failed to create credit: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Формируем ответ
	resp := struct {
		ID           int64     `json:"id"`
		UserID       int64     `json:"user_id"`
		Amount       float64   `json:"amount"`
		InterestRate float64   `json:"interest_rate"`
		TermMonths   int       `json:"term_months"`
		CreatedAt    time.Time `json:"created_at"`
	}{credit.ID, credit.UserID, credit.Amount, credit.InterestRate, credit.TermMonths, credit.CreatedAt}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode response: ", err)
	}
}

func (h *CreditHandler) GetCredits(w http.ResponseWriter, r *http.Request) {
	// Извлекаем user_id из контекста
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		h.logger.Error("user_id not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем кредиты
	credits, err := h.creditService.GetCredits(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get credits: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Формируем ответ
	resp := make([]struct {
		ID           int64     `json:"id"`
		UserID       int64     `json:"user_id"`
		Amount       float64   `json:"amount"`
		InterestRate float64   `json:"interest_rate"`
		TermMonths   int       `json:"term_months"`
		CreatedAt    time.Time `json:"created_at"`
	}, len(credits))
	for i, credit := range credits {
		resp[i] = struct {
			ID           int64     `json:"id"`
			UserID       int64     `json:"user_id"`
			Amount       float64   `json:"amount"`
			InterestRate float64   `json:"interest_rate"`
			TermMonths   int       `json:"term_months"`
			CreatedAt    time.Time `json:"created_at"`
		}{credit.ID, credit.UserID, credit.Amount, credit.InterestRate, credit.TermMonths, credit.CreatedAt}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode response: ", err)
	}
}

func (h *CreditHandler) GetPaymentSchedules(w http.ResponseWriter, r *http.Request) {
	// Извлекаем user_id из контекста
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		h.logger.Error("user_id not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Извлекаем credit_id из URL
	vars := mux.Vars(r)
	creditID, err := strconv.ParseInt(vars["credit_id"], 10, 64)
	if err != nil {
		h.logger.Error("Invalid credit ID: ", err)
		http.Error(w, "Invalid credit ID", http.StatusBadRequest)
		return
	}

	// Получаем график платежей
	schedules, err := h.creditService.GetPaymentSchedules(r.Context(), creditID, userID)
	if err != nil {
		h.logger.Error("Failed to get payment schedules: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Формируем ответ
	resp := make([]struct {
		ID          int64     `json:"id"`
		CreditID    int64     `json:"credit_id"`
		PaymentDate time.Time `json:"payment_date"`
		Amount      float64   `json:"amount"`
		Paid        bool      `json:"paid"`
		Penalty     float64   `json:"penalty"`
	}, len(schedules))
	for i, schedule := range schedules {
		resp[i] = struct {
			ID          int64     `json:"id"`
			CreditID    int64     `json:"credit_id"`
			PaymentDate time.Time `json:"payment_date"`
			Amount      float64   `json:"amount"`
			Paid        bool      `json:"paid"`
			Penalty     float64   `json:"penalty"`
		}{schedule.ID, schedule.CreditID, schedule.PaymentDate, schedule.Amount, schedule.Paid, schedule.Penalty}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode response: ", err)
	}
}
