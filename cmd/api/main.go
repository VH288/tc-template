package main

import (
	"context"
	"fmt"
	"os"

	"github.com/VH288/tc-template/internal/app/bootstrap"
	"github.com/VH288/tc-template/internal/config"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Get Config
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	container, err := bootstrap.Build(ctx, cfg)
	if err != nil {
		return err
	}

	_ = container

	return nil
}
