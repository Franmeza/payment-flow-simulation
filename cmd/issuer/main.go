package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/franmeza/payment-flow-simulation/internal/db"
	"github.com/franmeza/payment-flow-simulation/internal/rules"
	"github.com/franmeza/payment-flow-simulation/internal/models"
)

var database *db.DB

func main() {
	database = db.New()
	log.Println("Issuer service starting on :8082")

	http.HandleFunc("/authorize", handleAuthorize)
	http.HandleFunc("/health", handleHealth)

	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatalf("issuer failed to start: %v", err)
	}
}

func handleAuthorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("[ISSUER] Auth request — card: %s amount: $%.2f", req.CardUID, req.Amount)

	// Simulate issuer network processing time
	time.Sleep(120 * time.Millisecond)

	resp := processAuth(req)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func processAuth(req models.AuthRequest) models.AuthResponse {
	// Look up card
	card, err := database.GetCard(req.CardUID)
	if err != nil {
		log.Printf("[ISSUER] Unknown card: %s", req.CardUID)
		return models.AuthResponse{
			Approved:      false,
			DeclineReason: "card not found",
			LastFour:      lastFour(req.CardUID),
			Timestamp:     time.Now(),
		}
	}

	// Run rules checks
	result := rules.Check(card, req.Amount)

	// Save transaction record
	tx := models.Transaction{
		ID:         fmt.Sprintf("TXN-%d", time.Now().UnixNano()),
		CardUID:    req.CardUID,
		MerchantID: req.MerchantID,
		Amount:     req.Amount,
		Approved:   result.Approved,
		Reason:     result.Reason,
		Timestamp:  time.Now(),
	}
	database.SaveTransaction(tx)

	if !result.Approved {
		log.Printf("[ISSUER] DECLINED — %s", result.Reason)
		return models.AuthResponse{
			Approved:      false,
			DeclineReason: result.Reason,
			LastFour:      lastFour(req.CardUID),
			Timestamp:     time.Now(),
		}
	}

	// Deduct balance
	database.DeductBalance(req.CardUID, req.Amount)
	log.Printf("[ISSUER] APPROVED — %s $%.2f (new balance: $%.2f)",
		card.CardHolder, req.Amount, card.Balance-req.Amount)

	return models.AuthResponse{
		Approved:   true,
		CardHolder: card.CardHolder,
		LastFour:   lastFour(req.CardUID),
		Timestamp:  time.Now(),
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("issuer ok"))
}

// lastFour returns the last 2 bytes of the UID as a display string
func lastFour(uid string) string {
	if len(uid) >= 5 {
		return uid[len(uid)-5:]
	}
	return uid
}