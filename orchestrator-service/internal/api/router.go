package api

import (
	"orchestrator/internal/api/handlers"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewRouter sets up the HTTP routes for the service.
func NewRouter() *gin.Engine {
	router := gin.Default()

	// Register handlers
	router.GET("/health", handlers.HealthHandler)
	router.POST("/dataflow/flex/submit", handlers.SubmitDataflowJobHandler)

	// Register Swagger UI handler
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
