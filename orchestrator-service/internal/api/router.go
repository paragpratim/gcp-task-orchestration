package api

import (
	"orchestrator/internal/config"
	"orchestrator/internal/core/bigquery"
	"orchestrator/internal/core/dataflow"
	"orchestrator/internal/core/gcs"
	"orchestrator/internal/core/intake"
	"orchestrator/internal/gcp"
	"orchestrator/internal/routes"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(infra *gcp.Platform, appCfg *config.AppConfig) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	initHandler := intake.NewHandler(intake.NewService(infra.IntakeJobs, infra.JobStatus, infra.CloudTasks, intake.Config{
		JobsCollection:   appCfg.JobsCollection,
		StatusCollection: appCfg.JobStatusCollection,
		IntakeQueueName:  appCfg.IntakeQueueName,
	}))
	gcsHandler := gcs.NewHandler(gcs.NewService(infra.IntakeJobs, infra.JobStatus, infra.CloudTasks, infra.Storage, gcs.Config{
		JobsCollection:   appCfg.JobsCollection,
		StatusCollection: appCfg.JobStatusCollection,
		GcsQueueName:     appCfg.GcsQueueName,
		BqQueueName:      appCfg.BqQueueName,
		IntakeQueueName:  appCfg.IntakeQueueName,
	}))
	bqHandler := bigquery.NewHandler(bigquery.NewService(infra.IntakeJobs, infra.JobStatus, infra.CloudTasks, infra.BigQuery, bigquery.Config{
		JobsCollection:   appCfg.JobsCollection,
		StatusCollection: appCfg.JobStatusCollection,
		GcsQueueName:     appCfg.GcsQueueName,
		BqQueueName:      appCfg.BqQueueName,
	}))
	dfHandler := dataflow.NewHandler(dataflow.NewService(infra.Firestore, infra.CloudTasks))

	v1 := router.Group(routes.APIPrefix)
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
