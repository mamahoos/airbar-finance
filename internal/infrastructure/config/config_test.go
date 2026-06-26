package config

import (
	"testing"
)

func TestLoadRequiredFields(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("REDIS_URL", "")
	t.Setenv("FINANCE_PUBLIC_BASE_URL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when required env vars are missing")
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://airbar:airbar@localhost:5434/airbar_finance?sslmode=disable")
	t.Setenv("REDIS_URL", "redis://localhost:6379/1")
	t.Setenv("FINANCE_PUBLIC_BASE_URL", "http://localhost:8080")
	t.Setenv("ZIBAL_SANDBOX", "true")
	t.Setenv("GRPC_PORT", "")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("PLATFORM_FEE_PERCENT", "")

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
}
