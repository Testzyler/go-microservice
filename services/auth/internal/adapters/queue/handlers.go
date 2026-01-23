package queue

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type MailerHandler struct {
	logger *zap.Logger
}

func NewMailerHandler(logger *zap.Logger) *MailerHandler {
	return &MailerHandler{logger: logger}
}

func (h *MailerHandler) Register(mux *asynq.ServeMux) {
	mux.HandleFunc(TaskWelcomeEmailV1, h.HandleWelcomeEmail)
	mux.HandleFunc(TaskSendMFAV1, h.HandleSendMFA)
}

func (h *MailerHandler) HandleWelcomeEmail(ctx context.Context, task *asynq.Task) error {
	ctx, span := otel.Tracer("auth.worker.mailer").Start(ctx, "HandleWelcomeEmail")
	defer span.End()

	var payload WelcomeEmailPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}
	h.logger.Info("send welcome email", zap.String("user_id", payload.UserID), zap.String("email", payload.Email))
	return nil
}

func (h *MailerHandler) HandleSendMFA(ctx context.Context, task *asynq.Task) error {
	ctx, span := otel.Tracer("auth.worker.mailer").Start(ctx, "HandleSendMFA")
	defer span.End()

	var payload SendMFAPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}
	h.logger.Info("send mfa challenge", zap.String("user_id", payload.UserID), zap.String("channel", payload.Channel), zap.String("masked", payload.Masked))
	return nil
}

type AuditHandler struct {
	logger *zap.Logger
}

func NewAuditHandler(logger *zap.Logger) *AuditHandler {
	return &AuditHandler{logger: logger}
}

func (h *AuditHandler) Register(mux *asynq.ServeMux) {
	mux.HandleFunc(TaskAuditLogV1, h.HandleAuditLog)
	mux.HandleFunc(TaskHeartbeatV1, h.HandleHeartbeat)
}

func (h *AuditHandler) HandleAuditLog(ctx context.Context, task *asynq.Task) error {
	ctx, span := otel.Tracer("auth.worker.audit").Start(ctx, "HandleAuditLog")
	defer span.End()

	var payload AuditLogPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}
	h.logger.Info("audit auth event",
		zap.String("action", payload.Action),
		zap.String("user_id", payload.UserID),
		zap.String("session_id", payload.SessionID),
		zap.String("ip", payload.IP),
		zap.String("user_agent", payload.UserAgent),
		zap.String("reason", payload.Reason),
	)
	return nil
}

func (h *AuditHandler) HandleHeartbeat(ctx context.Context, task *asynq.Task) error {
	ctx, span := otel.Tracer("auth.worker.audit").Start(ctx, "HandleHeartbeat")
	defer span.End()

	var payload HeartbeatPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}
	h.logger.Debug("heartbeat", zap.String("message", payload.Message))
	return nil
}

type CleanupHandler struct {
	logger *zap.Logger
}

func NewCleanupHandler(logger *zap.Logger) *CleanupHandler {
	return &CleanupHandler{logger: logger}
}

func (h *CleanupHandler) Register(mux *asynq.ServeMux) {
	mux.HandleFunc(TaskCleanupSessionsV1, h.HandleCleanupSessions)
	mux.HandleFunc(TaskHeartbeatV1, h.HandleHeartbeat)
}

func (h *CleanupHandler) HandleCleanupSessions(ctx context.Context, task *asynq.Task) error {
	ctx, span := otel.Tracer("auth.worker.cleanup").Start(ctx, "HandleCleanupSessions")
	defer span.End()

	var payload CleanupSessionsPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}
	h.logger.Info("cleanup stale sessions", zap.Int("grace_period_minutes", payload.GracePeriodMinutes))
	return nil
}

func (h *CleanupHandler) HandleHeartbeat(ctx context.Context, task *asynq.Task) error {
	ctx, span := otel.Tracer("auth.worker.cleanup").Start(ctx, "HandleHeartbeat")
	defer span.End()

	var payload HeartbeatPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}
	h.logger.Debug("heartbeat", zap.String("message", payload.Message))
	return nil
}
