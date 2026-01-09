package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a temporary .env file for testing
	envContent := "WEB=:8080\nREDIS=localhost:6379\nREDIS_USER=testuser\nREDIS_PASS=testpass\nRDAP_API=https://rdap.db.ripe.net/ip/{REMOTE_IP}\n"
	tmpEnv := ".env.test"
	err := os.WriteFile(tmpEnv, []byte(envContent), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpEnv)

	// Mock envFileName for test
	originalEnvFileName := envFileName
	envFileName = tmpEnv
	defer func() { envFileName = originalEnvFileName }()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.WebAddr != ":8080" {
		t.Errorf("WebAddr expected :8080, got %s", cfg.WebAddr)
	}
	if cfg.Redis != "localhost:6379" {
		t.Errorf("Redis expected localhost:6379, got %s", cfg.Redis)
	}
	if cfg.RedisUser != "testuser" {
		t.Errorf("RedisUser expected testuser, got %s", cfg.RedisUser)
	}
	if cfg.RedisPass != "testpass" {
		t.Errorf("RedisPass expected testpass, got %s", cfg.RedisPass)
	}
}

func TestLoadEnvFile(t *testing.T) {
	content := "KEY1=VALUE1\n# Comment\n  KEY2 = VALUE2  \n"
	tmpFile := "test.env"
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile)

	err = loadEnvFile(tmpFile)
	if err != nil {
		t.Errorf("loadEnvFile failed: %v", err)
	}

	if os.Getenv("KEY1") != "VALUE1" {
		t.Errorf("KEY1 expected VALUE1, got %s", os.Getenv("KEY1"))
	}
	if os.Getenv("KEY2") != "VALUE2" {
		t.Errorf("KEY2 expected VALUE2, got %s", os.Getenv("KEY2"))
	}
}
