package main

import (
	"net/http"

	"github.com/franmeza/payment-flow-simulation/cmd/issuer/handlers"
	"github.com/franmeza/payment-flow-simulation/internal/db"
	"github.com/franmeza/payment-flow-simulation/internal/fraud"
	"github.com/franmeza/payment-flow-simulation/internal/httputil"
	"github.com/franmeza/payment-flow-simulation/internal/logger"
	"go.uber.org/zap"
)

func main() {
	logger.Init("issuer")
	defer logger.Log.Sync()

	database := db.New()
	velocity := fraud.NewVelocity()
	issuerHandlers := handlers.New(database, velocity)

	logger.Log.Info("Issuer service starting", zap.String("port", "8082"))
	// Register HTTP routes
	http.HandleFunc("/authorize", issuerHandlers.Authorize)
	http.HandleFunc("/transactions", issuerHandlers.Transactions)
	http.HandleFunc("/stats", issuerHandlers.Stats)
	http.HandleFunc("/health", httputil.HealthHandler("issuer"))

	if err := http.ListenAndServe(":8082", nil); err != nil {
		logger.Log.Fatal("issuer failed to start", zap.Error(err))
	}
}