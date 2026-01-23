package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/cobra"

	migrateCmd "github.com/Testzyler/go-microservice/cmd/migrate"
	authSchedulerCmd "github.com/Testzyler/go-microservice/services/auth/cmd/scheduler"
	authCmd "github.com/Testzyler/go-microservice/services/auth/cmd/serve"
	authWorkersCmd "github.com/Testzyler/go-microservice/services/auth/cmd/workers"
	gatewayCmd "github.com/Testzyler/go-microservice/services/internal-gateway/cmd/serve"
)

func main() {
	root := &cobra.Command{
		Use:   "go-microservice",
		Short: "Go microservice template entrypoint",
		Run: func(cmd *cobra.Command, args []string) {
			runGateway()
		},
	}
	root.AddCommand(
		gatewayCmd.NewCommand(),
		authCmd.NewCommand(),
		authWorkersCmd.NewCommand(),
		authSchedulerCmd.NewCommand(),
		newServeAllCommand(),
		migrateCmd.NewCommand(),
	)

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "command error: %v\n", err)
		os.Exit(1)
	}
}

func runGateway() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("🚀 Starting internal gateway for local development...")
	fmt.Println("   Press Ctrl+C to stop")
	fmt.Println("")

	// Wait for signal or error
	select {
	case sig := <-sigChan:
		fmt.Printf("\n⏹ Received signal %v, shutting down gateway...\n", sig)
		cancel()
	}

	if err := gatewayCmd.NewCommand().ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Gateway error: %v\n", err)
	}
}

func newServeAllCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "serve-all",
		Short: "Run gateway, auth API, auth workers, and scheduler together for local debugging",
		Run: func(cmd *cobra.Command, args []string) {
			runAllServices()
		},
	}
}

func runAllServices() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	type svc struct {
		name string
		cmd  *cobra.Command
		args []string
	}
	services := []svc{
		{name: "gateway", cmd: gatewayCmd.NewCommand()},
		{name: "auth-api", cmd: authCmd.NewCommand()},
		{name: "auth-worker-mailer", cmd: authWorkersCmd.NewCommand(), args: []string{"mailer"}},
		{name: "auth-worker-auditor", cmd: authWorkersCmd.NewCommand(), args: []string{"auditor"}},
		{name: "auth-worker-cleanup", cmd: authWorkersCmd.NewCommand(), args: []string{"cleanup"}},
		{name: "auth-scheduler", cmd: authSchedulerCmd.NewCommand()},
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(services))

	fmt.Println("🚀 Starting all services (gateway, auth API, workers, scheduler)...")
	fmt.Println("   Press Ctrl+C to stop all")
	fmt.Println("")

	for _, svc := range services {
		wg.Add(1)
		go func(name string, c *cobra.Command, args []string) {
			defer wg.Done()
			fmt.Printf("▶ Starting %s...\n", name)
			if len(args) > 0 {
				c.SetArgs(args)
			}
			if err := c.ExecuteContext(ctx); err != nil {
				errChan <- fmt.Errorf("%s: %w", name, err)
			}
		}(svc.name, svc.cmd, svc.args)
	}

	select {
	case sig := <-sigChan:
		fmt.Printf("\n⏹ Received signal %v, shutting down all services...\n", sig)
		cancel()
	case err := <-errChan:
		fmt.Printf("\n❌ Service error: %v\n", err)
		cancel()
	}

	wg.Wait()
	fmt.Println("✅ All services stopped")
}
