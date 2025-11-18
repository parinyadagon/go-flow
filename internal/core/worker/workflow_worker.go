package worker

import (
	"context"
	"log"
	"time"

	"github.com/parinyadagon/go-workflow/gen/go_flow/model"
	"github.com/parinyadagon/go-workflow/internal/core/port"
	"github.com/parinyadagon/go-workflow/internal/core/service"
)

type WorkflowWorker struct {
	repo port.WorkflowRepository
}

func NewWorkflowWorker(repo port.WorkflowRepository) *WorkflowWorker {
	return &WorkflowWorker{repo: repo}
}

func (w *WorkflowWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second) // เช็กงานทุก 5 วิ
	defer ticker.Stop()

	log.Println("👷 Worker started: Waiting for jobs...")

	for {
		select {
		case <-ctx.Done(): // สั้งปิด Work
			log.Println("🛑 Worker stopping...")
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *WorkflowWorker) processBatch(ctx context.Context) {
	// 	1. ดึงงาน PENDING มา 10 งาน
	tasks, err := w.repo.GetTaskPending(ctx, 10)
	if err != nil {
		log.Printf("❌ Error fetching tasks: %v", err)
		return
	}

	if len(tasks) == 0 {
		return // ไม่มีงานก็ให้นอนต่อ
	}

	log.Printf("⚡ Found %d jobs! Processing...", len(tasks))

	// 2. รันงาน (Concurrency!)
	for _, task := range tasks {
		// ส่ง job เข้า Go Routine แยก เพื่อให้ทำพร้อมกันได้
		go w.executeTask(ctx, task)
	}
}

func (w *WorkflowWorker) executeTask(ctx context.Context, task model.Tasks) {
	log.Printf("▶️ Doing Task: %s (WID: %s)", task.TaskName, task.WorkflowInstanceID)

	time.Sleep(2 * time.Second)

	// ✅ Task นี้เสร็จแล้ว
	w.repo.UpdateTaskStatus(ctx, int(task.ID), "COMPLETED")

	// 🧠 The Brain Logic: จะไปไหนต่อ?
	w.orchestrateNextStep(ctx, task)
}

func (w *WorkflowWorker) orchestrateNextStep(ctx context.Context, currentTask model.Tasks) {
	// 1. ไปดึงชื่อ Workflow มาก่อน (ต้อง Query join หรือดึงแยก)
	wf, _ := w.repo.GetWorkflowByID(ctx, currentTask.WorkflowInstanceID)

	// 2. ดูลายแทง
	steps := service.WorkflowDefinitions[wf.WorkflowName]

	// 3. หาว่าเราอยู่ Step ไหน
	currentStepIndex := -1
	for i, name := range steps {
		if name == currentTask.TaskName {
			currentStepIndex = i
			break
		}
	}

	// 4. ตัดสินใจ
	if currentStepIndex != -1 && currentStepIndex < len(steps)-1 {
		// 👉 มี Step ถัดไป! สร้าง Task ใหม่รอไว้เลย
		status := model.TasksStatus_Pending
		nextTaskName := steps[currentStepIndex+1]
		log.Printf("➡️ Moving to next step: %s", nextTaskName)

		newTask := &model.Tasks{
			WorkflowInstanceID: currentTask.WorkflowInstanceID,
			TaskName:           nextTaskName,
			Status:             &status,
			InputPayload:       currentTask.OutputPayload,
		}

		w.repo.CreateTask(ctx, newTask)
	} else {
		// 🏁 ไม่มี Step ถัดไปแล้ว -> จบงานใหญ่!
		log.Printf("🎉 Workflow %s COMPLETED!", wf.WorkflowName)
		w.repo.UpdateWorkflowStatus(ctx, wf.ID, "COMPLETED")
	}
}
