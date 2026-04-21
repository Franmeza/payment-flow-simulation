package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/franmeza/payment-flow-simulation/internal/cardutil"
	"github.com/franmeza/payment-flow-simulation/internal/models"
)

// Simulated merchant database.
var merchants = map[string]models.Merchant{
	"M001": {ID: "M001", Name: "Tim Hortons YYC", Status: "active"},
	"M002": {ID: "M002", Name: "Calgary Co-op", Status: "active"},
	"M003": {ID: "M003", Name: "Blocked Merchant", Status: "blocked"},
}

type Handler struct {
	networkURL string
	client     *http.Client
}

// New builds the acquirer handler with routing dependencies.
func New(networkURL string, client *http.Client) *Handler {
	if client == nil {
		client = http.DefaultClient
	}

	return &Handler{
		networkURL: networkURL,
		client:     client,
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

	log.Printf("[ACQUIRER] Auth request — merchant: %s card: %s amount: $%.2f",
		req.MerchantID, req.CardUID, req.Amount)

	// Simulate acquirer processing time.
	time.Sleep(50 * time.Millisecond)

	// Validate merchant existence and status before routing.
	merchant, exists := merchants[req.MerchantID]
	if !exists {
		log.Printf("[ACQUIRER] DECLINED — unknown merchant: %s", req.MerchantID)
		writeResponse(w, models.AuthResponse{
			Approved:      false,
			DeclineReason: "unknown merchant",
			LastFour:      cardutil.LastFour(req.CardUID),
			Timestamp:     time.Now(),
		})
		return
	}

	if merchant.Status != "active" {
		log.Printf("[ACQUIRER] DECLINED — merchant blocked: %s", req.MerchantID)
		writeResponse(w, models.AuthResponse{
			Approved:      false,
			DeclineReason: "merchant account suspended",
			LastFour:      cardutil.LastFour(req.CardUID),
			Timestamp:     time.Now(),
		})
		return
	}

	resp, err := h.forwardToNetwork(req)
	if err != nil {
		log.Printf("[ACQUIRER] Failed to reach network router: %v", err)
		http.Error(w, "network unavailable", http.StatusServiceUnavailable)
		return
	}

	log.Printf("[ACQUIRER] Final response — approved: %v", resp.Approved)
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
