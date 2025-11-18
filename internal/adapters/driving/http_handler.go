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
