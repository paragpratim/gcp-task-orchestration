package gcs

import "github.com/gin-gonic/gin"

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	gcs := rg.Group("/gcs")
	gcs.GET("/files", h.listFiles)
	gcs.POST("/files/move", h.moveFiles)
}

// listFiles 	 Lists the files in a specific GCS bucket.
// @Summary      List GCS files
// @Description  Lists the files in a specific GCS bucket.
// @Tags         gcs
// @Produce      json
// @Param        bucket   query      string  true  "GCS Bucket Name"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Router       /gcs/files [get]
func (h *Handler) listFiles(c *gin.Context) {
	//TODO: Implement the list GCS handler Implement the list GCS handler
}

// moveFiles 	 Moves files within a specific GCS bucket.
// @Summary      Move GCS files
// @Description  Moves files within a specific GCS bucket.
// @Tags         gcs
// @Accept       json
// @Produce      json
// @Param        source_bucket   query      string  true  "Source GCS Bucket Name"
// @Param        destination_bucket   query      string  true  "Destination GCS Bucket Name"
// @Param        files   query      []string  true  "Files to move"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Router       /gcs/files/move [post]
func (h *Handler) moveFiles(c *gin.Context) {
	//TODO: Implement the move GCS handler
}
