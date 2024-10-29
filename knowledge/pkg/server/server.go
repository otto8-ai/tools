package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gptscript-ai/knowledge/pkg/datastore"
	"github.com/gptscript-ai/knowledge/pkg/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Config struct {
	ServerURL, Port, APIBase string
}

type Server struct {
	*datastore.Datastore
}

func NewServer(d *datastore.Datastore) *Server {
	return &Server{Datastore: d}
}

// Start starts the server with the given configuration.
func (s *Server) Start(ctx context.Context, cfg Config) error {
	router := gin.Default()

	slog.SetLogLoggerLevel(slog.LevelDebug)

	// Database migration
	if err := s.Index.AutoMigrate(); err != nil {
		return err
	}

	// API routes
	docs.SwaggerInfo.BasePath = cfg.APIBase
	swaggerHandler := ginSwagger.WrapHandler(swaggerFiles.Handler)
	router.GET("/swagger/*any", swaggerHandler) // http://localhost:8080/swagger/index.html

	// @title Knowledge API
	// @version 1
	// @description This is the Knowledge API server for GPTStudio.
	// @contact.name Acorn Labs Inc.
	v1 := router.Group(cfg.APIBase)
	{
		// Swagger >>>
		v1.GET("/docs/*any", swaggerHandler)
		v1.GET("/docs", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, cfg.APIBase+"/docs/index.html")
		})
		// <<< Swagger

		v1Datasets := v1.Group("/datasets")
		{
			v1Datasets.GET("/", s.ListDS)
			v1Datasets.GET("/:id", s.GetDS)
			v1Datasets.POST("/create", s.CreateDS)
			v1Datasets.DELETE("/:id", s.DeleteDS)
			v1Datasets.POST("/:id/ingest", s.IngestIntoDS)
			v1Datasets.POST("/:id/retrieve", s.RetrieveFromDS)
			v1Datasets.DELETE("/:id/documents/:doc_id", s.RemoveDocFromDS)
			v1Datasets.DELETE("/:id/files/:file_id", s.RemoveFileFromDS)
		}
	}

	// Start server
	return router.Run(":" + cfg.Port)
}

// Note: Make sure to implement logic for database operations and handle errors appropriately.
