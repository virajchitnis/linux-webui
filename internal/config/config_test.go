package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := Defaults()
	if cfg.Server.Listen != "0.0.0.0:8443" {
		t.Errorf("default listen = %q, want 0.0.0.0:8443", cfg.Server.Listen)
	}
	if cfg.Auth.SessionTimeoutMinutes != 30 {
		t.Errorf("default session timeout = %d, want 30", cfg.Auth.SessionTimeoutMinutes)
	}
	if cfg.Auth.BcryptCost != 12 {
		t.Errorf("default bcrypt cost = %d, want 12", cfg.Auth.BcryptCost)
	}
	if cfg.Auth.MaxLoginAttempts != 5 {
		t.Errorf("default max login attempts = %d, want 5", cfg.Auth.MaxLoginAttempts)
	}
	if cfg.Ollama.Model != "llama3.2" {
		t.Errorf("default ollama model = %q, want llama3.2", cfg.Ollama.Model)
	}
	if cfg.Audit.RetentionDays != 90 {
		t.Errorf("default retention days = %d, want 90", cfg.Audit.RetentionDays)
	}
}

func TestLoad_NonExistent(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.toml")
	if err != nil {
		t.Fatalf("Load(nonexistent) error: %v", err)
	}
	// Should return defaults.
	if cfg.Server.Listen != "0.0.0.0:8443" {
		t.Error("nonexistent config should return defaults")
	}
}

func TestLoad_ValidTOML(t *testing.T) {
	f, err := os.CreateTemp("", "config-*.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	_, _ = f.WriteString(`
[server]
listen = "0.0.0.0:9443"

[auth]
session_timeout_minutes = 60
bcrypt_cost = 10

[ollama]
enabled = true
model = "mistral"
`)
	f.Close()

	cfg, err := Load(f.Name())
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.Server.Listen != "0.0.0.0:9443" {
		t.Errorf("server.listen = %q, want 0.0.0.0:9443", cfg.Server.Listen)
	}
	if cfg.Auth.SessionTimeoutMinutes != 60 {
		t.Errorf("auth.session_timeout_minutes = %d, want 60", cfg.Auth.SessionTimeoutMinutes)
	}
	if cfg.Auth.BcryptCost != 10 {
		t.Errorf("auth.bcrypt_cost = %d, want 10", cfg.Auth.BcryptCost)
	}
	if !cfg.Ollama.Enabled {
		t.Error("ollama.enabled should be true")
	}
	if cfg.Ollama.Model != "mistral" {
		t.Errorf("ollama.model = %q, want mistral", cfg.Ollama.Model)
	}
	// Unset fields should retain defaults.
	if cfg.Auth.MaxLoginAttempts != 5 {
		t.Errorf("max_login_attempts should default to 5, got %d", cfg.Auth.MaxLoginAttempts)
	}
}

func TestLoad_InvalidTOML(t *testing.T) {
	f, err := os.CreateTemp("", "config-bad-*.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	_, _ = f.WriteString("this is [not valid toml {{{{")
	f.Close()

	_, err = Load(f.Name())
	if err == nil {
		t.Error("Load(invalid TOML) should return error")
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.toml")
	_ = os.WriteFile(path, []byte{}, 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load(empty) error: %v", err)
	}
	// Empty file should return defaults.
	if cfg.Server.Listen != "0.0.0.0:8443" {
		t.Error("empty config should return defaults")
	}
}
