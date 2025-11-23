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

func (r *workflowRepo) ListWorkflows(ctx context.Context, limit int, offset int) ([]model.WorkflowInstances, error) {
	var dest []model.WorkflowInstances

	stmt := table.WorkflowInstances.SELECT(
		table.WorkflowInstances.AllColumns,
	).FROM(
		table.WorkflowInstances,
	).ORDER_BY(
		table.WorkflowInstances.CreatedAt.DESC(),
	).LIMIT(int64(limit)).OFFSET(int64(offset))

	err := stmt.QueryContext(ctx, r.db, &dest)

	return dest, err
}

func (r *workflowRepo) CountWorkflows(ctx context.Context) (int64, error) {
	var count struct {
		Total int64
	}

	stmt := mysql.SELECT(
		mysql.COUNT(mysql.STAR).AS("total"),
	).FROM(
		table.WorkflowInstances,
	)

	err := stmt.QueryContext(ctx, r.db, &count)
	if err != nil {
		return 0, err
	}

	return count.Total, nil
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

func (r *workflowRepo) GetTasksByWorkflowID(ctx context.Context, wfID string) ([]model.Tasks, error) {
	var dest []model.Tasks

	stmt := table.Tasks.SELECT(
		table.Tasks.AllColumns,
	).FROM(
		table.Tasks,
	).WHERE(
		table.Tasks.WorkflowInstanceID.EQ(mysql.String(wfID)),
	).ORDER_BY(
		table.Tasks.ID.ASC(),
	)

	err := stmt.QueryContext(ctx, r.db, &dest)

	return dest, err
}

func (r *workflowRepo) CreateActivityLog(ctx context.Context, log *model.ActivityLogs) error {
	stmt := table.ActivityLogs.
		INSERT(
			table.ActivityLogs.WorkflowInstanceID,
			table.ActivityLogs.TaskName,
			table.ActivityLogs.EventType,
			table.ActivityLogs.Details,
		).MODEL(log)

	_, err := stmt.ExecContext(ctx, r.db)

	return err
}

func (r *workflowRepo) GetActivityLogsByWorkflowID(ctx context.Context, wfID string) ([]model.ActivityLogs, error) {
	var dest []model.ActivityLogs

	stmt := table.ActivityLogs.SELECT(
		table.ActivityLogs.AllColumns,
	).FROM(
		table.ActivityLogs,
	).WHERE(
		table.ActivityLogs.WorkflowInstanceID.EQ(mysql.String(wfID)),
	).ORDER_BY(
		table.ActivityLogs.CreatedAt.ASC(),
	)

	err := stmt.QueryContext(ctx, r.db, &dest)

	return dest, err
}

func (r *workflowRepo) UpdateTaskRetryCount(ctx context.Context, id int, retryCount int) error {
	stmt := table.Tasks.UPDATE(
		table.Tasks.RetryCount,
	).SET(
		retryCount,
	).WHERE(
		table.Tasks.ID.EQ(mysql.Int(int64(id))),
	)

	_, err := stmt.ExecContext(ctx, r.db)

	return err
}

func (r *workflowRepo) GetTasksForRetry(ctx context.Context, limit int) ([]model.Tasks, error) {
	var dest []model.Tasks
	stmt := table.Tasks.SELECT(
		table.Tasks.AllColumns,
	).FROM(
		table.Tasks,
	).WHERE(
		table.Tasks.Status.EQ(mysql.String("FAILED")),
	).LIMIT(int64(limit))

	err := stmt.QueryContext(ctx, r.db, &dest)

	return dest, err
}
