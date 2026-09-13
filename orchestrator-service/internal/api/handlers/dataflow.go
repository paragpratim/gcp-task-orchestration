package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/paragpratim/gcp-task-orchestration/orchestrator-service/internal/core/dataflow"
	"github.com/paragpratim/gcp-task-orchestration/orchestrator-service/internal/logger"
	"github.com/paragpratim/gcp-task-orchestration/orchestrator-service/internal/models"
)

// SubmitDataflowJobHandler submits a new Dataflow Flex job.
// @Summary      Submit Dataflow Flex Job
// @Description  Accepts parameters to start or stop a Dataflow Flex job via the orchestrator.
// @Tags         dataflow
// @Accept       json
// @Produce      json
// @Param        request  body      models.DataflowRequest  true  "Dataflow Job Parameters"
// @Success      202   {object}  models.DataflowResponse
// @Failure      400   {string}  string "Bad Request"
// @Router       /dataflow/flex/submit [post]
func SubmitDataflowJobHandler(c *gin.Context) {
	var req models.DataflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	ctx := c.Request.Context()
	dfService, err := dataflow.NewService(ctx)
	if err != nil {
		logger.Error("failed to create dataflow service", "error", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}

	jobID, isRunning, err := dfService.CheckJobActive(req.ProjectID, req.Region, req.JobName)
	if err != nil {
		logger.Error("failed to check job active status", "error", err, "project_id", req.ProjectID, "job_name", req.JobName)
		c.String(http.StatusInternalServerError, "Failed to check job status")
		return
	}

	if strings.ToLower(req.Action) == "stop" {
		if !isRunning {
			c.JSON(http.StatusNotFound, models.DataflowResponse{
				Message: fmt.Sprintf("Job %s is not running", req.JobName),
			})
			return
		}

		updatedJob, err := dfService.StopJob(req.ProjectID, req.Region, jobID, req.StopMode)
		if err != nil {
			logger.Error("failed to stop job", "error", err, "job_id", jobID, "job_name", req.JobName)
			c.String(http.StatusInternalServerError, "Failed to stop job")
			return
		}

		c.JSON(http.StatusAccepted, models.DataflowResponse{
			Message: fmt.Sprintf("Successfully requested to %s job %s", req.StopMode, req.JobName),
			JobID:   updatedJob.Id,
		})
		return
	}

	// Default to "start" action
	if isRunning {
		c.JSON(http.StatusOK, models.DataflowResponse{
			Message: fmt.Sprintf("Job %s is already running", req.JobName),
			JobID:   jobID,
		})
		return
	}

	resp, err := dfService.LaunchFlexJob(req)
	if err != nil {
		logger.Error("failed to launch job", "error", err, "job_name", req.JobName, "project_id", req.ProjectID)
		c.String(http.StatusInternalServerError, "Failed to launch job")
		return
	}

	c.JSON(http.StatusAccepted, models.DataflowResponse{
		Message: fmt.Sprintf("Successfully launched job %s", req.JobName),
		JobID:   resp.Job.Id,
	})
}
