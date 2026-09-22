package api

import (
	"orchestrator/internal/config"
	"orchestrator/internal/core/bigquery"
	"orchestrator/internal/core/dataflow"
	"orchestrator/internal/core/gcs"
	"orchestrator/internal/core/intake"
	"orchestrator/internal/gcp"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(infra *gcp.Platform, appCfg *config.AppConfig) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	initHandler := intake.NewHandler(intake.NewService(infra.Firestore, infra.CloudTasks, intake.Config{
		JobsCollection:   appCfg.IntakeJobsCollection,
		StatusCollection: appCfg.JobStatusCollection,
		QueueName:        appCfg.IntakeQueueName,
	}))
	gcsHandler := gcs.NewHandler(gcs.NewService(infra.Firestore, infra.CloudTasks, infra.Storage))
	bqHandler := bigquery.NewHandler(bigquery.NewService(infra.Firestore, infra.CloudTasks, infra.BigQuery))
	dfHandler := dataflow.NewHandler(dataflow.NewService(infra.Firestore, infra.CloudTasks))

	v1 := router.Group("/api/v1")
	{
		initHandler.RegisterRoutes(v1)
		gcsHandler.RegisterRoutes(v1)
		bqHandler.RegisterRoutes(v1)
		dfHandler.RegisterRoutes(v1)
	}

	// Register Swagger UI handler
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
