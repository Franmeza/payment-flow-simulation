package handlers

import "github.com/franmeza/payment-flow-simulation/internal/models"

// merchants is a simulated merchant registry for acquirer validation.
var merchants = map[string]models.Merchant{
	"M001": {ID: "M001", Name: "Kingly", Status: "active"},
	"M002": {ID: "M002", Name: "Kinjo Sushi and Grill", Status: "active"},
	"M003": {ID: "M003", Name: "Blocked Merchant", Status: "blocked"},
}
