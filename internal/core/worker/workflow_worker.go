package worker

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/parinyadagon/go-workflow/gen/go_flow/model"
	"github.com/parinyadagon/go-workflow/internal/core/port"
	"github.com/parinyadagon/go-workflow/internal/core/service"
	"github.com/parinyadagon/go-workflow/pkg/logger"
)

type WorkflowWorker struct {
	repo         port.WorkflowRepository
	pollInterval time.Duration
	batchSize    int
	taskTimeout  time.Duration
}

func NewWorkflowWorker(repo port.WorkflowRepository, pollInterval time.Duration, batchSize int, taskTimeout time.Duration) *WorkflowWorker {
	return &WorkflowWorker{
		repo:         repo,
		pollInterval: pollInterval,
		batchSize:    batchSize,
		taskTimeout:  taskTimeout,
	}
}

func (w *WorkflowWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	logger.Info().
		Dur("poll_interval", w.pollInterval).
		Int("batch_size", w.batchSize).
		Dur("task_timeout", w.taskTimeout).
		Msg("Worker started: Waiting for jobs...")

	for {
		select {
		case <-ctx.Done(): // สั้งปิด Work
			logger.Info().Msg("Worker stopping...")
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *WorkflowWorker) processBatch(ctx context.Context) {
	// 	1. ดึงงาน PENDING ตาม batch size ที่กำหนด
	tasks, err := w.repo.GetTaskPending(ctx, w.batchSize)
	if err != nil {
		logger.Error().Err(err).Msg("Error fetching tasks")
		return
	}

	if len(tasks) == 0 {
		return // ไม่มีงานก็ให้นอนต่อ
	}

	logger.Info().Int("count", len(tasks)).Msg("Found pending jobs! Processing...")

	// 2. รันงาน (Concurrency!) with WaitGroup
	var wg sync.WaitGroup
	for _, task := range tasks {
		wg.Add(1)
		// ส่ง job เข้า Go Routine แยก เพื่อให้ทำพร้อมกันได้
		go func(t model.Tasks) {
			defer wg.Done()
			w.executeTask(ctx, t)
		}(task)
	}

	// รอให้ทุก goroutine ทำงานเสร็จก่อน return
	wg.Wait()
}

func (w *WorkflowWorker) executeTask(ctx context.Context, task model.Tasks) {
	logger.Info().
		Str("task_name", task.TaskName).
		Str("workflow_id", task.WorkflowInstanceID).
		Int64("task_id", task.ID).
		Msg("Executing task")

	w.repo.UpdateTaskStatus(ctx, int(task.ID), "IN_PROGRESS")

	// Log task start
	eventType := "TASK_STARTED"
	detailsMap := map[string]interface{}{
		"task_id":     task.ID,
		"task_name":   task.TaskName,
		"workflow_id": task.WorkflowInstanceID,
	}
	detailsJSON, err := json.Marshal(detailsMap)
	if err != nil {
		logger.Error().Err(err).Int64("task_id", task.ID).Msg("Failed to marshal task start details")
		return
	}
	details := string(detailsJSON)
	if err := w.repo.CreateActivityLog(ctx, &model.ActivityLogs{
		WorkflowInstanceID: task.WorkflowInstanceID,
		TaskName:           &task.TaskName,
		EventType:          &eventType,
		Details:            &details,
	}); err != nil {
		logger.Error().Err(err).Int64("task_id", task.ID).Msg("Failed to create task start activity log")
	}

	time.Sleep(2 * time.Second)

	time.Sleep(2 * time.Second)

	// ✅ Task นี้เสร็จแล้ว
	w.repo.UpdateTaskStatus(ctx, int(task.ID), "COMPLETED")

	// Log task completion
	eventTypeComplete := "TASK_COMPLETED"
	detailsCompleteMap := map[string]interface{}{
		"task_id":     task.ID,
		"task_name":   task.TaskName,
		"workflow_id": task.WorkflowInstanceID,
		"status":      "success",
	}
	detailsCompleteJSON, err := json.Marshal(detailsCompleteMap)
	if err != nil {
		logger.Error().Err(err).Int64("task_id", task.ID).Msg("Failed to marshal task completion details")
		return
	}
	detailsComplete := string(detailsCompleteJSON)
	if err := w.repo.CreateActivityLog(ctx, &model.ActivityLogs{
		WorkflowInstanceID: task.WorkflowInstanceID,
		TaskName:           &task.TaskName,
		EventType:          &eventTypeComplete,
		Details:            &detailsComplete,
	}); err != nil {
		logger.Error().Err(err).Int64("task_id", task.ID).Msg("Failed to create task completion activity log")
	}

	// 🧠 The Brain Logic: จะไปไหนต่อ?
	w.orchestrateNextStep(ctx, task)
}

func (w *WorkflowWorker) orchestrateNextStep(ctx context.Context, currentTask model.Tasks) {
	// 1. ไปดึงชื่อ Workflow มาก่อน (ต้อง Query join หรือดึงแยก)
	wf, err := w.repo.GetWorkflowByID(ctx, currentTask.WorkflowInstanceID)
	if err != nil {
		logger.Error().Err(err).Str("workflow_id", currentTask.WorkflowInstanceID).Msg("Failed to get workflow")
		return
	}

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
		logger.Info().Str("next_task", nextTaskName).Str("workflow_id", currentTask.WorkflowInstanceID).Msg("Moving to next step")

		newTask := &model.Tasks{
			WorkflowInstanceID: currentTask.WorkflowInstanceID,
			TaskName:           nextTaskName,
			Status:             &status,
			InputPayload:       currentTask.OutputPayload,
		}

		if err := w.repo.CreateTask(ctx, newTask); err != nil {
			logger.Error().Err(err).
				Str("next_task", nextTaskName).
				Str("workflow_id", currentTask.WorkflowInstanceID).
				Msg("Failed to create next task")
			return
		}
	} else {
		// 🏁 ไม่มี Step ถัดไปแล้ว -> จบงานใหญ่!
		logger.Info().Str("workflow_name", wf.WorkflowName).Str("workflow_id", wf.ID).Msg("Workflow COMPLETED!")
		w.repo.UpdateWorkflowStatus(ctx, wf.ID, "COMPLETED")

		// Log workflow completion
		eventType := "WORKFLOW_COMPLETED"
		detailsMap := map[string]interface{}{
			"workflow_id":   wf.ID,
			"workflow_name": wf.WorkflowName,
			"total_tasks":   len(steps),
			"status":        "completed",
		}
		detailsJSON, err := json.Marshal(detailsMap)
		if err != nil {
			logger.Error().Err(err).Str("workflow_id", wf.ID).Msg("Failed to marshal workflow completion details")
			return
		}
		details := string(detailsJSON)
		if err := w.repo.CreateActivityLog(ctx, &model.ActivityLogs{
			WorkflowInstanceID: wf.ID,
			TaskName:           nil,
			EventType:          &eventType,
			Details:            &details,
		}); err != nil {
			logger.Error().Err(err).Str("workflow_id", wf.ID).Msg("Failed to create workflow completion activity log")
		}
	}
}
