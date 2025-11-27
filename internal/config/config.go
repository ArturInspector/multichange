package config

import (
	"fmt"
	"os"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Chains   map[string]ChainConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type ChainConfig struct {
	RPCURL      string
	ChainID     int64
	MinConfirmations int
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "exchange"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Chains: make(map[string]ChainConfig),
	}

	// Load chain configs
	chains := []string{"ethereum", "polygon", "bsc", "arbitrum", "optimism"}
	for _, chain := range chains {
		rpcURL := getEnv(fmt.Sprintf("%s_RPC_URL", chain), "")
		if rpcURL != "" {
			cfg.Chains[chain] = ChainConfig{
				RPCURL:          rpcURL,
				ChainID:         getEnvInt64(fmt.Sprintf("%s_CHAIN_ID", chain), 0),
				MinConfirmations: getEnvInt(fmt.Sprintf("%s_MIN_CONFIRMATIONS", chain), 1),
			}
		}
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	// TODO: parse int from env
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	// TODO: parse int64 from env
	return defaultValue
}

