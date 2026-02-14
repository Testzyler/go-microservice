package migrate

import (
	"errors"
	"fmt"
	"os"
	"strings"

	migratev4 "github.com/golang-migrate/migrate/v4"
	"github.com/sklinkert/go-ddd/internal/builders"
)

type Config struct {
	DatabaseURL    string
	MigrationsPath string
	Command        string
	Steps          int
	Version        int
}

func DefaultConfig() Config {
	migrationsPath := strings.TrimSpace(os.Getenv("MIGRATIONS_PATH"))
	if migrationsPath == "" {
		migrationsPath = "file://migrations"
	}

	return Config{
		DatabaseURL:    "",
		MigrationsPath: migrationsPath,
		Command:        "up",
		Steps:          -1,
		Version:        -1,
	}
}

func Run(cfg Config) error {
	databaseURL := cfg.DatabaseURL
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
	}
	if databaseURL == "" {
		return errors.New("database URL is required. Use --database-url flag or set DATABASE_URL environment variable")
	}

	migrator, err := builders.NewMigratorBuilder().
		WithDatabaseURL(databaseURL).
		WithMigrationsPath(cfg.MigrationsPath).
		Build()
	if err != nil {
		return err
	}
	defer migrator.Close()

	switch cfg.Command {
	case "up":
		err = migrator.Up(cfg.Steps)
		if err != nil && err != migratev4.ErrNoChange {
			return fmt.Errorf("migration up failed: %w", err)
		}
		if err == migratev4.ErrNoChange {
			fmt.Println("No migrations to apply")
		} else {
			fmt.Println("Migrations applied successfully")
		}

	case "down":
		err = migrator.Down(cfg.Steps)
		if err != nil && err != migratev4.ErrNoChange {
			return fmt.Errorf("migration down failed: %w", err)
		}
		if err == migratev4.ErrNoChange {
			fmt.Println("No migrations to rollback")
		} else {
			fmt.Println("Migrations rolled back successfully")
		}

	case "version":
		version, dirty, versionErr := migrator.Version()
		if versionErr != nil {
			return fmt.Errorf("failed to get version: %w", versionErr)
		}
		fmt.Printf("Current version: %d (dirty: %v)\n", version, dirty)

	case "force":
		if cfg.Version < 0 {
			return errors.New("version is required for force command. Use --version flag")
		}
		if forceErr := migrator.Force(cfg.Version); forceErr != nil {
			return fmt.Errorf("force migration failed: %w", forceErr)
		}
		fmt.Printf("Forced migration to version %d\n", cfg.Version)

	default:
		return fmt.Errorf("unknown command: %s", cfg.Command)
	}

	return nil
}
