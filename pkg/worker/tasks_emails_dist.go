package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// SendVerifyEmailPayload will contain all data that we want to store in the Redis queue.
type SendEmailPayload struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

const TaskSendVerifyEmail = "task:send_verify_email"
const TaskSendAccountDeletedEmail = "task:send_account_deleted_email"

// DistributeSendEmail implements the TaskDistributor interface and distributes email sending tasks.
func (d *RedisTaskDistributor) DistributeSendEmail(ctx context.Context, payload *SendEmailPayload, emailTask string) error {

	j, err := json.Marshal(payload)

	if err != nil {
		return fmt.Errorf("cannot marshal task payload: %w", err)
	}

	// Initialize a new task
	opts := []asynq.Option{
		asynq.MaxRetry(10),
		asynq.ProcessIn(10 * time.Second),
		asynq.Queue(QueueCritical),
	}
	task := asynq.NewTask(emailTask, j, opts...)

	// Send the task
	_, err = d.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("cannot enqueue task: %w", err)
	}
	return nil
}
