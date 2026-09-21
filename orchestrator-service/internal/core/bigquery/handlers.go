package bigquery

import "github.com/gin-gonic/gin"

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	bigquery := rg.Group("/bigquery")
	bigquery.POST("/job", h.createBQJob)
	bigquery.GET("/job/:id", h.getBQJob)
	bigquery.GET("/region", h.checkBQRegion)
}

// createBQJob 	 Creates a new BigQuery job.
// @Summary      Create BigQuery job
// @Description  Creates a new BigQuery job with the provided details.
// @Tags         bigquery
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Router       /bigquery/job [post]
func (h *Handler) createBQJob(c *gin.Context) {
	//TODO: Implement the create BigQuery job handler
}

// getBQJob 	 Retrieves the details of a specific BigQuery job.
// @Summary      Get BigQuery job
// @Description  Retrieves the details of a specific BigQuery job by its ID.
// @Tags         bigquery
// @Produce      json
// @Param        id   path      string  true  "BigQuery Job ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /bigquery/job/{id} [get]
func (h *Handler) getBQJob(c *gin.Context) {
	//TODO: Implement the get BigQuery job handler
}

// checkBQRegion 	 Checks the availability of a specific BigQuery region.
// @Summary      Check BigQuery region
// @Description  Checks the availability of a specific BigQuery region.
// @Tags         bigquery
// @Produce      json
// @Param        region   query      string  true  "BigQuery Region"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Router       /bigquery/region [get]
func (h *Handler) checkBQRegion(c *gin.Context) {
	//TODO: Implement the check BigQuery region handler
}
