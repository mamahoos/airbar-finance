package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds environment-driven settings for the finance service.
type Config struct {
	DatabaseURL          string
	RedisURL             string
	GRPCPort             int
	HTTPPort             int
	ZibalSandbox         bool
	ZibalMerchant        string
	FinancePublicBaseURL string
	PlatformFeePercent   float64
}

// Load reads configuration from environment variables.
func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:          os.Getenv("DATABASE_URL"),
		RedisURL:             os.Getenv("REDIS_URL"),
		FinancePublicBaseURL: os.Getenv("FINANCE_PUBLIC_BASE_URL"),
		ZibalMerchant:        os.Getenv("ZIBAL_MERCHANT"),
	}

	if cfg.DatabaseURL == "" {
		return cfg, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.RedisURL == "" {
		return cfg, fmt.Errorf("REDIS_URL is required")
	}
	if cfg.FinancePublicBaseURL == "" {
		return cfg, fmt.Errorf("FINANCE_PUBLIC_BASE_URL is required")
	}

	grpcPort, err := envInt("GRPC_PORT", 50051)
	if err != nil {
		return cfg, err
	}
	httpPort, err := envInt("HTTP_PORT", 8080)
	if err != nil {
		return cfg, err
	}
	cfg.GRPCPort = grpcPort
	cfg.HTTPPort = httpPort

	zibalSandbox, err := envBool("ZIBAL_SANDBOX", false)
	if err != nil {
		return cfg, err
	}
	cfg.ZibalSandbox = zibalSandbox

	if cfg.ZibalMerchant == "" {
		if cfg.ZibalSandbox {
			cfg.ZibalMerchant = "zibal"
		} else {
			return cfg, fmt.Errorf("ZIBAL_MERCHANT is required when ZIBAL_SANDBOX is false")
		}
	}

	feePercent, err := envFloat("PLATFORM_FEE_PERCENT", 10)
	if err != nil {
		return cfg, err
	}
	cfg.PlatformFeePercent = feePercent

	return cfg, nil
}

func envInt(key string, defaultValue int) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return value, nil
}

func envFloat(key string, defaultValue float64) (float64, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue, nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number: %w", key, err)
	}
	return value, nil
}

func envBool(key string, defaultValue bool) (bool, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", key, err)
	}
	return value, nil
}
