package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	db "github.com/sklinkert/go-ddd/internal/infrastructure/db/sqlc"
)

func NewConnection(ctx context.Context, dsn string) (*pgx.Conn, error) {
	return NewConnectionWithLogger(ctx, dsn, nil)
}

func NewConnectionWithLogger(ctx context.Context, dsn string, logger *zap.Logger) (*pgx.Conn, error) {
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	if logger != nil {
		cfg.Tracer = newPGXQueryTracer(logger)
	}

	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func NewQueries(conn *pgx.Conn) *db.Queries {
	return db.New(conn)
}
