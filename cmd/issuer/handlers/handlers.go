package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/franmeza/payment-flow-simulation/cmd/issuer/utils"
	"github.com/franmeza/payment-flow-simulation/internal/db"
	"github.com/franmeza/payment-flow-simulation/internal/models"
	"github.com/franmeza/payment-flow-simulation/internal/rules"
)

type Handler struct {
	database *db.DB
}

// New builds a Handler with its required dependencies.
func New(database *db.DB) *Handler {
	return &Handler{database: database}
}

// Authorize validates an auth request, delegates business logic, and returns JSON.
func (h *Handler) Authorize(w http.ResponseWriter, r *http.Request) {
	// This endpoint only accepts POST requests.
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse incoming JSON into the shared AuthRequest model.
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("[ISSUER] Auth request — card: %s amount: $%.2f", req.CardUID, req.Amount)

	// Simulate issuer-side processing latency.
	time.Sleep(120 * time.Millisecond)

	resp := h.processAuth(req)

	// Always return a JSON response payload.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// processAuth applies issuer rules and persists the auth attempt result.
func (h *Handler) processAuth(req models.AuthRequest) models.AuthResponse {
	// Fetch card details needed for validation and balance checks.
	card, err := h.database.GetCard(req.CardUID)
	if err != nil {
		log.Printf("[ISSUER] Unknown card: %s", req.CardUID)
		return models.AuthResponse{
			Approved:      false,
			DeclineReason: "card not found",
			LastFour:      utils.LastFour(req.CardUID),
			Timestamp:     time.Now(),
		}
	}

	// Run rules checks
	result := rules.Check(card, req.Amount)

	// Persist every auth attempt (approved or declined) for auditing.
	tx := models.Transaction{
		ID:         fmt.Sprintf("TXN-%d", time.Now().UnixNano()),
		CardUID:    req.CardUID,
		MerchantID: req.MerchantID,
		Amount:     req.Amount,
		Approved:   result.Approved,
		Reason:     result.Reason,
		Timestamp:  time.Now(),
	}
	if err := h.database.SaveTransaction(tx); err != nil {
		log.Printf("[ISSUER] failed to save transaction for card %s: %v", req.CardUID, err)
		return models.AuthResponse{
			Approved:      false,
			DeclineReason: "issuer database error",
			LastFour:      utils.LastFour(req.CardUID),
			Timestamp:     time.Now(),
		}
	}

	// Return early for declined authorizations after recording the attempt.
	if !result.Approved {
		log.Printf("[ISSUER] DECLINED — %s", result.Reason)
		return models.AuthResponse{
			Approved:      false,
			DeclineReason: result.Reason,
			LastFour:      utils.LastFour(req.CardUID),
			Timestamp:     time.Now(),
		}
	}

	// For approved requests, deduct funds from the card balance.
	if err := h.database.DeductBalance(req.CardUID, req.Amount); err != nil {
		log.Printf("[ISSUER] failed to deduct balance for card %s: %v", req.CardUID, err)
		return models.AuthResponse{
			Approved:      false,
			DeclineReason: "issuer database error",
			LastFour:      utils.LastFour(req.CardUID),
			Timestamp:     time.Now(),
		}
	}

	log.Printf("[ISSUER] APPROVED — %s $%.2f (new balance: $%.2f)",
		card.CardHolder, req.Amount, card.Balance-req.Amount)

	return models.AuthResponse{
		Approved:   true,
		CardHolder: card.CardHolder,
		LastFour:   utils.LastFour(req.CardUID),
		Timestamp:  time.Now(),
	}
}

// Health returns a lightweight readiness response for monitoring.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("issuer ok"))
}
