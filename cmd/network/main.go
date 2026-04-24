package main

import (
	"net/http"

	"github.com/franmeza/payment-flow-simulation/cmd/network/handlers"
	"github.com/franmeza/payment-flow-simulation/internal/httputil"
	"github.com/franmeza/payment-flow-simulation/internal/logger"
	"go.uber.org/zap"
)

const issuerURL = "http://localhost:8082/authorize"

func main() {
	logger.Init("network")
	defer logger.Log.Sync()

	networkHandlers := handlers.New(issuerURL, nil)

	logger.Log.Info("Network router starting", zap.String("port", "8081"))

	http.HandleFunc("/route", networkHandlers.Route)
	http.HandleFunc("/health", httputil.HealthHandler("network"))

	if err := http.ListenAndServe(":8081", nil); err != nil {
		logger.Log.Fatal("network router failed to start", zap.Error(err))
	}
}