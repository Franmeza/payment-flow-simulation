package main

import (
	"log"
	"net/http"

	"github.com/franmeza/payment-flow-simulation/internal/db"
	"github.com/franmeza/payment-flow-simulation/cmd/issuer/handlers"
)

func main() {
	database := db.New()
	issuerHandlers := handlers.New(database)

	log.Println("Issuer service starting on :8082")
	// Register HTTP routes
	http.HandleFunc("/authorize", issuerHandlers.Authorize)
	http.HandleFunc("/health", issuerHandlers.Health)

	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatalf("issuer failed to start: %v", err)
	}
}