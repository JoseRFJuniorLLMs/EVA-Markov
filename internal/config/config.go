package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database
	DatabaseURL string

	// Google Gemini
	GoogleAPIKey string

	// Scheduler
	CronSchedule          string
	AnalysisLookbackHours int

	// Meta-Agent
	MinConversationsForOptimization int
	OptimizationThresholdScore      float64
	MaxPromptIterations             int

	// Logging
	LogLevel  string
	LogFormat string

	// Environment
	Env string
}

func Load() (*Config, error) {
	// Carregar .env se existir
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL:                     getEnv("DATABASE_URL", ""),
		GoogleAPIKey:                    getEnv("GOOGLE_API_KEY", ""),
		CronSchedule:                    getEnv("CRON_SCHEDULE", "0 23 * * *"),
		AnalysisLookbackHours:           getEnvAsInt("ANALYSIS_LOOKBACK_HOURS", 24),
		MinConversationsForOptimization: getEnvAsInt("MIN_CONVERSATIONS_FOR_OPTIMIZATION", 5),
		OptimizationThresholdScore:      getEnvAsFloat("OPTIMIZATION_THRESHOLD_SCORE", 7.0),
		MaxPromptIterations:             getEnvAsInt("MAX_PROMPT_ITERATIONS", 3),
		LogLevel:                        getEnv("LOG_LEVEL", "info"),
		LogFormat:                       getEnv("LOG_FORMAT", "json"),
		Env:                             getEnv("ENV", "development"),
	}

	// Validações
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL é obrigatório")
	}
	if cfg.GoogleAPIKey == "" {
		return nil, fmt.Errorf("GOOGLE_API_KEY é obrigatório")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultValue
}
