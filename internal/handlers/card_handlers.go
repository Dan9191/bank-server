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

type CardHandler struct {
	cardService services.CardService
	logger      *logrus.Logger
}

func NewCardHandler(cardService services.CardService, logger *logrus.Logger) *CardHandler {
	return &CardHandler{
		cardService: cardService,
		logger:      logger,
	}
}

func (h *CardHandler) CreateCard(w http.ResponseWriter, r *http.Request) {
	// Извлекаем user_id из контекста (для логирования или других проверок)
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		h.logger.Error("user_id not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Декодируем тело запроса
	var req struct {
		AccountID  int64  `json:"account_id"`
		CardNumber string `json:"card_number"`
		ExpiryDate string `json:"expiry_date"`
		CVV        string `json:"cvv"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request: ", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Создаём карту
	card, err := h.cardService.CreateCard(r.Context(), req.AccountID, req.CardNumber, req.ExpiryDate, req.CVV)
	if err != nil {
		h.logger.WithField("user_id", userID).Error("Failed to create card: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Формируем ответ
	resp := struct {
		ID         int64  `json:"id"`
		AccountID  int64  `json:"account_id"`
		CardNumber string `json:"card_number"`
		ExpiryDate string `json:"expiry_date"`
		CreatedAt  string `json:"created_at"`
	}{
		ID:         card.ID,
		AccountID:  card.AccountID,
		CardNumber: card.CardNumber,
		ExpiryDate: card.ExpiryDate,
		CreatedAt:  card.CreatedAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode response: ", err)
	}
}

func (h *CardHandler) GetCards(w http.ResponseWriter, r *http.Request) {
	// Извлекаем user_id из контекста (для логирования или других проверок)
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		h.logger.Error("user_id not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Извлекаем account_id из URL
	vars := mux.Vars(r)
	accountID, err := strconv.ParseInt(vars["account_id"], 10, 64)
	if err != nil {
		h.logger.Error("Invalid account ID: ", err)
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	// Получаем карты
	cards, err := h.cardService.GetCards(r.Context(), accountID)
	if err != nil {
		h.logger.WithField("user_id", userID).Error("Failed to get cards: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Формируем ответ
	resp := make([]struct {
		ID         int64  `json:"id"`
		AccountID  int64  `json:"account_id"`
		CardNumber string `json:"card_number"`
		ExpiryDate string `json:"expiry_date"`
		CreatedAt  string `json:"created_at"`
	}, len(cards))
	for i, card := range cards {
		resp[i] = struct {
			ID         int64  `json:"id"`
			AccountID  int64  `json:"account_id"`
			CardNumber string `json:"card_number"`
			ExpiryDate string `json:"expiry_date"`
			CreatedAt  string `json:"created_at"`
		}{
			ID:         card.ID,
			AccountID:  card.AccountID,
			CardNumber: card.CardNumber,
			ExpiryDate: card.ExpiryDate,
			CreatedAt:  card.CreatedAt.Format(time.RFC3339),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode response: ", err)
	}
}
