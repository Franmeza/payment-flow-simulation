package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/franmeza/payment-flow-simulation/internal/cardutil"
	"github.com/franmeza/payment-flow-simulation/internal/idempotency"
	"github.com/franmeza/payment-flow-simulation/internal/logger"
	"github.com/franmeza/payment-flow-simulation/internal/models"
	"go.uber.org/zap"
)

type Handler struct {
	networkURL  string
	client      *http.Client
	idempotency *idempotency.Store
}

// New builds the acquirer handler with routing dependencies.
func New(networkURL string, client *http.Client, idempotency *idempotency.Store) *Handler {
	if client == nil {
		client = http.DefaultClient
	}

	return &Handler{
		networkURL:  networkURL,
		client:      client,
		idempotency: idempotency,
	}
}

// Authorize validates merchant info and forwards valid requests to network.
func (h *Handler) Authorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Acquirer is the source of transaction IDs for this flow.
	req.TransactionID = fmt.Sprintf("TXN-%d", time.Now().UnixNano())

	log := logger.Log.With(
		zap.String("transaction_id", req.TransactionID),
		zap.String("card_uid", req.CardUID),
		zap.String("merchant_id", req.MerchantID),
		zap.Float64("amount", req.Amount),
	)
	log.Info("Auth request received")

	// If the client supplied an idempotency key and we have already processed
	// an identical request, return the original response without re-charging.
	if req.IdempotencyKey != "" {
		if cached, ok := h.idempotency.Get(req.IdempotencyKey); ok {
			log.Info("Idempotent replay",
				zap.String("idempotency_key", req.IdempotencyKey),
			)
			w.Header().Set("Idempotent-Replayed", "true")
			writeResponse(w, *cached)
			return
		}
	}

	// Simulate acquirer processing time.
	time.Sleep(50 * time.Millisecond)

	// Validate merchant existence and status before routing.
	merchant, exists := merchants[req.MerchantID]
	if !exists {
		log.Warn("Declined: unknown merchant")
		writeResponse(w, models.AuthResponse{
			TransactionID: req.TransactionID,
			Approved:      false,
			DeclineReason: "unknown merchant",
			LastFour:      cardutil.LastFour(req.CardUID),
			Timestamp:     time.Now(),
		})
		return
	}

	if merchant.Status != "active" {
		log.Warn("Declined: merchant blocked")
		writeResponse(w, models.AuthResponse{
			TransactionID: req.TransactionID,
			Approved:      false,
			DeclineReason: "merchant account suspended",
			LastFour:      cardutil.LastFour(req.CardUID),
			Timestamp:     time.Now(),
		})
		return
	}

	resp, err := h.forwardToNetwork(req)
	if err != nil {
		log.Error("Failed to reach network router", zap.Error(err))
		http.Error(w, "network unavailable", http.StatusServiceUnavailable)
		return
	}

	log.Info("Final response", zap.Bool("approved", resp.Approved))

	// Cache the response so any retry with the same key gets this result back.
	if req.IdempotencyKey != "" {
		h.idempotency.Set(req.IdempotencyKey, *resp)
	}

	writeResponse(w, *resp)
}

func (h *Handler) forwardToNetwork(req models.AuthRequest) (*models.AuthResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, h.networkURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := h.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	var resp models.AuthResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func writeResponse(w http.ResponseWriter, resp models.AuthResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
