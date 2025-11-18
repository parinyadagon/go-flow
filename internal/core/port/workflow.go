package port

import (
	"context"

	"github.com/parinyadagon/go-workflow/gen/go_flow/model"
)

type CreateWorkflowRequest struct {
	WorkflowName string         `json:"workflow_name"`
	InputPayload map[string]any `json:"input_payload"`
}

type WorkflowRepository interface {
	// Workflow operation
	CreateWorkflow(ctx context.Context, workflow *model.WorkflowInstances) error
	GetWorkflowPending(ctx context.Context, limit int) ([]model.WorkflowInstances, error)
	UpdateWorkflowStatus(ctx context.Context, id string, status string) error
	GetWorkflowByID(cxt context.Context, id string) (*model.WorkflowInstances, error)

	// Task operation
	CreateTask(ctx context.Context, workflow *model.Tasks) error
	GetTaskPending(ctx context.Context, limit int) ([]model.Tasks, error)
	UpdateTaskStatus(ctx context.Context, id int, status string) error
}

type WorkflowService interface {
	StartNewWorkflow(ctx context.Context, req *CreateWorkflowRequest) (*model.WorkflowInstances, error)
}
