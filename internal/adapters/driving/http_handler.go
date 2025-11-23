package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/parinyadagon/go-workflow/internal/core/port"
)

type workflowHandler struct {
	svc port.WorkflowService
}

func NewWorkflowHandler(svc port.WorkflowService) *workflowHandler {
	return &workflowHandler{svc: svc}
}

func (h *workflowHandler) StartWorkflow(c echo.Context) error {
	req := &port.CreateWorkflowRequest{}

	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Invalid request body"})
	}

	result, err := h.svc.StartNewWorkflow(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Workflow stated successfully",
		"data":    result,
	})
}

// GET /workflows/:id
func (h *workflowHandler) GetWorkflowDetail(c echo.Context) error {
	id := c.Param("id")
	ctx := c.Request().Context()

	// 1. ดึงข้อมูล Workflow หลัก
	wf, err := h.svc.GetWorkflowByID(ctx, id) // (สมมติว่า Service expose Repo หรือ Wrapper)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	}

	// 2. ดึง Tasks ลูกๆ ทั้งหมด
	tasks, err := h.svc.GetTasksByWorkflowID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	}

	// 3. ดึง Activity Logs
	logs, err := h.svc.GetActivityLogsByWorkflowID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	}

	// 4. ส่งกลับไปพร้อมกัน
	return c.JSON(http.StatusOK, map[string]interface{}{
		"workflow":     wf,
		"tasks":        tasks,
		"activityLogs": logs,
	})
}

// GET /workflows
func (h *workflowHandler) ListWorkflows(c echo.Context) error {
	// Parse query parameters with defaults
	limit := 20
	offset := 0

	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := c.QueryParam("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	workflows, err := h.svc.ListWorkflows(c.Request().Context(), limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"workflows": workflows,
		"limit":     limit,
		"offset":    offset,
	})
}
