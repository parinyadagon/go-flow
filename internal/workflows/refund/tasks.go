package refund

import (
	"context"
	"time"

	"github.com/parinyadagon/go-workflow/gen/go_flow/model"
	"github.com/parinyadagon/go-workflow/pkg/logger"
)

func validateRefund(ctx context.Context, task *model.Tasks) error {
	logger.Info().Str("task", "ValidateRefund").Msg("Validating refund request")
	time.Sleep(1 * time.Second)

	// Add validation logic here
	// Check if order exists, refund eligibility, etc.

	return nil
}

func processRefund(ctx context.Context, task *model.Tasks) error {
	logger.Info().Str("task", "ProcessRefund").Msg("Processing refund")
	time.Sleep(2 * time.Second)

	// Add refund processing logic here
	// Call payment gateway API to process refund

	return nil
}

func notifyCustomer(ctx context.Context, task *model.Tasks) error {
	logger.Info().Str("task", "NotifyCustomer").Msg("Notifying customer")
	time.Sleep(1 * time.Second)

	// Add notification logic here
	// Send email or SMS to customer

	return nil
}
