package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/parinyadagon/go-workflow/gen/go_flow/model"
	"github.com/parinyadagon/go-workflow/internal/core/port"
)

type workflowService struct {
	repo port.WorkflowRepository
}

func NewWorkflowService(repo port.WorkflowRepository) port.WorkflowService {
	return &workflowService{repo: repo}
}

var WorkflowDefinitions = map[string][]string{
	"OrderProcess": {"ValidateOrder", "DeductMoney", "SendEmail"},
}

func (s *workflowService) StartNewWorkflow(ctx context.Context, req *port.CreateWorkflowRequest) (*model.WorkflowInstances, error) {
	newID := uuid.New().String()
	status := model.WorkflowInstancesStatus_Pending

	inputJSON, _ := json.Marshal(req.InputPayload)
	inputStr := string(inputJSON)

	wf := &model.WorkflowInstances{
		ID:           newID,
		WorkflowName: req.WorkflowName,
		Status:       &status,
		CurrentInput: &inputStr,
	}

	steps, exists := WorkflowDefinitions[req.WorkflowName]
	if !exists || len(steps) == 0 {
		return nil, fmt.Errorf("unknown workflow: %s", req.WorkflowName)
	}
	firstTaskName := steps[0] // "ValidateOrder"
	taskStatus := model.TasksStatus_Pending

	firstTask := &model.Tasks{
		WorkflowInstanceID: newID,
		TaskName:           firstTaskName,
		Status:             &taskStatus,
		InputPayload:       &inputStr,
	}

	if err := s.repo.CreateWorkflow(ctx, wf); err != nil {
		return nil, err
	}

	if err := s.repo.CreateTask(ctx, firstTask); err != nil {
		return nil, err
	}

	return wf, nil
}

func (s *workflowService) ListWorkflows(ctx context.Context, limit int, offset int) ([]model.WorkflowInstances, error) {
	return s.repo.ListWorkflows(ctx, limit, offset)
}

func (s *workflowService) GetWorkflowByID(ctx context.Context, id string) (*model.WorkflowInstances, error) {
	return s.repo.GetWorkflowByID(ctx, id)
}
func (s *workflowService) GetTasksByWorkflowID(ctx context.Context, wfID string) ([]model.Tasks, error) {
	return s.repo.GetTasksByWorkflowID(ctx, wfID)
}

func (s *workflowService) GetActivityLogsByWorkflowID(ctx context.Context, wfID string) ([]model.ActivityLogs, error) {
	return s.repo.GetActivityLogsByWorkflowID(ctx, wfID)
}
