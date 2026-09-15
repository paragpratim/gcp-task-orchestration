package init

import "github.com/gin-gonic/gin"

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/init/health", h.AppHealth)
}

// AppHealth 	 Responds to health check requests.
// @Summary      Health check
// @Description  Responds with an OK status if the service is running.
// @Tags         health
// @Produce      plain
// @Success      200  {string}  string  "OK"
// @Router       /health [get]
func (h *Handler) AppHealth(c *gin.Context) {
	h.service.HealthCheck(c.Request.Context())
}
