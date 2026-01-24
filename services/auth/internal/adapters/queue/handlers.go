package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/Testzyler/go-microservice/global/pkg/async"
	"github.com/Testzyler/go-microservice/global/pkg/email"
)

type MailerHandler struct {
	logger      *zap.Logger
	sender      email.Sender
	serviceName string
}

func NewMailerHandler(logger *zap.Logger, sender email.Sender, serviceName string) *MailerHandler {
	return &MailerHandler{
		logger:      logger,
		sender:      sender,
		serviceName: serviceName,
	}
}

func (h *MailerHandler) Register(mux *asynq.ServeMux) {
	mux.HandleFunc(TaskWelcomeEmailV1, h.HandleWelcomeEmail)
	mux.HandleFunc(TaskSendMFAV1, h.HandleSendMFA)
}

func (h *MailerHandler) HandleWelcomeEmail(ctx context.Context, task *asynq.Task) error {
	var payload WelcomeEmailPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}
	ctx, span := startTaskSpan(ctx, "auth.worker.mailer", taskSpanName(task), task, payload.Trace)
	defer span.End()
	subject := fmt.Sprintf("Welcome to %s", h.brandName())
	htmlBody := fmt.Sprintf("<p>Hi,</p><p>Welcome to %s.</p>", h.brandName())
	textBody := fmt.Sprintf("Hi,\n\nWelcome to %s.\n", h.brandName())
	if err := h.sendEmail(ctx, payload.Email, subject, htmlBody, textBody); err != nil {
		return err
	}
	h.logger.Info("sent welcome email", zap.String("user_id", payload.UserID), zap.String("email", payload.Email))
	return nil
}

func (h *MailerHandler) HandleSendMFA(ctx context.Context, task *asynq.Task) error {
	var payload SendMFAPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}
	ctx, span := startTaskSpan(ctx, "auth.worker.mailer", taskSpanName(task), task, payload.Trace)
	defer span.End()
	if !strings.EqualFold(payload.Channel, "email") {
		h.logger.Info("skip mfa send; unsupported channel",
			zap.String("user_id", payload.UserID),
			zap.String("channel", payload.Channel),
		)
		return nil
	}
	if strings.TrimSpace(payload.Email) == "" {
		return fmt.Errorf("mfa email missing: %w", asynq.SkipRetry)
	}
	subject := fmt.Sprintf("Your %s verification code", h.brandName())
	htmlBody := fmt.Sprintf("<p>Your verification code is <strong>%s</strong>.</p>", payload.Challenge)
	textBody := fmt.Sprintf("Your verification code is %s.", payload.Challenge)
	if err := h.sendEmail(ctx, payload.Email, subject, htmlBody, textBody); err != nil {
		return err
	}
	h.logger.Info("sent mfa email",
		zap.String("user_id", payload.UserID),
		zap.String("masked", payload.Masked),
	)
	return nil
}

func (h *MailerHandler) sendEmail(ctx context.Context, to, subject, htmlBody, textBody string) error {
	if h.sender == nil {
		return fmt.Errorf("email sender not configured: %w", asynq.SkipRetry)
	}
	if strings.TrimSpace(to) == "" {
		return fmt.Errorf("recipient email is required: %w", asynq.SkipRetry)
	}
	msg := email.Message{
		To:       to,
		Subject:  subject,
		HtmlBody: htmlBody,
		TextBody: textBody,
	}
	if err := h.sender.Send(ctx, msg); err != nil {
		return err
	}
	return nil
}

func (h *MailerHandler) brandName() string {
	name := strings.TrimSpace(h.serviceName)
	if name == "" {
		return "auth"
	}
	return name
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
	var payload AuditLogPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}
	ctx, span := startTaskSpan(ctx, "auth.worker.audit", taskSpanName(task), task, payload.Trace)
	defer span.End()
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
	var payload HeartbeatPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}
	ctx, span := startTaskSpan(ctx, "auth.worker.audit", taskSpanName(task), task, payload.Trace)
	defer span.End()
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
	var payload CleanupSessionsPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}
	ctx, span := startTaskSpan(ctx, "auth.worker.cleanup", taskSpanName(task), task, payload.Trace)
	defer span.End()
	h.logger.Info("cleanup stale sessions", zap.Int("grace_period_minutes", payload.GracePeriodMinutes))
	return nil
}

func (h *CleanupHandler) HandleHeartbeat(ctx context.Context, task *asynq.Task) error {
	var payload HeartbeatPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}
	ctx, span := startTaskSpan(ctx, "auth.worker.cleanup", taskSpanName(task), task, payload.Trace)
	defer span.End()
	h.logger.Debug("heartbeat", zap.String("message", payload.Message))
	return nil
}

func taskSpanName(task *asynq.Task) string {
	return "task " + task.Type()
}

func startTaskSpan(ctx context.Context, tracerName, spanName string, task *asynq.Task, carrier async.TraceCarrier) (context.Context, trace.Span) {
	ctx = async.ExtractTrace(ctx, carrier)
	ctx, span := otel.Tracer(tracerName).Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindConsumer))
	span.SetAttributes(
		attribute.String("messaging.system", "asynq"),
		attribute.String("messaging.operation", "process"),
		attribute.String("messaging.message_type", task.Type()),
	)
	if id, ok := asynq.GetTaskID(ctx); ok {
		span.SetAttributes(attribute.String("messaging.message_id", id))
	}
	if queue, ok := asynq.GetQueueName(ctx); ok {
		span.SetAttributes(attribute.String("messaging.destination", queue))
	}
	return ctx, span
}
