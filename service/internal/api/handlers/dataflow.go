package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/paragpratim/gcp-task-orchestration/service/internal/core/dataflow"
	"github.com/paragpratim/gcp-task-orchestration/service/internal/logger"
	"github.com/paragpratim/gcp-task-orchestration/service/internal/models"
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
func SubmitDataflowJobHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Warn("method not allowed on dataflow endpoint", "method", r.Method, "remote_addr", r.RemoteAddr)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.DataflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	dfService, err := dataflow.NewService(ctx)
	if err != nil {
		logger.Error("failed to create dataflow service", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	jobID, isRunning, err := dfService.CheckJobActive(req.ProjectID, req.Region, req.JobName)
	if err != nil {
		logger.Error("failed to check job active status", "error", err, "project_id", req.ProjectID, "job_name", req.JobName)
		http.Error(w, "Failed to check job status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if strings.ToLower(req.Action) == "stop" {
		if !isRunning {
			w.WriteHeader(http.StatusNotFound)
			if err := json.NewEncoder(w).Encode(models.DataflowResponse{
				Message: fmt.Sprintf("Job %s is not running", req.JobName),
			}); err != nil {
				logger.Error("failed to encode response", "error", err)
			}
			return
		}

		updatedJob, err := dfService.StopJob(req.ProjectID, req.Region, jobID, req.StopMode)
		if err != nil {
			logger.Error("failed to stop job", "error", err, "job_id", jobID, "job_name", req.JobName)
			http.Error(w, "Failed to stop job", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		if err := json.NewEncoder(w).Encode(models.DataflowResponse{
			Message: fmt.Sprintf("Successfully requested to %s job %s", req.StopMode, req.JobName),
			JobID:   updatedJob.Id,
		}); err != nil {
			logger.Error("failed to encode response", "error", err)
		}
		return
	}

	// Default to "start" action
	if isRunning {
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(models.DataflowResponse{
			Message: fmt.Sprintf("Job %s is already running", req.JobName),
			JobID:   jobID,
		}); err != nil {
			logger.Error("failed to encode response", "error", err)
		}
		return
	}

	resp, err := dfService.LaunchFlexJob(req)
	if err != nil {
		logger.Error("failed to launch job", "error", err, "job_name", req.JobName, "project_id", req.ProjectID)
		http.Error(w, "Failed to launch job", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(models.DataflowResponse{
		Message: fmt.Sprintf("Successfully launched job %s", req.JobName),
		JobID:   resp.Job.Id,
	}); err != nil {
		logger.Error("failed to encode response", "error", err)
	}
}
