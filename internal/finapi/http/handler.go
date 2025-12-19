package http

import (
	"encoding/json"
	"errors"
	stdhttp "net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"fin-track-app/internal/domain"
)

type transactionRequest struct {
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
	Type     string  `json:"type"`
}

func (s *Server) handleCreateTransaction(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		httpError(w, stdhttp.StatusBadRequest, "invalid user_id")
		return
	}

	var req transactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, stdhttp.StatusBadRequest, "invalid payload")
		return
	}

	txType := domain.TransactionType(req.Type)
	if txType != domain.TransactionTypeIncome && txType != domain.TransactionTypeExpense {
		httpError(w, stdhttp.StatusBadRequest, "type must be income or expense")
		return
	}

	tx := domain.Transaction{
		UserID:   userID,
		Amount:   req.Amount,
		Category: req.Category,
		Type:     txType,
	}

	created, err := s.service.CreateTransaction(r.Context(), tx)
	if err != nil {
		httpError(w, stdhttp.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, stdhttp.StatusCreated, created)
}

func (s *Server) handleListTransactions(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		httpError(w, stdhttp.StatusBadRequest, "invalid user_id")
		return
	}

	items, err := s.service.ListTransactions(r.Context(), userID)
	if err != nil {
		httpError(w, stdhttp.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, stdhttp.StatusOK, items)
}

func (s *Server) handleUpdateTransaction(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		httpError(w, stdhttp.StatusBadRequest, "invalid user_id")
		return
	}

	transactionID, err := parseTransactionID(r)
	if err != nil {
		httpError(w, stdhttp.StatusBadRequest, "invalid transaction id")
		return
	}

	var req transactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, stdhttp.StatusBadRequest, "invalid payload")
		return
	}

	txType := domain.TransactionType(req.Type)
	if txType != domain.TransactionTypeIncome && txType != domain.TransactionTypeExpense {
		httpError(w, stdhttp.StatusBadRequest, "type must be income or expense")
		return
	}

	updated, err := s.service.UpdateTransaction(r.Context(), domain.Transaction{
		ID:       transactionID,
		UserID:   userID,
		Amount:   req.Amount,
		Category: req.Category,
		Type:     txType,
	})
	if err != nil {
		if errors.Is(err, domain.ErrTransactionNotFound) {
			httpError(w, stdhttp.StatusNotFound, err.Error())
			return
		}
		httpError(w, stdhttp.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, stdhttp.StatusOK, updated)
}

func (s *Server) handleDeleteTransaction(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		httpError(w, stdhttp.StatusBadRequest, "invalid user_id")
		return
	}

	transactionID, err := parseTransactionID(r)
	if err != nil {
		httpError(w, stdhttp.StatusBadRequest, "invalid transaction id")
		return
	}

	if err := s.service.DeleteTransaction(r.Context(), userID, transactionID); err != nil {
		if errors.Is(err, domain.ErrTransactionNotFound) {
			httpError(w, stdhttp.StatusNotFound, err.Error())
			return
		}
		httpError(w, stdhttp.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(stdhttp.StatusNoContent)
}

func httpError(w stdhttp.ResponseWriter, code int, message string) {
	writeJSON(w, code, map[string]string{"error": message})
}

func writeJSON(w stdhttp.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func parseTransactionID(r *stdhttp.Request) (int64, error) {
	raw := chi.URLParam(r, "transactionID")
	return strconv.ParseInt(raw, 10, 64)
}
