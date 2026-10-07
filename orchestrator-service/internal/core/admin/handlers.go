package admin

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"orchestrator/internal/logger"
	"orchestrator/internal/models"
	"orchestrator/internal/routes"

	"github.com/gin-gonic/gin"
)

// Handler 	 Handles HTTP requests for admin job operations.
type Handler struct {
	service *Service
}

// NewHandler 	 Creates a new Handler instance with the provided Service.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes 	 Registers the admin job routes with the provided Gin router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET(routes.AdminHealth, h.appHealth)
	rg.POST(routes.AdminJobCreate, h.createJob)
	rg.GET(routes.AdminJobGet, h.getJob)
	rg.GET(routes.AdminJobGetAll, h.getAllJobs)
	rg.GET(routes.AdminJobsStatus, h.getJobStatus)
	rg.PUT(routes.AdminJobUpdate, h.updateJob)
	rg.DELETE(routes.AdminJobDelete, h.deleteJob)
	rg.POST(routes.AdminJobsQueue, h.queueJobs)
}

// handleJob 	 A generic handler for admin job operations that processes the request and executes the provided action.
func (h *Handler) handleJob(c *gin.Context, successStatus int, action func(c *gin.Context, req models.JobDefinition) (any, error)) {
	var req models.JobDefinition

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind JSON payload", "ERROR", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload schema: " + err.Error()})
		return
	}

	resp, err := action(c, req)
	if err != nil {
		logger.Error("Job execution failed", "ERROR", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(successStatus, resp)
}

// appHealth 	 Responds to health check requests.
// @Summary      Health check
// @Description  Responds with an OK status if the api is running.
// @Tags         health
// @Produce      plain
// @Success      200  {string}  string  "OK"
// @Router       /admin/health [get]
func (h *Handler) appHealth(c *gin.Context) {
	err := h.service.HealthCheck(c)
	if err != nil {
		logger.Error("Health check failed", "ERROR", err)
		c.String(http.StatusInternalServerError, "Service Unavailable")
		return
	}
	c.String(http.StatusOK, "OK")
}

// createJob 	 Handles the creation of a new admin job.
// @Summary      Create Job
// @Description  Accepts a new admin job request and processes it.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request  body      models.JobDefinition  true  "Job Request Payload"
// @Success      201  {object}  map[string]string "Job Created"
// @Failure      400  {object}  map[string]string "Bad Request"
// @Router       /admin/job [post]
func (h *Handler) createJob(c *gin.Context) {
	h.handleJob(c, http.StatusCreated, func(c *gin.Context, req models.JobDefinition) (any, error) {
		return h.service.CreateJob(c, req)
	})
}

// handleJobGet 	 Handles the retrieval of a specific admin job by its ID.
// @Summary      Get Job
// @Description  Retrieves an existing admin job by its ID.
// @Tags         admin
// @Produce      json
// @Param        id   path      string  true  "Job ID"
// @Success      200  {object}  models.JobDefinition "Job Retrieved"
// @Failure      400  {object}  map[string]string "Bad Request"
// @Router       /admin/job/{id} [get]
func (h *Handler) getJob(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		logger.Error("Job retrieval failed: missing job ID", "ERROR", nil)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Job tracking identifier 'id' is required"})
		return
	}

	job, err := h.service.GetJob(c, id)
	if err != nil {
		logger.Error("Job retrieval failed", "ERROR", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, job)
}

// handleJobGetAll 	 Handles the retrieval of all admin jobs.
// @Summary      Get All Jobs
// @Description  Retrieves all existing admin jobs.
// @Tags         admin
// @Produce      json
// @Success      200  {array}   models.JobDefinition "Jobs Retrieved"
// @Failure      400  {object}  map[string]string "Bad Request"
// @Router       /admin/jobs [get]
func (h *Handler) getAllJobs(c *gin.Context) {
	jobs, err := h.service.GetAllJobs(c)
	if err != nil {
		logger.Error("Jobs retrieval failed", "ERROR", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jobs)
}

// updateJob 	 Handles the update of an existing admin job.
// @Summary      Update Job
// @Description  Accepts an update request for an existing admin job and processes it.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request  body      models.JobDefinition  true  "Job Update Payload"
// @Success      200  {object}  map[string]string "Job Updated"
// @Failure      400  {object}  map[string]string "Bad Request"
// @Router       /admin/job [put]
func (h *Handler) updateJob(c *gin.Context) {
	h.handleJob(c, http.StatusOK, func(c *gin.Context, req models.JobDefinition) (any, error) {
		return h.service.UpdateJob(c, req)
	})
}

// deleteJob 	 Handles the deletion of an existing admin job.
// @Summary      Delete Job
// @Description  Deletes an existing admin job by its ID.
// @Tags         admin
// @Produce      json
// @Param        id   path      string  true  "Job ID"
// @Success      200  {object}  map[string]string "Job Deleted"
// @Failure      400  {object}  map[string]string "Bad Request"
// @Router       /admin/job/{id} [delete]
func (h *Handler) deleteJob(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		logger.Error("Job deletion failed: missing job ID", "ERROR", nil)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Job tracking identifier 'id' is required"})
		return
	}

	if err := h.service.DeleteJob(c, id); err != nil {
		logger.Error("Job deletion failed", "ERROR", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Job Deleted", "id": id})
}

// queueJobs 	 Queues all the active admin jobs.
// @Summary      Queue Active Jobs
// @Description  Queues all the active admin jobs for processing.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request  body      models.JobDefinition  true  "Queue Jobs Request Payload"
// @Success      200  {object}  map[string]string "Jobs Queued"
// @Failure      400  {object}  map[string]string "Bad Request"
// @Router       /admin/jobs/queue [post]
func (h *Handler) queueJobs(c *gin.Context) {
	var req models.JobDefinition

	// Read the raw body bytes directly
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Error("Failed to read request body", "ERROR", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	// unmarshal if a JSON payload was actually provided
	trimmedBody := bytes.TrimSpace(bodyBytes)
	if len(trimmedBody) > 0 {
		if err := json.Unmarshal(trimmedBody, &req); err != nil {
			logger.Error("Failed to unmarshal JSON payload for queue filter", "ERROR", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload schema: " + err.Error()})
			return
		}
	}

	// Pass the parsed (or empty) struct to your service layer
	resp, err := h.service.QueueActiveJobs(c, req)
	if err != nil {
		logger.Error("Queue jobs execution failed", "ERROR", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// getJobStatus 	 Handles the retrieval of the status of all admin jobs.
// @Summary      Get Job Statuses
// @Description  Retrieves the status of all admin jobs.
// @Tags         admin
// @Produce      json
// @Success      200  {array}   models.JobStatus "Job Statuses Retrieved"
// @Failure      400  {object}  map[string]string "Bad Request"
// @Router       /admin/jobs/status [get]
func (h *Handler) getJobStatus(c *gin.Context) {
	statuses, err := h.service.GetAllJobStatuses(c)
	if err != nil {
		logger.Error("Job statuses retrieval failed", "ERROR", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, statuses)
}
