package workers

import (
	"github.com/spf13/cobra"

	auditorcmd "github.com/Testzyler/go-microservice/services/auth/cmd/workers/auditor"
	cleanupcmd "github.com/Testzyler/go-microservice/services/auth/cmd/workers/cleanup"
	mailercmd "github.com/Testzyler/go-microservice/services/auth/cmd/workers/mailer"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve-auth-worker",
		Short: "Run auth workers (mailer, auditor, cleanup)",
	}
	cmd.AddCommand(
		mailercmd.NewCommand(),
		auditorcmd.NewCommand(),
		cleanupcmd.NewCommand(),
	)
	return cmd
}
