package api

import (
	"orchestrator/internal/config"
	"orchestrator/internal/core/admin"
	"orchestrator/internal/core/bigquery"
	"orchestrator/internal/core/dataflow"
	"orchestrator/internal/core/gcs"
	"orchestrator/internal/gcp"
	"orchestrator/internal/routes"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter initializes the core Gin routing engine with recovery middleware,
// global telemetry hooks, and explicit visual Swagger UI documentation.
//
// It compiles and cross-injects infrastructure tiers down into individual
// feature controllers (Admin, GCS, BigQuery, Dataflow) using pure, compile-safe
// constructor dependency injection.
func SetupRouter(infra *gcp.Platform, appCfg *config.AppConfig) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	// Add CORS middleware to handle cross-origin requests
	router.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Initialize feature handlers with their respective services and configurations
	initHandler := admin.NewHandler(admin.NewService(infra.Jobs, infra.JobStatus, infra.JobStatusLog, infra.CloudTasks, admin.Config{
		JobsCollection:   appCfg.JobsCollection,
		StatusCollection: appCfg.JobStatusCollection,
		LogCollection:    appCfg.JobStatusLogCollection,
		AdminQueueName:   appCfg.AdminQueueName,
	}))
	gcsHandler := gcs.NewHandler(gcs.NewService(infra.Jobs, infra.JobStatus, infra.JobStatusLog, infra.CloudTasks, infra.Storage, gcs.Config{
		JobsCollection:       appCfg.JobsCollection,
		StatusCollection:     appCfg.JobStatusCollection,
		LogCollection:        appCfg.JobStatusLogCollection,
		GcsQueueName:         appCfg.GcsQueueName,
		BqQueueName:          appCfg.BqQueueName,
		AdminQueueName:       appCfg.AdminQueueName,
		TaskFrequency:        appCfg.TaskFrequency,
		StatusCheckerFrequency: appCfg.StatusCheckerFrequency,
	}))
	bqHandler := bigquery.NewHandler(bigquery.NewService(infra.Jobs, infra.JobStatus, infra.JobStatusLog, infra.CloudTasks, infra.BigQuery, bigquery.Config{
		JobsCollection:         appCfg.JobsCollection,
		StatusCollection:       appCfg.JobStatusCollection,
		LogCollection:          appCfg.JobStatusLogCollection,
		GcsQueueName:           appCfg.GcsQueueName,
		BqQueueName:            appCfg.BqQueueName,
		BQLoadJobCheckFrequency: appCfg.BQLoadJobCheckFrequency,
		StatusCheckerFrequency:  appCfg.StatusCheckerFrequency,
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
