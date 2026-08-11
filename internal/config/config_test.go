package config_test

import (
	"testing"

	"github.com/marketing-digest/gateway/internal/config"
)

func TestLoadRequiresGRPCAddrs(t *testing.T) {
	t.Setenv("AUTH_SERVICE_GRPC_ADDR", "")
	t.Setenv("BLOG_SERVICE_GRPC_ADDR", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadOK(t *testing.T) {
	t.Setenv("AUTH_SERVICE_GRPC_ADDR", "localhost:50051")
	t.Setenv("BLOG_SERVICE_GRPC_ADDR", "localhost:50052")
	t.Setenv("HTTP_PORT", "8080")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Service != "gateway" {
		t.Fatalf("%+v", cfg)
	}
}
