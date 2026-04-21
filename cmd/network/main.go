package main

import (
	"log"
	"net/http"

	"github.com/franmeza/payment-flow-simulation/cmd/network/handlers"
	"github.com/franmeza/payment-flow-simulation/internal/httputil"
)

const issuerURL = "http://localhost:8082/authorize"

func main() {
	networkHandlers := handlers.New(issuerURL, nil)

	log.Println("Network router starting on :8081")

	http.HandleFunc("/route", networkHandlers.Route)
	http.HandleFunc("/health", httputil.HealthHandler("network"))

	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("network router failed to start: %v", err)
	}
}