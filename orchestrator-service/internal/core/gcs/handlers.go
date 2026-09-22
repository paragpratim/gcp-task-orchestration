package gcs

import (
	"net/http"
	"orchestrator/internal/models"
	"orchestrator/internal/routes"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST(routes.GCSListFiles, h.listFiles)
	rg.POST(routes.GCSMoveFiles, h.moveFiles)
}

// listFiles 	 Handles the initial step of the workflow loop by aggregating file paths matching patterns.
// @Summary      List and Track GCS Files
// @Description  Accepts a minimal task payload, verifies configuration status, aggregates matching files via globs, and triggers the move stage.
// @Tags         gcs
// @Accept       json
// @Produce      json
// @Param        request body      models.PipelineTaskPayload true "Lean Pipeline Transport Payload"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Router       /gcs/files/list [post]
func (h *Handler) listFiles(c *gin.Context) {
	var task models.PipelineTaskPayload

	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task payload schema: " + err.Error()})
		return
	}

	if err := h.service.ListFiles(c.Request.Context(), task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "File discovery complete; proceeding to move files stage"})
}

// moveFiles 	 Archives matching files to a processed directory and queues the downstream BigQuery load job.
// @Summary      Move Files and Queue Ingestion
// @Description  Relocates discovered files to a processed folder boundary, updates metadata structures, and schedules the BigQuery loading phase.
// @Tags         gcs
// @Accept       json
// @Produce      json
// @Param        request body      models.PipelineTaskPayload true "Lean Pipeline Transport Payload"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Router       /gcs/files/move [post]
func (h *Handler) moveFiles(c *gin.Context) {
	var task models.PipelineTaskPayload

	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task payload schema: " + err.Error()})
		return
	}

	if err := h.service.MoveFiles(c.Request.Context(), task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Files archived successfully; BigQuery ingestion task successfully scheduled"})
}
