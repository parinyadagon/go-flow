package order

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/parinyadagon/go-workflow/gen/go_flow/model"
	"github.com/parinyadagon/go-workflow/pkg/logger"
)

func deductMoney(ctx context.Context, task *model.Tasks) error {
	logger.Info().Str("task", "DeductMoney").Msg("Deducting money")

	var input map[string]interface{}
	if task.InputPayload != nil {
		if err := json.Unmarshal([]byte(*task.InputPayload), &input); err != nil {
			return err
		}
	}

	time.Sleep(2 * time.Second)

	// Random failure (20%)
	if time.Now().Unix()%10 < 2 {
		return errors.New("payment gateway timeout")
	}

	output := map[string]interface{}{
		"payment_status": "SUCCESS",
		"transaction_id": "TXN" + time.Now().Format("20060102150405"),
		"amount":         input["amount"],
		"deducted_at":    time.Now().Format(time.RFC3339),
	}
	outputJSON, _ := json.Marshal(output)
	outputStr := string(outputJSON)
	task.OutputPayload = &outputStr

	logger.Info().
		Str("task_name", task.TaskName).
		Int64("task_id", task.ID).
		Str("transaction_id", output["transaction_id"].(string)).
		Msg("Payment processed successfully")

	return nil
}
