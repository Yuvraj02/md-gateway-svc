package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/marketing-digest/pkg/logger"

	grpcclients "github.com/marketing-digest/gateway/internal/clients/grpc"
	"github.com/marketing-digest/gateway/internal/config"
	httpserver "github.com/marketing-digest/gateway/internal/transport/http"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "gateway fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	log := logger.New(cfg.Service)
	log.Info("starting", "env", cfg.AppEnv, "http_port", cfg.HTTPPort)

	clients, err := grpcclients.Dial(cfg.AuthServiceGRPCAddr, cfg.BlogServiceGRPCAddr)
	if err != nil {
		return fmt.Errorf("grpc clients: %w", err)
	}
	defer func() {
		if cerr := clients.Close(); cerr != nil {
			log.Error("grpc clients close", "error", cerr)
		}
	}()

	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	srv := httpserver.New(addr, clients, log, cfg.CORSAllowedOrigins, cfg.OwnerStudioSecret)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-sigCh:
		log.Info("shutdown signal", "signal", sig.String())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("http shutdown: %w", err)
	}
	log.Info("shutdown complete")
	return nil
}
