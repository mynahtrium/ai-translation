package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	GatewayPort     int
	ASRPort         int
	TranslatorPort  int
	TTSPort         int
	ASRAddress      string
	TranslatorAddr  string
	TTSAddress      string
	GeminiAPIKey    string
	GCPProjectID    string
	GCPCredentials  string
	LogLevel        string
	ShutdownTimeout time.Duration
}

func Load() *Config {
	return &Config{
		GatewayPort:     getEnvInt("GATEWAY_PORT", 8080),
		ASRPort:         getEnvInt("ASR_PORT", 50051),
		TranslatorPort:  getEnvInt("TRANSLATOR_PORT", 50052),
		TTSPort:         getEnvInt("TTS_PORT", 50053),
		ASRAddress:      getEnv("ASR_ADDRESS", "localhost:50051"),
		TranslatorAddr:  getEnv("TRANSLATOR_ADDRESS", "localhost:50052"),
		TTSAddress:      getEnv("TTS_ADDRESS", "localhost:50053"),
		GeminiAPIKey:    getEnv("GEMINI_API_KEY", ""),
		GCPProjectID:    getEnv("GCP_PROJECT_ID", ""),
		GCPCredentials:  getEnv("GOOGLE_APPLICATION_CREDENTIALS", ""),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		ShutdownTimeout: time.Duration(getEnvInt("SHUTDOWN_TIMEOUT_SEC", 30)) * time.Second,
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
