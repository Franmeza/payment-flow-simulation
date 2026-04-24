package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/franmeza/payment-flow-simulation/internal/logger"
	"github.com/franmeza/payment-flow-simulation/internal/models"
	"go.uber.org/zap"
)

type Handler struct {
	issuerURL string
	client    *http.Client
}

// New builds the network handler with its external routing dependencies.
func New(issuerURL string, client *http.Client) *Handler {
	if client == nil {
		client = http.DefaultClient
	}

	return &Handler{
		issuerURL: issuerURL,
		client:    client,
	}
}

// Route validates the request, forwards it to issuer, and returns issuer JSON.
func (h *Handler) Route(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
	log.Info("Routing request to issuer")

	// Simulate network routing latency
	time.Sleep(80 * time.Millisecond)

	resp, err := h.forwardToIssuer(req)
	if err != nil {
		log.Error("Failed to reach issuer", zap.Error(err))
		http.Error(w, "issuer unreachable", http.StatusServiceUnavailable)
		return
	}

	log.Info("Response received", zap.Bool("approved", resp.Approved))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) forwardToIssuer(req models.AuthRequest) (*models.AuthResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, h.issuerURL, bytes.NewBuffer(body))
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
