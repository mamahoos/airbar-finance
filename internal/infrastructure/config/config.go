package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds environment-driven settings for the finance service.
// All values are loaded from environment variables (or a local .env file).
// Production deployments must inject env via the orchestrator — never rely on defaults in code.
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

// Load reads configuration from the process environment.
// When a .env file exists (local dev), it is loaded first without overriding existing env vars.
func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{}

	databaseURL, err := requireString("DATABASE_URL")
	if err != nil {
		return cfg, err
	}
	cfg.DatabaseURL = databaseURL

	redisURL, err := requireString("REDIS_URL")
	if err != nil {
		return cfg, err
	}
	cfg.RedisURL = redisURL

	grpcPort, err := requireInt("GRPC_PORT")
	if err != nil {
		return cfg, err
	}
	cfg.GRPCPort = grpcPort

	httpPort, err := requireInt("HTTP_PORT")
	if err != nil {
		return cfg, err
	}
	cfg.HTTPPort = httpPort

	zibalSandbox, err := requireBool("ZIBAL_SANDBOX")
	if err != nil {
		return cfg, err
	}
	cfg.ZibalSandbox = zibalSandbox

	zibalMerchant, err := requireString("ZIBAL_MERCHANT")
	if err != nil {
		return cfg, err
	}
	cfg.ZibalMerchant = zibalMerchant

	financePublicBaseURL, err := requireString("FINANCE_PUBLIC_BASE_URL")
	if err != nil {
		return cfg, err
	}
	cfg.FinancePublicBaseURL = financePublicBaseURL

	feePercent, err := requireFloat("PLATFORM_FEE_PERCENT")
	if err != nil {
		return cfg, err
	}
	cfg.PlatformFeePercent = feePercent

	return cfg, nil
}

func requireString(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return value, nil
}

func requireInt(key string) (int, error) {
	raw, err := requireString(key)
	if err != nil {
		return 0, err
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return value, nil
}

func requireFloat(key string) (float64, error) {
	raw, err := requireString(key)
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number: %w", key, err)
	}
	return value, nil
}

func requireBool(key string) (bool, error) {
	raw, err := requireString(key)
	if err != nil {
		return false, err
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", key, err)
	}
	return value, nil
}
