package worker

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/parinyadagon/go-workflow/gen/go_flow/model"
	"github.com/parinyadagon/go-workflow/internal/core/port"
	"github.com/parinyadagon/go-workflow/internal/core/registry"
	"github.com/parinyadagon/go-workflow/pkg/logger"
)

type WorkflowWorker struct {
	repo         port.WorkflowRepository
	registry     *registry.WorkflowRegistry
	pollInterval time.Duration
	batchSize    int
	taskTimeout  time.Duration
	maxRetries   int
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
	// Get retry count (default to 0 if nil)
	retryCount := int32(0)
	if task.RetryCount != nil {
		retryCount = *task.RetryCount
	}

	logger.Info().
		Str("task_name", task.TaskName).
		Str("workflow_id", task.WorkflowInstanceID).
		Int64("task_id", task.ID).
		Int32("retry_count", retryCount).
		Msg("Executing task")

	// Update status to IN_PROGRESS or RETRYING
	status := "IN_PROGRESS"
	if retryCount > 0 {
		status = "RETRYING"
	}
	w.repo.UpdateTaskStatus(ctx, int(task.ID), status)

	// Log task start
	eventType := "TASK_STARTED"
	detailsMap := map[string]interface{}{
		"task_id":     task.ID,
		"task_name":   task.TaskName,
		"workflow_id": task.WorkflowInstanceID,
		"retry_count": retryCount,
	}
	detailsJSON, err := json.Marshal(detailsMap)
	if err != nil {
		logger.Error().Err(err).Int64("task_id", task.ID).Msg("Failed to marshal task start details")
		return
	}

	details := string(detailsJSON)
	err = w.repo.CreateActivityLog(ctx, &model.ActivityLogs{
		WorkflowInstanceID: task.WorkflowInstanceID,
		TaskName:           &task.TaskName,
		EventType:          &eventType,
		Details:            &details,
	})
	if err != nil {
		logger.Error().Err(err).Int64("task_id", task.ID).Msg("Failed to create task start activity log")
	}

	// ดึง workflow definition
	wf, err := w.repo.GetWorkflowByID(ctx, task.WorkflowInstanceID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get workflow")
		w.handleTaskFailure(ctx, task, retryCount, err)
		return
	}

	// ดึง task function จาก registry
	taskFunc, exists := w.registry.GetTaskFunc(wf.WorkflowName, task.TaskName)
	if !exists {
		err := errors.New("task function not found: " + task.TaskName)
		logger.Error().Err(err).Str("task_name", task.TaskName).Msg("No task function registered")
		w.handleTaskFailure(ctx, task, retryCount, err)
		return
	}

	// Execute with timeout
	execCtx, cancel := context.WithTimeout(ctx, w.taskTimeout)
	defer cancel()

	err = taskFunc(execCtx, &task)
	if err != nil {
		logger.Error().Err(err).
			Str("task_name", task.TaskName).
			Int64("task_id", task.ID).
			Msg("Task execution failed")
		w.handleTaskFailure(ctx, task, retryCount, err)
		return
	}

	// Task succeeded
	w.handleTaskSuccess(ctx, task, retryCount)
}

