package refund

import (
	"github.com/parinyadagon/go-workflow/internal/core/registry"
)

// Register registers the RefundProcess workflow with all its tasks
func Register(reg *registry.WorkflowRegistry) {
	reg.NewWorkflow("RefundProcess").
		AddTask("ValidateRefund", validateRefund).
		AddTask("ProcessRefund", processRefund).
		AddTask("NotifyCustomer", notifyCustomer).
		MustBuild()
}
