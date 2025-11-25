package order

import (
	"context"
	"encoding/json"
	"time"

	"github.com/parinyadagon/go-workflow/gen/go_flow/model"
	"github.com/parinyadagon/go-workflow/pkg/logger"
)

func sendEmail(ctx context.Context, task *model.Tasks) error {
	logger.Info().Str("task", "SendEmail").Msg("Sending email")

	var input map[string]interface{}
	if task.InputPayload != nil {
		if err := json.Unmarshal([]byte(*task.InputPayload), &input); err != nil {
			return err
		}
	}

	time.Sleep(1 * time.Second)

	output := map[string]interface{}{
		"email_sent":  true,
		"recipient":   "customer@example.com",
		"sent_at":     time.Now().Format(time.RFC3339),
		"order_id":    input["order_id"],
		"transaction": input["transaction_id"],
	}
	outputJSON, _ := json.Marshal(output)
	outputStr := string(outputJSON)
	task.OutputPayload = &outputStr

	logger.Info().
		Str("task_name", task.TaskName).
		Int64("task_id", task.ID).
		Str("recipient", output["recipient"].(string)).
		Msg("Email sent successfully")

	return nil
}
