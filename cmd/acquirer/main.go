package main

import (
	"net/http"

	"github.com/franmeza/payment-flow-simulation/cmd/acquirer/handlers"
	"github.com/franmeza/payment-flow-simulation/internal/httputil"
	"github.com/franmeza/payment-flow-simulation/internal/logger"
	"go.uber.org/zap"
)

const networkURL = "http://localhost:8081/route"

func main() {
	logger.Init("acquirer")
	defer logger.Log.Sync()

	acquirerHandlers := handlers.New(networkURL, nil)

	logger.Log.Info("Acquirer service starting", zap.String("port", "8080"))

	http.HandleFunc("/authorize", acquirerHandlers.Authorize)
	http.HandleFunc("/health", httputil.HealthHandler("acquirer"))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Log.Fatal("acquirer failed to start", zap.Error(err))
	}
}