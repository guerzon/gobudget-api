package worker

import (
	"context"

	"github.com/hibiken/asynq"
)

// This file contains the code to create tasks and distribute to the workers via Redis queue.

// Generic interface for a task distributor.
type TaskDistributor interface {
	DistributeSendEmail(ctx context.Context, payload *SendEmailPayload, emailTask string) error
}

// Implements the TaskDistributor interface
type RedisTaskDistributor struct {
	// the client used to send the task to Redis queue
	client *asynq.Client
}

// Returns a new RedisTaskDistributor
func NewRedisTaskDistributor(redisOpts asynq.RedisClientOpt) TaskDistributor {
	return &RedisTaskDistributor{
		client: asynq.NewClient(redisOpts),
	}
}
