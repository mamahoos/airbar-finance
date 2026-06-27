package config

import (
	"testing"
)

func setValidEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://airbar:airbar@localhost:5434/airbar_finance?sslmode=disable")
	t.Setenv("REDIS_URL", "redis://localhost:6381/1")
	t.Setenv("GRPC_PORT", "50051")
	t.Setenv("HTTP_PORT", "8080")
	t.Setenv("ZIBAL_SANDBOX", "true")
	t.Setenv("ZIBAL_MERCHANT", "zibal")
	t.Setenv("FINANCE_PUBLIC_BASE_URL", "http://localhost:8080")
	t.Setenv("PLATFORM_FEE_PERCENT", "10")
}

func TestLoadRequiredFields(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("REDIS_URL", "")
	t.Setenv("GRPC_PORT", "")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("ZIBAL_SANDBOX", "")
	t.Setenv("ZIBAL_MERCHANT", "")
	t.Setenv("FINANCE_PUBLIC_BASE_URL", "")
	t.Setenv("PLATFORM_FEE_PERCENT", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when required env vars are missing")
	}
}

func TestLoadFromEnv(t *testing.T) {
	setValidEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.GRPCPort != 50051 {
		t.Fatalf("GRPCPort = %d, want 50051", cfg.GRPCPort)
	}
	if cfg.HTTPPort != 8080 {
		t.Fatalf("HTTPPort = %d, want 8080", cfg.HTTPPort)
	}
	if cfg.ZibalMerchant != "zibal" {
		t.Fatalf("ZibalMerchant = %q, want zibal", cfg.ZibalMerchant)
	}
	if cfg.PlatformFeePercent != 10 {
		t.Fatalf("PlatformFeePercent = %v, want 10", cfg.PlatformFeePercent)
	}
	if !cfg.ZibalSandbox {
		t.Fatal("ZibalSandbox = false, want true")
	}
}
