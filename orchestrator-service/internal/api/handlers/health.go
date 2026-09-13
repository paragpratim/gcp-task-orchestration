package handlers

import (
	"net/http"
	"orchestrator/internal/logger"

	"github.com/gin-gonic/gin"
)

// HealthHandler responds to health check requests.
// @Summary      Health check
// @Description  Responds with an OK status if the service is running.
// @Tags         health
// @Produce      plain
// @Success      200  {string}  string  "OK"
// @Router       /health [get]
func HealthHandler(c *gin.Context) {
	logger.Debug("health check OK", "remote_addr", c.ClientIP())
	c.String(http.StatusOK, "OK")
}
