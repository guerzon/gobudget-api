package worker

import (
	"context"

	"github.com/guerzon/gobudget-api/pkg/db"
	"github.com/guerzon/gobudget-api/pkg/util"
	"github.com/hibiken/asynq"
)

// The processor picks up the task from the Redis queue and process them.

const (
	QueueDefault  = "default"
	QueueCritical = "critical"
)

// Generic interface so it's easier to mock for unit testing
type TaskProcessor interface {
	Start() error
	ProcessSendVerifyEmail(ctx context.Context, task *asynq.Task) error
	ProcessSendAccountDeletedEmail(ctx context.Context, task *asynq.Task) error
}

// Implements the TaskProcessor interface
type RedisTaskProcessor struct {
	// the server is used to TODO
	server *asynq.Server
	store  db.Store
	mailer util.EmailSender
}

// Creates a new Redis task processor.
func NewRedisTaskProcessor(redisOpts asynq.RedisClientOpt, store db.Store, mailer util.EmailSender) TaskProcessor {

	server := asynq.NewServer(redisOpts, asynq.Config{
		Queues: map[string]int{
			QueueDefault:  5,
			QueueCritical: 10,
		},
	})

	return &RedisTaskProcessor{
		server: server,
		store:  store,
		mailer: mailer,
	}
}

// Registers tasks with the processor
func (p *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	// Register tasks here
	mux.HandleFunc(TaskSendVerifyEmail, p.ProcessSendVerifyEmail)
	mux.HandleFunc(TaskSendAccountDeletedEmail, p.ProcessSendAccountDeletedEmail)

	return p.server.Start(mux)
}
