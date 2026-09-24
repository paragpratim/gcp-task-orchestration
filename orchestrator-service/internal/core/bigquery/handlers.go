package bigquery

import (
	"context"
	"net/http"
	"orchestrator/internal/logger"
	"orchestrator/internal/models"
	"orchestrator/internal/routes"

	"github.com/gin-gonic/gin"
)

// Handler struct to manage BigQuery related HTTP requests.
type Handler struct {
	service *Service
}

// NewHandler creates a new instance of Handler with the provided service.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers the BigQuery related routes with the provided router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST(routes.BigQueryLoadJobCreate, h.createBQJob)
	rg.POST(routes.BigQueryLoadJobCheck, h.getBQJob)
	rg.POST(routes.BigQueryRegionCheck, h.checkBQRegion)
}

// handleTask is a helper method to process a PipelineTaskPayload and execute the provided action.
func (h *Handler) handleTask(c *gin.Context, successMsg string, action func(ctx context.Context, task models.PipelineTaskPayload) error) {
	var task models.PipelineTaskPayload

	if err := c.ShouldBindJSON(&task); err != nil {
		logger.Error("Failed to bind JSON payload", "ERROR", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task payload schema: " + err.Error()})
		return
	}

	if err := action(c.Request.Context(), task); err != nil {
		logger.Error("Task Execution failed", "ERROR", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": successMsg})
}

// createBQJob 	 Creates a new BigQuery job.
// @Summary      Create BigQuery job
// @Description  Creates a new BigQuery job with the provided details.
// @Tags         bigquery
// @Accept       json
// @Produce      json
// @Param        request body      models.PipelineTaskPayload true "Lean Pipeline Transport Payload"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Router       /bigquery/job/create [post]
func (h *Handler) createBQJob(c *gin.Context) {
	h.handleTask(c, "BigQuery load job creation triggered", h.service.CreateLoadJob)
}

// getBQJob 	 Retrieves the details of a specific BigQuery job.
// @Summary      Get BigQuery job
// @Description  Retrieves the details of a specific BigQuery job by its ID.
// @Tags         bigquery
// @Produce      json
// @Param        request body      models.PipelineTaskPayload true "Lean Pipeline Transport Payload"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /bigquery/job/check [post]
func (h *Handler) getBQJob(c *gin.Context) {
	h.handleTask(c, "BigQuery load job status check complete", h.service.CheckLoadJobStatus)
}

// checkBQRegion 	 Checks the availability of a specific BigQuery region.
// @Summary      Check BigQuery region
// @Description  Checks the availability of a specific BigQuery region.
// @Tags         bigquery
// @Produce      json
// @Param        request body      models.PipelineTaskPayload true "Lean Pipeline Transport Payload"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Router       /bigquery/region/check [post]
func (h *Handler) checkBQRegion(c *gin.Context) {
	h.handleTask(c, "BigQuery dataset region check complete", h.service.CheckDatasetRegion)
}
