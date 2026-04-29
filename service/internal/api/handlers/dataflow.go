package handlers

import (
	"encoding/json"
	"log"
	"net/http"

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

	// For now, we simply log the payload to prove we received it.
	// TODO: Integrate with cloud.google.com/go/dataflow/apiv1beta3 to launch the job on GCP.
	log.Printf("Received Dataflow Request: Action=%s, JobName=%s, ProjectID=%s, Region=%s", req.Action, req.JobName, req.ProjectID, req.Region)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(models.DataflowResponse{
		Message: "Dataflow request received successfully",
	})
}
