package repository

import (
	"context"
	"database/sql"

	"github.com/go-jet/jet/v2/mysql"
	"github.com/parinyadagon/go-workflow/gen/go_flow/model"
	"github.com/parinyadagon/go-workflow/gen/go_flow/table"
	"github.com/parinyadagon/go-workflow/internal/core/port"
)

type workflowRepo struct {
	db *sql.DB
}

func NewWorkflowRepository(db *sql.DB) port.WorkflowRepository {
	return &workflowRepo{db: db}
}

func (r *workflowRepo) CreateWorkflow(ctx context.Context, wf *model.WorkflowInstances) error {
	stmt := table.WorkflowInstances.
		INSERT(
			table.WorkflowInstances.ID,
			table.WorkflowInstances.WorkflowName,
			table.WorkflowInstances.Status,
			table.WorkflowInstances.CurrentInput,
		).MODEL(wf) // map struct เข้า db อัตโนมัตฺิ

	_, err := stmt.ExecContext(ctx, r.db)

	return err
}
func (r *workflowRepo) CreateTask(ctx context.Context, task *model.Tasks) error {
	stmt := table.Tasks.
		INSERT(
			table.Tasks.WorkflowInstanceID,
			table.Tasks.TaskName,
			table.Tasks.Status,
			table.Tasks.InputPayload,
		).MODEL(task) // map struct เข้า db อัตโนมัตฺิ

	_, err := stmt.ExecContext(ctx, r.db)

	return err
}

func (r *workflowRepo) GetWorkflowPending(ctx context.Context, limit int) ([]model.WorkflowInstances, error) {
	var dest []model.WorkflowInstances

	stmt := table.WorkflowInstances.SELECT(
		table.WorkflowInstances.AllColumns,
	).FROM(
		table.WorkflowInstances,
	).WHERE(
		table.WorkflowInstances.Status.EQ(mysql.String("PENDING")),
	).LIMIT(int64(limit))

	err := stmt.QueryContext(ctx, r.db, &dest)

	return dest, err
}

func (r *workflowRepo) UpdateWorkflowStatus(ctx context.Context, id string, status string) error {
	stmt := table.WorkflowInstances.UPDATE(
		table.WorkflowInstances.Status,
	).SET(
		status,
	).WHERE(
		table.WorkflowInstances.ID.EQ(mysql.String(id)),
	)

	_, err := stmt.ExecContext(ctx, r.db)

	return err
}

func (r *workflowRepo) GetWorkflowByID(ctx context.Context, id string) (*model.WorkflowInstances, error) {
	var dest model.WorkflowInstances
	stmt := table.WorkflowInstances.SELECT(
		table.WorkflowInstances.AllColumns,
	).WHERE(table.WorkflowInstances.ID.IN(mysql.String(id)))

	err := stmt.QueryContext(ctx, r.db, &dest)

	return &dest, err
}

func (r *workflowRepo) GetTaskPending(ctx context.Context, limit int) ([]model.Tasks, error) {
	var dest []model.Tasks
	stmt := table.Tasks.SELECT(
		table.Tasks.AllColumns,
	).FROM(
		table.Tasks,
	).WHERE(
		table.Tasks.Status.EQ(mysql.String("PENDING")),
	).LIMIT(int64(limit))

	err := stmt.QueryContext(ctx, r.db, &dest)

	return dest, err
}

func (r *workflowRepo) UpdateTaskStatus(ctx context.Context, id int, status string) error {
	stmt := table.Tasks.UPDATE(
		table.Tasks.Status,
	).SET(
		status,
	).WHERE(
		table.Tasks.ID.EQ(mysql.Int(int64(id))),
	)

	_, err := stmt.ExecContext(ctx, r.db)

	return err
}
