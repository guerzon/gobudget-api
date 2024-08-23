package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/guerzon/gobudget-api/pkg/db"
	"github.com/guerzon/gobudget-api/pkg/util"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"golang.org/x/exp/slog"
)

const EmailVerificationExpiration = time.Duration(time.Minute * 15)

// ProcessSendVerifyEmail implements the TaskProcessor interface and processes the task task:send_verify_email from the background worker
func (p *RedisTaskProcessor) ProcessSendVerifyEmail(ctx context.Context, task *asynq.Task) error {

	var payload SendEmailPayload

	// unmarshal the payload inside the task
	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return fmt.Errorf("cannot unmarshal task payload: %w", asynq.SkipRetry)
	}

	user, err := p.store.GetUserByUsername(ctx, payload.Username)
	if err != nil {
		if err == sql.ErrNoRows || err == pgx.ErrNoRows {
			return fmt.Errorf("user does not exist: %w", asynq.SkipRetry) // don't retry
		}
		return fmt.Errorf("failed to get user: %w", err) // this will retry
	}

	// Create an entry in verify_emails
	emailVerArg := db.CreateVerifyEmailsParams{
		Username:  user.Username,
		Email:     user.Email,
		Code:      util.RandomString(32, ""),
		ExpiresAt: time.Now().Add(EmailVerificationExpiration),
	}
	e, err := p.store.CreateVerifyEmails(ctx, emailVerArg)
	if err != nil {
		slog.Error("cannot create record for send email verification")
		return fmt.Errorf("failed to create verify email record: %v", err)
	}

	// Send the email to the user
	// TODO: improve this
	s := "Welcome to gobudget API"
	c := `
	<p>Hello ` + e.Username + `,</p>
	<br/>
	<p>Thanks for creating an account. Kindly verify your email by clicking <a href="` + "http://localhost:8080/beta/verify_email?id=" + strconv.Itoa(int(e.ID)) + `&code=` + e.Code + `">here.</a>
	</p>
	<br/>
	Thanks!
	`
	err = p.mailer.SendEmail(s, c, []string{user.Email}, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("cannot send verification email: %w", err)
	}

	slog.Info(fmt.Sprintf("[processed_task] email=%s", user.Email))

	return nil
}

// ProcessSendVerifyEmail implements the TaskProcessor interface and processes the task task:send_verify_email from the background worker
func (p *RedisTaskProcessor) ProcessSendAccountDeletedEmail(ctx context.Context, task *asynq.Task) error {

	var payload SendEmailPayload

	// unmarshal the payload inside the task
	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return fmt.Errorf("cannot unmarshal task payload: %w", asynq.SkipRetry)
	}

	// Send the email to the user
	// TODO: improve this:
	s := "We're sorry to see you go"
	c := `
	<p>Hello ` + payload.Username + `,</p>
	<br/>
	<p>Your account has been deleted from our website. In case you wish to come back in the future, remember that we will be here for you!
	</p>
	<br/>
	Thanks!
	`
	err = p.mailer.SendEmail(s, c, []string{payload.Email}, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("cannot send account deletion email: %w", err)
	}

	slog.Info(fmt.Sprintf("[processed_task] email=%s", payload.Email))

	return nil
}
