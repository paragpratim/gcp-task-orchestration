package api

import (
	initpkg "orchestrator/internal/core/init"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(initH *initpkg.Handler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	v1 := router.Group("/api/v1")
	{
		initH.RegisterRoutes(v1)
	}

	// Register Swagger UI handler
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
