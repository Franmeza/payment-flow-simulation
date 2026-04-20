package rules

import "github.com/franmeza/payment-flow-simulation/internal/models"

const maxTransactionAmount = 1000.00

type Result struct {
	Approved bool
	Reason   string
}

func Check(card *models.Card, amount float64) Result {
	if card.Status == "blocked" {
		return Result{false, "card is blocked"}
	}

	if card.Status != "active" {
		return Result{false, "card is not active"}
	}

	if card.Balance < amount {
		return Result{false, "insufficient funds"}
	}

	if amount > maxTransactionAmount {
		return Result{false, "amount exceeds transaction limit"}
	}

	if amount <= 0 {
		return Result{false, "invalid amount"}
	}

	return Result{true, ""}
}