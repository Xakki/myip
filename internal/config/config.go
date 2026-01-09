package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var envFileName = ".env"

// Config holds service settings loaded from the environment.
type Config struct {
	WebAddr   string
	Redis     string
	RedisUser string
	RedisPass string
	RDAPAPI   string
	LogType   string
	LogAddr   string
}

// Load reads .env and merges it with existing environment values.
func Load() (Config, error) {
	if err := loadEnvFile(envFileName); err != nil {
		return Config{}, fmt.Errorf("load env file: %w", err)
	}

	cfg := Config{
		WebAddr:   strings.TrimSpace(os.Getenv("WEB")),
		Redis:     strings.TrimSpace(os.Getenv("REDIS")),
		RedisUser: strings.TrimSpace(os.Getenv("REDIS_USER")),
		RedisPass: strings.TrimSpace(os.Getenv("REDIS_PASS")),
		RDAPAPI:   strings.TrimSpace(os.Getenv("RDAP_API")),
		LogType:   strings.TrimSpace(os.Getenv("LOG_TYPE")),
		LogAddr:   strings.TrimSpace(os.Getenv("LOG_ADDR")),
	}

	if cfg.LogType == "" {
		cfg.LogType = "console"
	}

	if cfg.WebAddr == "" {
		return Config{}, fmt.Errorf("WEB is required")
	}
	if cfg.Redis == "" {
		return Config{}, fmt.Errorf("REDIS is required")
	}

	return cfg, nil
}

func loadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("invalid env line: %s", line)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			return fmt.Errorf("empty env key in line: %s", line)
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set env %s: %w", key, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan env file: %w", err)
	}
	return nil
}
