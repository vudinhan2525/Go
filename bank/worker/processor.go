package worker

import (
	"context"
	db "main/db/sqlc"
	pkg "main/pkg/mail"

	"github.com/hibiken/asynq"
)

type TaskProcessor interface {
	Start() error
	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
}
type RedisTaskProcessor struct {
	server *asynq.Server
	store  db.Store
	mailer pkg.EmailSender
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store, mailer pkg.EmailSender) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{},
	)
	return &RedisTaskProcessor{
		server: server,
		store:  store,
		mailer: mailer,
	}
}
func (processor *RedisTaskProcessor) Start() error {

	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskVerifyEmail, processor.ProcessTaskSendVerifyEmail)

	return processor.server.Start(mux)
}
