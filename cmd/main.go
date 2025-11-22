package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/parinyadagon/go-workflow/config"
	"github.com/parinyadagon/go-workflow/db"
	repository "github.com/parinyadagon/go-workflow/internal/adapters/driven"
	handler "github.com/parinyadagon/go-workflow/internal/adapters/driving"
	"github.com/parinyadagon/go-workflow/internal/core/service"
	"github.com/parinyadagon/go-workflow/internal/core/worker"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting application in %s mode", cfg.Environment)

	db, err := db.NewConnection(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := repository.NewWorkflowRepository(db)
	svc := service.NewWorkflowService(repo)
	hdl := handler.NewWorkflowHandler(svc)

	workerNode := worker.NewWorkflowWorker(repo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go workerNode.Start(ctx)

	e := echo.New()

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

	log.Println("Gracefully shutting down...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctxShutdown); err != nil {
		e.Logger.Fatal(err)
	}

	log.Println("Server exited")

}
