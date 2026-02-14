package builders

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	migratev4 "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const defaultMigrationsPath = "file://migrations"

type MigratorBuilder struct {
	databaseURL    string
	migrationsPath string
}

type Migrator struct {
	db *sql.DB
	m  *migratev4.Migrate
}

func NewMigratorBuilder() *MigratorBuilder {
	return &MigratorBuilder{
		migrationsPath: defaultMigrationsPath,
	}
}

func (b *MigratorBuilder) WithDatabaseURL(databaseURL string) *MigratorBuilder {
	b.databaseURL = strings.TrimSpace(databaseURL)
	return b
}

func (b *MigratorBuilder) WithMigrationsPath(migrationsPath string) *MigratorBuilder {
	if strings.TrimSpace(migrationsPath) != "" {
		b.migrationsPath = strings.TrimSpace(migrationsPath)
	}
	return b
}

func (b *MigratorBuilder) Build() (*Migrator, error) {
	if b.databaseURL == "" {
		return nil, errors.New("database URL is required")
	}

	db, err := sql.Open("pgx", b.databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create database driver: %w", err)
	}

	m, err := migratev4.NewWithDatabaseInstance(b.migrationsPath, "postgres", driver)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return &Migrator{db: db, m: m}, nil
}

func (m *Migrator) Up(steps int) error {
	if steps > 0 {
		return m.m.Steps(steps)
	}
	return m.m.Up()
}

func (m *Migrator) Down(steps int) error {
	if steps > 0 {
		return m.m.Steps(-steps)
	}
	return m.m.Down()
}

func (m *Migrator) Version() (uint, bool, error) {
	return m.m.Version()
}

func (m *Migrator) Force(version int) error {
	return m.m.Force(version)
}

func (m *Migrator) Close() error {
	var closeErr error
	if m.m != nil {
		sourceErr, databaseErr := m.m.Close()
		if sourceErr != nil {
			closeErr = sourceErr
		}
		if databaseErr != nil && closeErr == nil {
			closeErr = databaseErr
		}
	}

	if m.db != nil {
		if err := m.db.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
	}

	return closeErr
}