func (w *WorkflowWorker) orchestrateNextStep(ctx context.Context, currentTask model.Tasks) {
	// 1. ไปดึงชื่อ Workflow มาก่อน (ต้อง Query join หรือดึงแยก)
	wf, err := w.repo.GetWorkflowByID(ctx, currentTask.WorkflowInstanceID)
	if err != nil {
		logger.Error().Err(err).Str("workflow_id", currentTask.WorkflowInstanceID).Msg("Failed to get workflow")
		return
	}

	def, exists := w.registry.GetDefinition(wf.WorkflowName)
	if !exists {
		logger.Error().Str("workflow_name", wf.WorkflowName).Msg("Workflow definition not found")
		return
	}

	// 3. หาว่าเราอยู่ Step ไหน
	currentStepIndex := -1
	for i, name := range def.TaskNames {
		if name == currentTask.TaskName {
			currentStepIndex = i
			break
		}
	}

	// 4. ตัดสินใจ
	if currentStepIndex != -1 && currentStepIndex < len(def.TaskNames)-1 {
		// 👉 มี Step ถัดไป! สร้าง Task ใหม่รอไว้เลย
		status := model.TasksStatus_Pending
		nextTaskName := def.TaskNames[currentStepIndex+1]
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
			"total_tasks":   len(def.TaskNames),
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

// handlerTaskFailure handles task failure with retry logic
func (w *WorkflowWorker) handleTaskFailure(ctx context.Context, task model.Tasks, retryCount int32, taskErr error) {
	// Check if we should retry
	if retryCount >= int32(w.maxRetries) {
		// Max retries reached - mark as FAILED
		logger.Warn().
			Int64("task_id", task.ID).
			Int32("retry_count", retryCount).
			Msg("Task failed after max retries")

		w.repo.UpdateTaskStatus(ctx, int(task.ID), "FAILED")

		// Log failure in activity logs
		eventTypeFailed := "TASK_FAILED"
		failureDefaults := map[string]any{
			"task_id":     task.ID,
			"task_name":   task.TaskName,
			"retry_count": retryCount,
			"reason":      "Max retries exceeded",
			"error":       taskErr.Error(),
		}
		failureDefaultsJSON, _ := json.Marshal(failureDefaults)
		failureDefaultsStr := string(failureDefaultsJSON)
		w.repo.CreateActivityLog(ctx, &model.ActivityLogs{
			WorkflowInstanceID: task.WorkflowInstanceID,
			TaskName:           &task.TaskName,
			EventType:          &eventTypeFailed,
			Details:            &failureDefaultsStr,
		})

		// Mark workflow as FAILED
		w.repo.UpdateWorkflowStatus(ctx, task.WorkflowInstanceID, "FAILED")

		return
	}

	// Increment retry count and schedule retry
	newRetryCount := int(retryCount) + 1
	if err := w.repo.UpdateTaskRetryCount(ctx, int(task.ID), newRetryCount); err != nil {
		logger.Error().Err(err).Int64("task_id", task.ID).Msg("Failed to update retry count")
		return
	}

	// Calculate exponential backoff delay (2^retryCount second)
	backoffDelay := time.Duration(1<<uint(newRetryCount)) * time.Second
	logger.Info().
		Int64("task_id", task.ID).
		Int("retry_count", newRetryCount).
		Dur("backoff_delay", backoffDelay).
		Str("error", taskErr.Error()).
		Msg("Task failed, scheduling retry with exponential backoff")

	// Update to FAILED status temporarily
	w.repo.UpdateTaskStatus(ctx, int(task.ID), "FAILED")

	// Log retry in activity logs
	eventTypeRetry := "TASK_RETRY"
	retryDetails := map[string]any{
		"task_id":       task.ID,
		"task_name":     task.TaskName,
		"retry_count":   newRetryCount,
		"backoff_delay": backoffDelay.String(),
		"error":         taskErr.Error(),
	}
	retryDetailsJSON, _ := json.Marshal(retryDetails)
	retryDetailsStr := string(retryDetailsJSON)
	w.repo.CreateActivityLog(ctx, &model.ActivityLogs{
		WorkflowInstanceID: task.WorkflowInstanceID,
		TaskName:           &task.TaskName,
		EventType:          &eventTypeRetry,
		Details:            &retryDetailsStr,
	})

	// Sleep for exponential backoff
	time.Sleep(backoffDelay)

	// Reset to PENDING so worker can pick it up again
	w.repo.UpdateTaskStatus(ctx, int(task.ID), "PENDING")
}

// handleTaskSuccess handles successfully task completion
func (w *WorkflowWorker) handleTaskSuccess(ctx context.Context, task model.Tasks, retryCount int32) {
	// Update task output payload if executor modified it
	if task.OutputPayload != nil {
		// Optionally save output payload to database
		// w.repo.UpdateTaskOutput(ctx, int(task.ID), *task.OutputPayload)
	}

	// Mark task as COMPLETED
	w.repo.UpdateTaskStatus(ctx, int(task.ID), "COMPLETED")

	// Log task completion
	eventTypeComplete := "TASK_COMPLETED"
	detailsCompleteMap := map[string]any{
		"task_id":     task.ID,
		"task_name":   task.TaskName,
		"workflow_id": task.WorkflowInstanceID,
		"status":      "success",
		"retry_count": retryCount,
	}
	detailsCompleteJSON, _ := json.Marshal(detailsCompleteMap)
	detailsComplete := string(detailsCompleteJSON)
	w.repo.CreateActivityLog(ctx, &model.ActivityLogs{
		WorkflowInstanceID: task.WorkflowInstanceID,
		TaskName:           &task.TaskName,
		EventType:          &eventTypeComplete,
		Details:            &detailsComplete,
	})

	// Orchestrate next step
	w.orchestrateNextStep(ctx, task)
}

// =================================
// Builder Pattern
// =================================

// WorkerBuilder providers a fluent API for constructing WorkflowWorker
type WorkBuilder struct {
	repo         port.WorkflowRepository
	registry     *registry.WorkflowRegistry
	pollInterval time.Duration
	batchSize    int
	taskTimeout  time.Duration
	maxRetries   int
}

// NewWorkerBuilder creates a new WorkerBuilder with default values
func NewWorkerBuilder(repo port.WorkflowRepository, reg *registry.WorkflowRegistry) *WorkBuilder {
	return &WorkBuilder{
		repo:         repo,
		registry:     reg,
		pollInterval: 5 * time.Second,
		batchSize:    10,
		taskTimeout:  30 * time.Second,
		maxRetries:   3,
	}
}

// WithPollInterval sets the polling interval
func (b *WorkBuilder) WithPollInterval(d time.Duration) *WorkBuilder {
	b.pollInterval = d

	return b
}

// WithBatchSize sets the batch size for task processing
func (b *WorkBuilder) WithBatchSize(size int) *WorkBuilder {
	b.batchSize = size

	return b
}

// WithTaskTimeout sets the timeout for task execution
func (b *WorkBuilder) WithTaskTimeout(d time.Duration) *WorkBuilder {
	b.taskTimeout = d

	return b
}

// WithMaxRetries size the maximum retry attempts
func (b *WorkBuilder) WithMaxRetries(n int) *WorkBuilder {
	b.maxRetries = n

	return b
}

// Build creates the workflowWorker instance
func (b *WorkBuilder) Build() *WorkflowWorker {
	return &WorkflowWorker{
		repo:         b.repo,
		registry:     b.registry,
		pollInterval: b.pollInterval,
		batchSize:    b.batchSize,
		taskTimeout:  b.taskTimeout,
		maxRetries:   b.maxRetries,
	}
}
