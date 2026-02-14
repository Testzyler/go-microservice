package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type queryTraceContextKey struct{}

type queryTraceData struct {
	start time.Time
	sql   string
	args  []any
}

type pgxQueryTracer struct {
	logger *zap.Logger
}

func newPGXQueryTracer(logger *zap.Logger) *pgxQueryTracer {
	return &pgxQueryTracer{logger: logger}
}

func (t *pgxQueryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	return context.WithValue(ctx, queryTraceContextKey{}, &queryTraceData{
		start: time.Now(),
		sql:   compactSQL(data.SQL),
		args:  sanitizeArgs(data.Args),
	})
}

func (t *pgxQueryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	if t == nil || t.logger == nil {
		return
	}

	trace, _ := ctx.Value(queryTraceContextKey{}).(*queryTraceData)
	if trace == nil {
		trace = &queryTraceData{start: time.Now()}
	}

	fields := []zap.Field{
		zap.String("sql", trace.sql),
		zap.Any("args", trace.args),
		zap.Duration("duration", time.Since(trace.start)),
		zap.String("command_tag", data.CommandTag.String()),
		zap.Int64("rows_affected", data.CommandTag.RowsAffected()),
	}

	if data.Err != nil {
		t.logger.Error("db query failed", append(fields, zap.Error(data.Err))...)
		return
	}

	t.logger.Debug("db query executed", fields...)
}

func compactSQL(sql string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(sql)), " ")
}

func sanitizeArgs(args []any) []any {
	if len(args) == 0 {
		return nil
	}

	out := make([]any, len(args))
	for i, arg := range args {
		switch v := arg.(type) {
		case []byte:
			out[i] = fmt.Sprintf("[bytes:%d]", len(v))
		case string:
			if len(v) > 200 {
				out[i] = v[:200] + "...(truncated)"
				continue
			}
			out[i] = v
		default:
			out[i] = arg
		}
	}

	return out
}
