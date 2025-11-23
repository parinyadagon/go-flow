package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/parinyadagon/go-workflow/config"
	"github.com/parinyadagon/go-workflow/db"
	repository "github.com/parinyadagon/go-workflow/internal/adapters/driven"
	handler "github.com/parinyadagon/go-workflow/internal/adapters/driving"
	"github.com/parinyadagon/go-workflow/internal/core/service"
	"github.com/parinyadagon/go-workflow/internal/core/worker"
	"github.com/parinyadagon/go-workflow/pkg/logger"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize logger
	logger.Init(cfg.Environment)

	logger.Info().Str("environment", cfg.Environment).Msg("Starting application")

	db, err := db.NewConnection(&cfg.Database)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	repo := repository.NewWorkflowRepository(db)
	svc := service.NewWorkflowService(repo)
	hdl := handler.NewWorkflowHandler(svc)

	workerNode := worker.NewWorkflowWorker(
		repo,
		cfg.Worker.PollInterval,
		cfg.Worker.BatchSize,
		cfg.Worker.TaskTimeout,
		cfg.Worker.MaxRetries,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go workerNode.Start(ctx)

	e := echo.New()

	// CORS middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
	}))

	// Health check endpoints
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "ok",
			"service": "go-flow",
		})
	})

	e.GET("/readiness", func(c echo.Context) error {
		// Check database connection
		if err := db.Ping(); err != nil {
			logger.Error().Err(err).Msg("Database ping failed")
			return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
				"status": "unavailable",
				"error":  "database connection failed",
			})
		}
		return c.JSON(http.StatusOK, map[string]string{
			"status":   "ready",
			"database": "connected",
		})
	})

	// Workflow endpoints
	e.GET("/workflows", hdl.ListWorkflows)
	e.POST("/workflows", hdl.StartWorkflow)
	e.GET("/workflows/:id", hdl.GetWorkflowDetail)

	// 4. Start Server
	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}

	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Gracefully shutting down...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctxShutdown); err != nil {
		logger.Fatal().Err(err).Msg("Server shutdown failed")
	}

	logger.Info().Msg("Server exited")

}
