package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/franmeza/payment-flow-simulation/internal/cardutil"
	"github.com/franmeza/payment-flow-simulation/internal/db"
	"github.com/franmeza/payment-flow-simulation/internal/fraud"
	"github.com/franmeza/payment-flow-simulation/internal/logger"
	"github.com/franmeza/payment-flow-simulation/internal/models"
	"github.com/franmeza/payment-flow-simulation/internal/rules"
	"go.uber.org/zap"
)

type Handler struct {
	database *db.DB
	velocity *fraud.Velocity
}

// New builds a Handler with its required dependencies.
func New(database *db.DB, velocity *fraud.Velocity) *Handler {
	return &Handler{database: database, velocity: velocity}
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

	log := logger.Log.With(
		zap.String("transaction_id", req.TransactionID),
		zap.String("card_uid", req.CardUID),
		zap.Float64("amount", req.Amount),
	)
	log.Info("Auth request received")

	// Simulate issuer-side processing latency.
	time.Sleep(120 * time.Millisecond)

	resp := h.processAuth(req, log)

	// Always return a JSON response payload.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// processAuth applies issuer rules and persists the auth attempt result.
func (h *Handler) processAuth(req models.AuthRequest, log *zap.Logger) models.AuthResponse {
	// Fetch card details needed for validation and balance checks.
	card, err := h.database.GetCard(req.CardUID)
	if err != nil {
		log.Warn("Declined: card not found")
		return models.AuthResponse{
			TransactionID: req.TransactionID,
			Approved:      false,
			DeclineReason: "card not found",
			LastFour:      cardutil.LastFour(req.CardUID),
			Timestamp:     time.Now(),
		}
	}

	// Resolve the transaction ID early so all decline paths can reference it.
	transactionID := req.TransactionID
	if transactionID == "" {
		transactionID = fmt.Sprintf("TXN-%d", time.Now().UnixNano())
	}

	// Velocity fraud check: decline if the card fires 3+ transactions within 60 s.
	if h.velocity.Check(req.CardUID) {
		log.Warn("Declined: velocity fraud rule triggered",
			zap.String("card_uid", req.CardUID),
		)
		tx := models.Transaction{
			ID:         transactionID,
			CardUID:    req.CardUID,
			MerchantID: req.MerchantID,
			Amount:     req.Amount,
			Approved:   false,
			Reason:     "velocity fraud rule triggered",
			Timestamp:  time.Now(),
		}
		_ = h.database.SaveTransaction(tx)
		return models.AuthResponse{
			TransactionID: transactionID,
			Approved:      false,
			DeclineReason: "velocity fraud rule triggered",
			LastFour:      cardutil.LastFour(req.CardUID),
			Timestamp:     time.Now(),
		}
	}

	// Run rules checks
	result := rules.Check(card, req.Amount)

	tx := models.Transaction{
		ID:         transactionID,
		CardUID:    req.CardUID,
		MerchantID: req.MerchantID,
		Amount:     req.Amount,
		Approved:   result.Approved,
		Reason:     result.Reason,
		Timestamp:  time.Now(),
	}
	if err := h.database.SaveTransaction(tx); err != nil {
		log.Error("Failed to save transaction", zap.Error(err))
		return models.AuthResponse{
			TransactionID: transactionID,
			Approved:      false,
			DeclineReason: "issuer database error",
			LastFour:      cardutil.LastFour(req.CardUID),
			Timestamp:     time.Now(),
		}
	}

	// Return early for declined authorizations after recording the attempt.
	if !result.Approved {
		log.Warn("Declined", zap.String("reason", result.Reason))
		return models.AuthResponse{
			TransactionID: transactionID,
			Approved:      false,
			DeclineReason: result.Reason,
			LastFour:      cardutil.LastFour(req.CardUID),
			Timestamp:     time.Now(),
		}
	}

	// For approved requests, deduct funds from the card balance.
	if err := h.database.DeductBalance(req.CardUID, req.Amount); err != nil {
		log.Error("Failed to deduct balance", zap.Error(err))
		return models.AuthResponse{
			TransactionID: transactionID,
			Approved:      false,
			DeclineReason: "issuer database error",
			LastFour:      cardutil.LastFour(req.CardUID),
			Timestamp:     time.Now(),
		}
	}

	log.Info("Approved",
		zap.String("card_holder", card.CardHolder),
		zap.Float64("amount", req.Amount),
		zap.Float64("new_balance", card.Balance-req.Amount),
	)

	return models.AuthResponse{
		TransactionID: transactionID,
		Approved:      true,
		CardHolder:    card.CardHolder,
		LastFour:      cardutil.LastFour(req.CardUID),
		Timestamp:     time.Now(),
	}
}
