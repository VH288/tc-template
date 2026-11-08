package main

import (
	"context"
	"fmt"
	"os"

	"github.com/VH288/tc-template/internal/app/bootstrap"
	grpcapp "github.com/VH288/tc-template/internal/app/grpc"
	"github.com/VH288/tc-template/internal/app/server"
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

	httpSrv := server.NewHTTP(
		":"+cfg.HTTP.Port,
		container.HTTPRouter(),
		container.Logger,
		cfg.HTTP.ReadTimeout,
		cfg.HTTP.WriteTimeout,
	)
	grpcServer := grpcapp.NewServer(grpcapp.Dependencies{
		Sample: container.Domain.Sample.GRPC,
	})

	grpcSrv, err := server.NewGRPC(":"+cfg.GRPC.Port, grpcServer, container.Logger)
	if err != nil {
		return err
	}

	errCh := make(chan error, 2)
	go func() { errCh <- httpSrv.Start() }()
	go func() { errCh <- grpcSrv.Start() }()

	container.Lifecycle.Add(func(ctx context.Context) error {
		return httpSrv.Shutdown(ctx)
	})
	container.Lifecycle.Add(func(ctx context.Context) error {
		grpcSrv.GracefulStop()
		return nil
	})
	go func() {
		if err := <-errCh; err != nil {
			container.Logger.Error("server error", "error", err)
			cancel()
		}
	}()

	container.Lifecycle.Wait(ctx)
	return nil
}
