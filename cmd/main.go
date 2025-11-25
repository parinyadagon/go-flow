package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/parinyadagon/go-workflow/config"
	"github.com/parinyadagon/go-workflow/db"
	"github.com/parinyadagon/go-workflow/gen/go_flow/model"
	repository "github.com/parinyadagon/go-workflow/internal/adapters/driven"
	handler "github.com/parinyadagon/go-workflow/internal/adapters/driving"
	"github.com/parinyadagon/go-workflow/internal/core/registry"
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

	// สร้าง registry และ define workflows แบบ inline
	workflowRegistry := registry.NewWorkflowRegistry()

	// ===================================
	// Define OrderProcess Workflow
	// ===================================
	workflowRegistry.NewWorkflow("OrderProcess").
		AddTask("ValidateOrder", func(ctx context.Context, task *model.Tasks) error {
			logger.Info().Str("task", "ValidateOrder").Msg("Validating order")

			var input map[string]interface{}
			if task.InputPayload != nil {
				json.Unmarshal([]byte(*task.InputPayload), &input)
			}

			time.Sleep(1 * time.Second)

			// Validation logic
			if orderID, ok := input["order_id"].(string); ok && orderID == "" {
				return errors.New("order_id is required")
			}
			if amount, ok := input["amount"].(float64); ok && amount <= 0 {
				return errors.New("amount must be positive")
			}

			// Random failure for testing (30%)
			if time.Now().Unix()%10 < 3 {
				return errors.New("validation failed randomly")
			}

			output := map[string]interface{}{
				"validated":    true,
				"order_id":     input["order_id"],
				"amount":       input["amount"],
				"validated_at": time.Now().Format(time.RFC3339),
			}
			outputJSON, _ := json.Marshal(output)
			outputStr := string(outputJSON)
			task.OutputPayload = &outputStr

			return nil
		}).
		AddTask("DeductMoney", func(ctx context.Context, task *model.Tasks) error {
			logger.Info().Str("task", "DeductMoney").Msg("Deducting money")

			var input map[string]interface{}
			if task.InputPayload != nil {
				json.Unmarshal([]byte(*task.InputPayload), &input)
			}

			time.Sleep(2 * time.Second)

			// Random failure (20%)
			if time.Now().Unix()%10 < 2 {
				return errors.New("payment gateway timeout")
			}

			output := map[string]interface{}{
				"payment_status": "SUCCESS",
				"transaction_id": "TXN" + time.Now().Format("20060102150405"),
				"amount":         input["amount"],
				"deducted_at":    time.Now().Format(time.RFC3339),
			}
			outputJSON, _ := json.Marshal(output)
			outputStr := string(outputJSON)
			task.OutputPayload = &outputStr

			return nil
		}).
		AddTask("SendEmail", func(ctx context.Context, task *model.Tasks) error {
			logger.Info().Str("task", "SendEmail").Msg("Sending email")

			var input map[string]interface{}
			if task.InputPayload != nil {
				json.Unmarshal([]byte(*task.InputPayload), &input)
			}

			time.Sleep(1 * time.Second)

			output := map[string]interface{}{
				"email_sent":  true,
				"recipient":   "customer@example.com",
				"sent_at":     time.Now().Format(time.RFC3339),
				"order_id":    input["order_id"],
				"transaction": input["transaction_id"],
			}
			outputJSON, _ := json.Marshal(output)
			outputStr := string(outputJSON)
			task.OutputPayload = &outputStr

			return nil
		}).
		MustBuild()

	// =============================================
	// Define RefundProcess Workflow (ตัวอย่างเพิ่ม)
	// =============================================
	workflowRegistry.NewWorkflow("RefundProcess").
		AddTask("ValidateRefund", func(ctx context.Context, task *model.Tasks) error {
			logger.Info().Str("task", "ValidateRefund").Msg("Validating refund request")
			time.Sleep(1 * time.Second)
			return nil
		}).
		AddTask("ProcessRefund", func(ctx context.Context, task *model.Tasks) error {
			logger.Info().Str("task", "ProcessRefund").Msg("Processing refund")
			time.Sleep(2 * time.Second)
			return nil
		}).
		AddTask("NotifyCustomer", func(ctx context.Context, task *model.Tasks) error {
			logger.Info().Str("task", "NotifyCustomer").Msg("Notifying customer")
			time.Sleep(1 * time.Second)
			return nil
		}).
		MustBuild()

	repo := repository.NewWorkflowRepository(db)
	svc := service.NewWorkflowService(repo, workflowRegistry)
	hdl := handler.NewWorkflowHandler(svc)

	workerNode := worker.NewWorkflowWorker(repo, workflowRegistry, &cfg.Worker)

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
	e.GET("/workflows/available", hdl.ListAvailableWorkflows)
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
