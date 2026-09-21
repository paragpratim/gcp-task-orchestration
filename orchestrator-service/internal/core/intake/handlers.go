package intake

import (
	"net/http"
	"orchestrator/internal/models"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	intake := rg.Group("/intake")
	intake.GET("/health", h.appHealth)
	intake.POST("/job", h.createJob)
	intake.PUT("/job", h.updateJob)
	intake.DELETE("/job/:id", h.deleteJob)
	intake.POST("/jobs/queue", h.queueJobs)
}

// appHealth 	 Responds to health check requests.
// @Summary      Health check
// @Description  Responds with an OK status if the api is running.
// @Tags         health
// @Produce      plain
// @Success      200  {string}  string  "OK"
// @Router       /health [get]
func (h *Handler) appHealth(c *gin.Context) {
	err := h.service.HealthCheck(c)
	if err != nil {
		c.String(http.StatusInternalServerError, "Service Unavailable")
		return
	}
	c.String(http.StatusOK, "OK")
}

// createJob 	 Handles the creation of a new intake job.
// @Summary      Create Intake Job
// @Description  Accepts a new intake job request and processes it.
// @Tags         intake
// @Accept       json
// @Produce      json
// @Param        request  body      models.IntakeJobDefinition  true  "Intake Job Request Payload"
// @Success      201  {object}  map[string]string "Job Created"
// @Failure      400  {object}  map[string]string "Bad Request"
// @Router       /intake/job [post]
func (h *Handler) createJob(c *gin.Context) {
	var req models.IntakeJobDefinition

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload schema: " + err.Error()})
		return
	}

	resp, err := h.service.CreateIntakeJob(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// updateJob 	 Handles the update of an existing intake job.
// @Summary      Update Intake Job
// @Description  Accepts an update request for an existing intake job and processes it.
// @Tags         intake
// @Accept       json
// @Produce      json
// @Param        request  body      models.IntakeJobDefinition  true  "Intake Job Update Payload"
// @Success      200  {object}  map[string]string "Job Updated"
// @Failure      400  {object}  map[string]string "Bad Request"
// @Router       /intake/job [put]
func (h *Handler) updateJob(c *gin.Context) {
	var req models.IntakeJobDefinition

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload schema: " + err.Error()})
		return
	}

	resp, err := h.service.UpdateIntakeJob(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// deleteJob 	 Handles the deletion of an existing intake job.
// @Summary      Delete Intake Job
// @Description  Deletes an existing intake job by its ID.
// @Tags         intake
// @Produce      json
// @Param        id   path      string  true  "Intake Job ID"
// @Success      200  {object}  map[string]string "Job Deleted"
// @Failure      400  {object}  map[string]string "Bad Request"
// @Router       /intake/job/{id} [delete]
func (h *Handler) deleteJob(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Job tracking identifier 'id' is required"})
		return
	}

	if err := h.service.DeleteIntakeJob(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Job Deleted", "id": id})
}

// queueJobs 	 Queues all the active intake jobs.
// @Summary      Queue Active Intake Jobs
// @Description  Queues all the active intake jobs for processing.
// @Tags         intake
// @Accept       json
// @Produce      json
// @Param        request  body      models.IntakeJobDefinition  true  "Queue Jobs Request Payload"
// @Success      200  {object}  map[string]string "Jobs Queued"
// @Failure      400  {object}  map[string]string "Bad Request"
// @Router       /intake/jobs/queue [post]
func (h *Handler) queueJobs(c *gin.Context) {
	var req models.IntakeJobDefinition

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload schema: " + err.Error()})
		return
	}

	resp, err := h.service.QueueActiveJobs(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
