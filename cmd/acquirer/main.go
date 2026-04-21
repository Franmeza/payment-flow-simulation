package main

import (
	"log"
	"net/http"

	"github.com/franmeza/payment-flow-simulation/cmd/acquirer/handlers"
	"github.com/franmeza/payment-flow-simulation/internal/httputil"
)

const networkURL = "http://localhost:8081/route"

func main() {
	acquirerHandlers := handlers.New(networkURL, nil)

	log.Println("Acquirer service starting on :8080")

	http.HandleFunc("/authorize", acquirerHandlers.Authorize)
	http.HandleFunc("/health", httputil.HealthHandler("acquirer"))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("acquirer failed to start: %v", err)
	}
}