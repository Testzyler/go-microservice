package root

import (
	"github.com/sklinkert/go-ddd/cmd/migrate"
	"github.com/sklinkert/go-ddd/cmd/serve"
	"github.com/spf13/cobra"
)

type CommandBuilder struct {
	serveConfig   serve.Config
	migrateConfig migrate.Config
}

func NewCommandBuilder() *CommandBuilder {
	return &CommandBuilder{
		serveConfig:   serve.DefaultConfig(),
		migrateConfig: migrate.DefaultConfig(),
	}
}

func (b *CommandBuilder) Build() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "go-ddd",
		Short:         "go-ddd marketplace service",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return serve.Run(cmd.Context(), b.serveConfig)
		},
	}

	rootCmd.Flags().StringVar(&b.serveConfig.DSN, "dsn", b.serveConfig.DSN, "PostgreSQL DSN for serve command")
	rootCmd.Flags().StringVar(&b.serveConfig.Port, "port", b.serveConfig.Port, "HTTP listen port (for example :8080)")
	rootCmd.Flags().StringSliceVar(&b.serveConfig.CORSOrigins, "cors-origins", b.serveConfig.CORSOrigins, "Allowed CORS origins")
	rootCmd.Flags().StringVar(&b.serveConfig.Environment, "app-env", b.serveConfig.Environment, "Application environment (development|production)")
	rootCmd.Flags().StringVar(&b.serveConfig.LogLevel, "log-level", b.serveConfig.LogLevel, "Zap log level (debug|info|warn|error)")

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start HTTP server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return serve.Run(cmd.Context(), b.serveConfig)
		},
	}
	serveCmd.Flags().StringVar(&b.serveConfig.DSN, "dsn", b.serveConfig.DSN, "PostgreSQL DSN for serve command")
	serveCmd.Flags().StringVar(&b.serveConfig.Port, "port", b.serveConfig.Port, "HTTP listen port (for example :8080)")
	serveCmd.Flags().StringSliceVar(&b.serveConfig.CORSOrigins, "cors-origins", b.serveConfig.CORSOrigins, "Allowed CORS origins")
	serveCmd.Flags().StringVar(&b.serveConfig.Environment, "app-env", b.serveConfig.Environment, "Application environment (development|production)")
	serveCmd.Flags().StringVar(&b.serveConfig.LogLevel, "log-level", b.serveConfig.LogLevel, "Zap log level (debug|info|warn|error)")

	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migration command",
		RunE: func(_ *cobra.Command, _ []string) error {
			return migrate.Run(b.migrateConfig)
		},
	}
	migrateCmd.Flags().StringVar(&b.migrateConfig.DatabaseURL, "database-url", b.migrateConfig.DatabaseURL, "Database connection URL")
	migrateCmd.Flags().StringVar(&b.migrateConfig.MigrationsPath, "migrations-path", b.migrateConfig.MigrationsPath, "Path to migrations directory")
	migrateCmd.Flags().StringVar(&b.migrateConfig.Command, "command", b.migrateConfig.Command, "Migration command: up, down, version, force")
	migrateCmd.Flags().IntVar(&b.migrateConfig.Steps, "steps", b.migrateConfig.Steps, "Number of migration steps for up/down")
	migrateCmd.Flags().IntVar(&b.migrateConfig.Version, "version", b.migrateConfig.Version, "Target version for force command")

	rootCmd.AddCommand(serveCmd, migrateCmd)

	return rootCmd
}
