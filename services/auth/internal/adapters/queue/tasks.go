package queue

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"

	asyncpkg "github.com/Testzyler/go-microservice/pkg/async"
)

const (
	QueueCritical  = "critical"
	QueueDefault   = "default"
	QueueLow       = "low"
	QueueAudit     = "audit"
	QueueHeartbeat = "heartbeat"

	TaskWelcomeEmailV1    = "auth.v1.welcome_email"
	TaskSendMFAV1         = "auth.v1.send_mfa_challenge"
	TaskAuditLogV1        = "auth.v1.audit_log"
	TaskCleanupSessionsV1 = "auth.v1.cleanup_sessions"

	TaskHeartbeatV1 = "system.v1.heartbeat"
)

type WelcomeEmailPayload struct {
	Trace  asyncpkg.TraceCarrier `json:"trace,omitempty"`
	UserID string                `json:"user_id"`
	Email  string                `json:"email"`
}

type SendMFAPayload struct {
	Trace     asyncpkg.TraceCarrier `json:"trace,omitempty"`
	UserID    string                `json:"user_id"`
	Email     string                `json:"email,omitempty"`
	Channel   string                `json:"channel"` // email or sms
	Masked    string                `json:"masked"`
	Challenge string                `json:"challenge"`
}

type AuditLogPayload struct {
	Trace     asyncpkg.TraceCarrier `json:"trace,omitempty"`
	UserID    string                `json:"user_id"`
	Action    string                `json:"action"`
	SessionID string                `json:"session_id,omitempty"`
	IP        string                `json:"ip,omitempty"`
	UserAgent string                `json:"user_agent,omitempty"`
	Reason    string                `json:"reason,omitempty"`
}

type CleanupSessionsPayload struct {
	Trace              asyncpkg.TraceCarrier `json:"trace,omitempty"`
	GracePeriodMinutes int                   `json:"grace_period_minutes"`
}

type HeartbeatPayload struct {
	Trace   asyncpkg.TraceCarrier `json:"trace,omitempty"`
	Message string                `json:"message"`
}

func NewWelcomeEmailTask(payload WelcomeEmailPayload) (*asynq.Task, []asynq.Option, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}
	task := asynq.NewTask(TaskWelcomeEmailV1, body)
	opts := []asynq.Option{
		asynq.Queue(QueueDefault),
		asynq.Timeout(30 * time.Second),
		asynq.MaxRetry(12),
		asynq.Unique(15 * time.Minute),
		asynq.Retention(24 * time.Hour),
	}
	return task, opts, nil
}

func NewSendMFATask(payload SendMFAPayload) (*asynq.Task, []asynq.Option, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}
	task := asynq.NewTask(TaskSendMFAV1, body)
	opts := []asynq.Option{
		asynq.Queue(QueueCritical),
		asynq.Timeout(20 * time.Second),
		asynq.MaxRetry(8),
		asynq.Unique(10 * time.Minute),
		asynq.Retention(12 * time.Hour),
	}
	return task, opts, nil
}

func NewAuditLogTask(payload AuditLogPayload) (*asynq.Task, []asynq.Option, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}
	task := asynq.NewTask(TaskAuditLogV1, body)
	opts := []asynq.Option{
		asynq.Queue(QueueAudit),
		asynq.Timeout(10 * time.Second),
		asynq.MaxRetry(5),
		asynq.Retention(48 * time.Hour),
	}
	return task, opts, nil
}

func NewCleanupSessionsTask(payload CleanupSessionsPayload) (*asynq.Task, []asynq.Option, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}
	task := asynq.NewTask(TaskCleanupSessionsV1, body)
	opts := []asynq.Option{
		asynq.Queue(QueueLow),
		asynq.Timeout(60 * time.Second),
		asynq.MaxRetry(3),
		asynq.Retention(24 * time.Hour),
		asynq.Unique(15 * time.Minute),
	}
	return task, opts, nil
}

func NewHeartbeatTask(payload HeartbeatPayload) (*asynq.Task, []asynq.Option, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}
	task := asynq.NewTask(TaskHeartbeatV1, body)
	opts := []asynq.Option{
		asynq.Queue(QueueHeartbeat),
		asynq.Timeout(10 * time.Second),
		asynq.MaxRetry(3),
		asynq.Unique(5 * time.Minute),
		asynq.Retention(6 * time.Hour),
	}
	return task, opts, nil
}
