package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/parinyadagon/go-workflow/internal/core/port"
)

type workflowHandler struct {
	svc port.WorkflowService
}

func NewWorkflowHandler(svc port.WorkflowService) *workflowHandler {
	return &workflowHandler{svc: svc}
}

func (h *workflowHandler) StartWorkflow(c *fiber.Ctx) error {
	req := &port.CreateWorkflowRequest{}

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	result, err := h.svc.StartNewWorkflow(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Workflow stated successfully",
		"data":    result,
	})
}

// GET /workflows/:id
func (h *workflowHandler) GetWorkflowDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx := c.Context()

	// 1. ดึงข้อมูล Workflow หลัก
	wf, err := h.svc.GetWorkflowByID(ctx, id) // (สมมติว่า Service expose Repo หรือ Wrapper)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 2. ดึง Tasks ลูกๆ ทั้งหมด
	tasks, err := h.svc.GetTasksByWorkflowID(ctx, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 3. ส่งกลับไปพร้อมกัน
	return c.JSON(fiber.Map{
		"workflow": wf,
		"tasks":    tasks,
	})
}
