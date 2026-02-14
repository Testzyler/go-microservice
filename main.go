package main

import (
	"errors"
	"log"
	"os"

	"github.com/sklinkert/go-ddd/cmd/root"
	"github.com/sklinkert/go-ddd/internal/config/envfile"
)

func main() {
	if err := envfile.Load(".env"); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("failed to load .env: %v", err)
	}

	rootCmd := root.NewCommandBuilder().Build()
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
