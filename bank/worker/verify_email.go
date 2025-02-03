package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	db "main/db/sqlc"
	"main/pkg/log"
	"main/util"

	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
)

const (
	TaskVerifyEmail = "task:send_verify_email"
)

type PayloadSendVerifyEmail struct {
	UserID int64 `json:"user_id"`
}

func (distributor *RedisTaskDistributor) DistributeTaskSendVerifyEmail(ctx context.Context, payload *PayloadSendVerifyEmail, opt ...asynq.Option) error {
	jsonMarshal, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	task := asynq.NewTask(TaskVerifyEmail, jsonMarshal, opt...)

	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}
	fields := logrus.Fields{
		"type":      task.Type(),
		"queue":     info.Queue,
		"max_retry": info.MaxRetry,
	}
	log.Logger.WithFields(fields).Info("enqueued task")
	return nil
}
func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerifyEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	user, err := processor.store.GetUser(ctx, payload.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user doesn't exist: %w", asynq.SkipRetry)

		}
		return fmt.Errorf("failed to get user: %w", err)

	}

	emailVerify, err := processor.store.CreateVerifyEmail(ctx, db.CreateVerifyEmailParams{
		UserID:     user.UserID,
		Email:      user.Email,
		SecretCode: util.RandomStr(32),
	})
	if err != nil {
		return fmt.Errorf("create email verify failed: %w", err)
	}
	subject := "Welcome to Simple Bank"
	verifyUrl := fmt.Sprintf("http://simple-bank.org/verify-email?id=%d&secret_code=%s", emailVerify.ID, emailVerify.SecretCode)
	content := fmt.Sprintf(`Hello %s,<br/>
	Thank you for registering with us!<br/>
	Please <a href="%s">click here</a> to verify your email address.<br/>
	`, user.FullName, verifyUrl)
	to := []string{user.Email}

	err = processor.mailer.SendEmail(subject, content, to, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to send verify email: %w", err)
	}
	fields := logrus.Fields{
		"type":  task.Type(),
		"email": user.Email,
	}
	log.Logger.WithFields(fields).Info("processed task")
	return nil
}
