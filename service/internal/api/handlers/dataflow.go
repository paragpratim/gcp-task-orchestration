package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/paragpratim/gcp-task-orchestration/service/internal/core/dataflow"
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
		log.Printf("Failed to create dataflow service: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	jobID, isRunning, err := dfService.CheckJobActive(req.ProjectID, req.Region, req.JobName)
	if err != nil {
		log.Printf("Failed to check job active status: %v", err)
		http.Error(w, "Failed to check job status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if strings.ToLower(req.Action) == "stop" {
		if !isRunning {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(models.DataflowResponse{
				Message: fmt.Sprintf("Job %s is not running", req.JobName),
			})
			return
		}

		updatedJob, err := dfService.StopJob(req.ProjectID, req.Region, jobID, req.StopMode)
		if err != nil {
			log.Printf("Failed to stop job: %v", err)
			http.Error(w, "Failed to stop job", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(models.DataflowResponse{
			Message: fmt.Sprintf("Successfully requested to %s job %s", req.StopMode, req.JobName),
			JobID:   updatedJob.Id,
		})
		return
	}

	// Default to "start" action
	if isRunning {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.DataflowResponse{
			Message: fmt.Sprintf("Job %s is already running", req.JobName),
			JobID:   jobID,
		})
		return
	}

	resp, err := dfService.LaunchFlexJob(req)
	if err != nil {
		log.Printf("Failed to launch job: %v", err)
		http.Error(w, "Failed to launch job", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(models.DataflowResponse{
		Message: fmt.Sprintf("Successfully launched job %s", req.JobName),
		JobID:   resp.Job.Id,
	})
}
