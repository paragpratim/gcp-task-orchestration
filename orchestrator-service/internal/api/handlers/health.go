package handlers

import (
	"net/http"

	"github.com/paragpratim/gcp-task-orchestration/orchestrator-service/internal/logger"
)

// HealthHandler responds to health check requests.
// @Summary      Health check
// @Description  Responds with an OK status if the service is running.
// @Tags         health
// @Produce      plain
// @Success      200  {string}  string  "OK"
// @Router       /health [get]
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		logger.Warn("method not allowed on health endpoint", "method", r.Method, "remote_addr", r.RemoteAddr)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("health check OK", "remote_addr", r.RemoteAddr)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		logger.Error("failed to write health response", "error", err)
	}
}
