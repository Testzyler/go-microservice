package migratecmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	envconfig "github.com/Testzyler/go-microservice/global/pkg/config"
)

func NewCommand() *cobra.Command {
	var envFiles string

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, envFiles)
		},
	}

	flags := cmd.Flags()
	flags.String("path", "", "path to migrations directory")
	flags.String("database", "", "database URL")
	flags.String("action", "up", "migration action: up|down|version|force")
	flags.Int("steps", 0, "steps for up/down (0 = all)")
	flags.Int("force", 0, "force version for action=force")
	flags.StringVar(&envFiles, "env-files", "", "comma-separated env files to load")

	return cmd
}

func run(cmd *cobra.Command, envFiles string) error {
	v, err := envconfig.NewViper(envconfig.Options{
		EnvFiles: envconfig.ParseEnvFiles(envFiles),
		Defaults: map[string]interface{}{
			"action": "up",
			"steps":  0,
			"force":  0,
		},
	})
	if err != nil {
		return err
	}

	flags := cmd.Flags()
	if err := bindFlags(v, flags); err != nil {
		return err
	}
	if err := v.BindEnv("database", "DATABASE_URL"); err != nil {
		return err
	}
	if err := v.BindEnv("path", "MIGRATIONS_PATH", "MIGRATION_PATH"); err != nil {
		return err
	}

	path := strings.TrimSpace(v.GetString("path"))
	databaseURL := strings.TrimSpace(v.GetString("database"))
	action := strings.ToLower(strings.TrimSpace(v.GetString("action")))
	steps := v.GetInt("steps")
	forceVer := v.GetInt("force")

	if path == "" {
		return errors.New("path is required")
	}
	if databaseURL == "" {
		return errors.New("database is required")
	}

	return runMigrations(path, databaseURL, action, steps, forceVer)
}

func bindFlags(v *viper.Viper, flags *pflag.FlagSet) error {
	if err := v.BindPFlag("path", flags.Lookup("path")); err != nil {
		return err
	}
	if err := v.BindPFlag("database", flags.Lookup("database")); err != nil {
		return err
	}
	if err := v.BindPFlag("action", flags.Lookup("action")); err != nil {
		return err
	}
	if err := v.BindPFlag("steps", flags.Lookup("steps")); err != nil {
		return err
	}
	if err := v.BindPFlag("force", flags.Lookup("force")); err != nil {
		return err
	}
	return nil
}

func runMigrations(path, databaseURL, action string, steps, forceVer int) error {
	m, err := migrate.New("file://"+path, databaseURL)
	if err != nil {
		return fmt.Errorf("migration init failed: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	switch action {
	case "up":
		if steps > 0 {
			err = m.Steps(steps)
		} else {
			err = m.Up()
		}
	case "down":
		if steps > 0 {
			err = m.Steps(-steps)
		} else {
			err = m.Down()
		}
	case "version":
		ver, dirty, verr := m.Version()
		if verr == migrate.ErrNilVersion {
			fmt.Println("no version")
			return nil
		}
		if verr != nil {
			return fmt.Errorf("version check failed: %w", verr)
		}
		fmt.Printf("version=%d dirty=%v\n", ver, dirty)
		return nil
	case "force":
		if forceVer <= 0 {
			return errors.New("force requires -force version")
		}
		err = m.Force(forceVer)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}
	return nil
}
