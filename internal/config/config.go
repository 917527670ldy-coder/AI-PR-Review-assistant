package config

import (
	"os"
	"strconv"
)

// Config 应用程序的所有配置
type Config struct {
	// 服务器配置
	ServerPort string

	// GitHub 配置
	GitHubToken         string
	GitHubWebhookSecret string

	// AI 配置
	ClaudeAPIKey string

	// 数据库配置
	DatabaseURL string

	// Redis 配置
	RedisURL string
}

// Load 从环境变量读取配置
func Load() *Config {
	return &Config{
		ServerPort:          getEnv("SERVER_PORT", "8081"),
		GitHubToken:         os.Getenv("GITHUB_TOKEN"),
		GitHubWebhookSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
		ClaudeAPIKey:        os.Getenv("CLAUDE_API_KEY"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/xengineer?sslmode=disable"),
		RedisURL:            getEnv("REDIS_URL", "redis://localhost:6379"),
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为整数，如果不存在则返回默认值
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
