package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
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

	app := fiber.New()

	app.Post("/workflows", hdl.StartWorkflow)

	// 4. Start Server
	log.Println("🚀 Go-Flow Engine starting on :8080")
	log.Fatal(app.Listen(":8080"))

}
