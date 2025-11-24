package registry

import (
	"context"
	"errors"
	"sync"

	"github.com/parinyadagon/go-workflow/gen/go_flow/model"
)

// TaskFunc is the function signature for task execute
type TaskFunc func(ctx context.Context, task *model.Tasks) error

// WorkflowDefinition holds workflow name, tasks, and their functions
type WorkflowDefinition struct {
	Name      string
	TaskNames []string
	TaskFuncs map[string]TaskFunc
}

// WorkflowRegistry manages workflow definitions
type WorkflowRegistry struct {
	mu          sync.RWMutex
	definitions map[string]*WorkflowDefinition
}

// NameWorkflowRegistry creates a new registry
func NewWorkflowRegistry() *WorkflowRegistry {
	return &WorkflowRegistry{
		definitions: make(map[string]*WorkflowDefinition),
	}
}

// Register adds a complete workflow definition
func (r *WorkflowRegistry) Register(def *WorkflowDefinition) error {
	if def.Name == "" {
		return errors.New("workflow name cannot be empty")
	}
	if len(def.TaskNames) == 0 {
		return errors.New("workflow must have at least one task")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.definitions[def.Name]; exists {
		return errors.New("workflow already registered: " + def.Name)
	}

	r.definitions[def.Name] = def

	return nil
}

// GetDefinition retrieves workflow definition
func (r *WorkflowRegistry) GetDefinition(name string) (*WorkflowDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	def, exists := r.definitions[name]
	return def, exists
}

// GetTaskFunc retrieves a specific task function
func (r *WorkflowRegistry) GetTaskFunc(workflowName, taskName string) (TaskFunc, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	def, exists := r.definitions[workflowName]
	if !exists {
		return nil, false
	}

	fn, exists := def.TaskFuncs[taskName]

	return fn, exists
}

// ListWorkflows returns all registered workflow names
func (r *WorkflowRegistry) ListWorkflows() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.definitions))
	for name := range r.definitions {
		names = append(names, name)
	}

	return names
}

// =========================
// Workflow Builder Pattern
// =========================

// WorkflowBuilder provides fluent API for builder workflows
type WorkflowBuilder struct {
	registry  *WorkflowRegistry
	name      string
	taskNames []string
	taskFuncs map[string]TaskFunc
}

// NewWorkflow creates a new workflow builder
func (r *WorkflowRegistry) NewWorkflow(name string) *WorkflowBuilder {
	return &WorkflowBuilder{
		registry:  r,
		name:      name,
		taskNames: []string{},
		taskFuncs: make(map[string]TaskFunc),
	}
}

// AddTask adds a task with its execution function
func (b *WorkflowBuilder) AddTask(taskName string, fn TaskFunc) *WorkflowBuilder {
	b.taskNames = append(b.taskNames, taskName)
	b.taskFuncs[taskName] = fn

	return b
}

// Builder registers the workflow
func (b *WorkflowBuilder) Build() error {
	def := &WorkflowDefinition{
		Name:      b.name,
		TaskNames: b.taskNames,
		TaskFuncs: b.taskFuncs,
	}

	return b.registry.Register(def)
}

// MustBuild registers the workflow and panics on error
func (b *WorkflowBuilder) MustBuild() {
	if err := b.Build(); err != nil {
		panic("failed to build workflow: " + err.Error())
	}
}
