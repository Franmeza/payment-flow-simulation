package models

import "time"

// Card represents a card in the issuer database
type Card struct {
	UID        string  `json:"uid"`
	CardHolder string  `json:"card_holder"`
	Balance    float64 `json:"balance"`
	Status     string  `json:"status"` // "active", "blocked", "insufficient_funds"
}

// Merchant represents a registered merchant
type Merchant struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"` // "active", "blocked"
}

// AuthRequest is what the ESP32 terminal sends to the acquirer
type AuthRequest struct {
	TransactionID string  `json:"transaction_id"`
	CardUID    string  `json:"card_uid"`
	MerchantID string  `json:"merchant_id"`
	Amount     float64 `json:"amount"`
}

// AuthResponse is what travels back to the terminal
type AuthResponse struct {
	TransactionID string  `json:"transaction_id"`
	Approved      bool      `json:"approved"`
	DeclineReason string    `json:"decline_reason,omitempty"`
	CardHolder    string    `json:"card_holder,omitempty"`
	LastFour      string    `json:"last_four"`
	Timestamp     time.Time `json:"timestamp"`
}

// Transaction is a record of every auth attempt
type Transaction struct {
	ID         string    `json:"id"`
	CardUID    string    `json:"card_uid"`
	MerchantID string    `json:"merchant_id"`
	Amount     float64   `json:"amount"`
	Approved   bool      `json:"approved"`
	Reason     string    `json:"reason,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}