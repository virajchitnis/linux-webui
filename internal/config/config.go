package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server     ServerConfig     `toml:"server"`
	Auth       AuthConfig       `toml:"auth"`
	Features   FeaturesConfig   `toml:"features"`
	Ollama     OllamaConfig     `toml:"ollama"`
	Monitoring MonitoringConfig `toml:"monitoring"`
	Audit      AuditConfig      `toml:"audit"`
}

type ServerConfig struct {
	Listen      string `toml:"listen"`
	TLSCert     string `toml:"tls_cert"`
	TLSKey      string `toml:"tls_key"`
	ACMEDomain  string `toml:"acme_domain"`
	HTTPListen  string `toml:"http_listen"`
}

type AuthConfig struct {
	SessionTimeoutMinutes int `toml:"session_timeout_minutes"`
	MaxLoginAttempts      int `toml:"max_login_attempts"`
	LockoutMinutes        int `toml:"lockout_minutes"`
	BcryptCost            int `toml:"bcrypt_cost"`
	TerminalReconnectSecs int `toml:"terminal_reconnect_seconds"`
	MaxWSConnsPerIP       int `toml:"max_ws_conns_per_ip"`
}

type FeaturesConfig struct {
	TerminalEnabled     bool     `toml:"terminal_enabled"`
	FileBrowserEnabled  bool     `toml:"file_browser_enabled"`
	FileBrowserRoots    []string `toml:"file_browser_roots"`
	MaxTerminalSessions int      `toml:"max_terminal_sessions"`
}

type OllamaConfig struct {
	Enabled  bool   `toml:"enabled"`
	Endpoint string `toml:"endpoint"`
	Model    string `toml:"model"`
}

type MonitoringConfig struct {
	Prometheus bool `toml:"prometheus"`
}

type AuditConfig struct {
	RetentionDays int `toml:"retention_days"`
}

func Defaults() *Config {
	return &Config{
		Server: ServerConfig{
			Listen:     "0.0.0.0:8443",
			TLSCert:    "/etc/linux-admin/tls/cert.pem",
			TLSKey:     "/etc/linux-admin/tls/key.pem",
			HTTPListen: "0.0.0.0:8080",
		},
		Auth: AuthConfig{
			SessionTimeoutMinutes: 30,
			MaxLoginAttempts:      5,
			LockoutMinutes:        15,
			BcryptCost:            12,
			TerminalReconnectSecs: 30,
			MaxWSConnsPerIP:       20,
		},
		Features: FeaturesConfig{
			TerminalEnabled:     true,
			FileBrowserEnabled:  true,
			FileBrowserRoots:    []string{"/etc", "/var/log", "/home"},
			MaxTerminalSessions: 3,
		},
		Ollama: OllamaConfig{
			Enabled:  false,
			Endpoint: "http://127.0.0.1:11434",
			Model:    "llama3.2",
		},
		Monitoring: MonitoringConfig{
			Prometheus: false,
		},
		Audit: AuditConfig{
			RetentionDays: 90,
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := Defaults()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
