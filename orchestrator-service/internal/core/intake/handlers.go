package intake

import "github.com/gin-gonic/gin"

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
	h.service.HealthCheck(c.Request.Context())
}

// createJob 	 Handles the creation of a new intake job.
// @Summary      Create Intake Job
// @Description  Accepts a new intake job request and processes it.
// @Tags         intake
// @Accept       json
// @Produce      json
// @Param        request  body      intake.IntakeJobRequest  true  "Intake Job Request Payload"
// @Success      201  {object}  map[string]string "Job Created"
// @Failure      400  {object}  map[string]string "Bad Request"
// @Router       /intake/job [post]
func (h *Handler) createJob(c *gin.Context) {
	//TODO: Implement the tasks handler
}

// updateJob 	 Handles the update of an existing intake job.
// @Summary      Update Intake Job
// @Description  Accepts an update request for an existing intake job and processes it.
// @Tags         intake
// @Accept       json
// @Produce      json
// @Param        request  body      intake.IntakeJobRequest  true  "Intake Job Update Payload"
// @Success      200  {object}  map[string]string "Job Updated"
// @Failure      400  {object}  map[string]string "Bad Request"
// @Router       /intake/job [put]
func (h *Handler) updateJob(c *gin.Context) {
	//TODO: Implement the update job handler
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
	//TODO: Implement the delete job handler
}

// queueJobs 	 Queues all the active intake jobs.
// @Summary      Queue Active Intake Jobs
// @Description  Queues all the active intake jobs for processing.
// @Tags         intake
// @Produce      json
// @Success      200  {object}  map[string]string "Jobs Queued"
// @Failure      400  {object}  map[string]string "Bad Request"
// @Router       /intake/jobs/queue [post]
func (h *Handler) queueJobs(c *gin.Context) {
	//TODO: Implement queue job handler
}
