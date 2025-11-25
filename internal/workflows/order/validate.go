package order

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/parinyadagon/go-workflow/gen/go_flow/model"
	"github.com/parinyadagon/go-workflow/pkg/logger"
)

func validateOrder(ctx context.Context, task *model.Tasks) error {
	logger.Info().Str("task", "ValidateOrder").Msg("Validating order")

	var input map[string]interface{}
	if task.InputPayload != nil {
		if err := json.Unmarshal([]byte(*task.InputPayload), &input); err != nil {
			return err
		}
	}

	time.Sleep(1 * time.Second)

	// Validation logic
	if orderID, ok := input["order_id"].(string); ok && orderID == "" {
		return errors.New("order_id is required")
	}
	if amount, ok := input["amount"].(float64); ok && amount <= 0 {
		return errors.New("amount must be positive")
	}

	// Random failure for testing (30%)
	if time.Now().Unix()%10 < 3 {
		return errors.New("validation failed randomly")
	}

	output := map[string]interface{}{
		"validated":    true,
		"order_id":     input["order_id"],
		"amount":       input["amount"],
		"validated_at": time.Now().Format(time.RFC3339),
	}
	outputJSON, _ := json.Marshal(output)
	outputStr := string(outputJSON)
	task.OutputPayload = &outputStr

	logger.Info().
		Str("task_name", task.TaskName).
		Int64("task_id", task.ID).
		Msg("Order validated successfully")

	return nil
}
