package httputil

import "net/http"

// HealthHandler returns a shared health endpoint handler for a service.
func HealthHandler(service string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(service + " ok"))
	}
}
