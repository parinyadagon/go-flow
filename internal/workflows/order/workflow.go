package order

import (
	"github.com/parinyadagon/go-workflow/internal/core/registry"
)

// Register registers the OrderProcess workflow with all its tasks
func Register(reg *registry.WorkflowRegistry) {
	reg.NewWorkflow("OrderProcess").
		AddTask("ValidateOrder", validateOrder).
		AddTask("DeductMoney", deductMoney).
		AddTask("SendEmail", sendEmail).
		MustBuild()
}
