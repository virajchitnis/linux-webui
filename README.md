# linux-admin

A modern, secure web-based administration panel for Linux servers. Built with Go and React, it replaces the traditional SSH terminal workflow with a clean browser interface for daily server management tasks.

[![Go 1.22+](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE.md)

## Features

**System**
- Live dashboard: CPU, memory, disk, network I/O charts (1-second updates)
- Hardware sensor readings (CPU temperature, fan RPM)
- System reboot with countdown

**Services**
- Dynamic systemd service discovery (all installed units, not a fixed list)
- Start / stop / restart / enable / disable via D-Bus (no sudo required)
- Real-time status polling

**Networking**
- Network interface stats (addresses, RX/TX bytes, MTU, state)
- WireGuard peer management (if `wg` is installed)
- Tailscale status and peer table (if `tailscale` is installed)

**Logs & Terminal**
- Live journal log viewer with unit and priority filters
- Full web terminal (xterm.js + PTY), audit-logged

**Packages & Processes**
- APT package manager: list upgradable, apply updates with live streaming output
- Process manager: list, kill, renice

**Security & Users**
- Local user and group management
- UFW firewall rule management
- Two-factor authentication (TOTP) with QR enrollment and recovery codes
- Per-user session management with revocation
- Full audit log of all state-changing actions

**Files & Cron**
- Read-only file browser (path-jailed to `/etc`, `/var/log`, `/home`)
- Cron job editor with human-readable schedule descriptions

**AI Assistant**
- Embedded Ollama chat panel with live server context injection
- Explain log lines, analyze performance, get command suggestions
- Suggested commands are never auto-executed

**Docker** (optional)
- Container list, start/stop/restart/remove
- Real-time resource stats and log streaming

## Requirements

- Ubuntu 22.04+ / Debian 12+ / any systemd Linux distribution
- Go 1.22+ (to build from source)
- Node.js 20+ (to build from source)

## Quick Install (binary)

```bash
curl -fsSL https://github.com/virajchitnis/linux-webui/releases/latest/download/install.sh | sudo bash
```

Then visit `https://your-server:8443` and complete the first-run wizard.

## Manual Installation

1. Download the pre-built binary from [Releases](https://github.com/virajchitnis/linux-webui/releases).
2. Run the installer:
   ```bash
   sudo ./linux-admin --install
   ```
   This creates the `linux-admin` system user, polkit rule, sudoers entries, and a systemd service unit.
3. Start the service:
   ```bash
   sudo systemctl enable --now linux-admin
   ```
4. Navigate to `https://your-server:8443` and complete the setup wizard (set your admin password).

## Build from Source

```bash
git clone https://github.com/virajchitnis/linux-webui.git
cd linux-webui
make build
# Produces: ./linux-admin
```

Prerequisites: Go 1.22+, Node.js 20+, npm.

## Configuration

The config file lives at `/etc/linux-admin/config.toml` (created by the installer). All settings have sensible defaults. Example:

```toml
[server]
listen     = "0.0.0.0:8443"
tls_cert   = "/etc/linux-admin/tls/cert.pem"
tls_key    = "/etc/linux-admin/tls/key.pem"
http_listen = "0.0.0.0:8080"  # redirects to HTTPS

[auth]
session_timeout_minutes = 30
max_login_attempts      = 5
lockout_minutes         = 15
bcrypt_cost             = 12

[features]
terminal_enabled      = true
file_browser_enabled  = true
file_browser_roots    = ["/etc", "/var/log", "/home"]
max_terminal_sessions = 3

[ollama]
enabled  = false
endpoint = "http://127.0.0.1:11434"
model    = "llama3.2"

[monitoring]
prometheus = false   # expose /metrics (auth required)

[audit]
retention_days = 90  # 0 = keep forever
```

A self-signed TLS certificate is generated automatically on first run. To use Let's Encrypt, set `[server] acme_domain = "myserver.example.com"`.

## Optional Features

### AI Assistant (Ollama)

Install [Ollama](https://ollama.com) and pull a model:

```bash
curl -fsSL https://ollama.com/install.sh | sh
ollama pull llama3.2
```

The panel auto-detects Ollama on startup. Set `ollama.enabled = true` in config to force-enable it, or leave it unset for auto-detection.

### Docker

Add the `linux-admin` user to the `docker` group:

```bash
sudo usermod -aG docker linux-admin
sudo systemctl restart linux-admin
```

> **Security note**: Docker group membership is equivalent to root access. Any process in the group can escape to root via a bind mount. Consider disabling the web terminal when Docker is enabled, or restricting access to admin users only.

### WireGuard

Ensure `wg` and `wg-quick` are in PATH. The panel auto-detects and shows the WireGuard module.

### Tailscale

Ensure the `tailscale` daemon is running. The panel reads status via the local daemon socket.

### Let's Encrypt (ACME)

```toml
[server]
listen      = "0.0.0.0:443"
http_listen = "0.0.0.0:80"   # required for HTTP-01 challenge
acme_domain = "myserver.example.com"
```

The binary handles the ACME HTTP-01 challenge itself on port 80. Certificates are cached in `/etc/linux-admin/tls/` and renewed automatically.

### Prometheus

```toml
[monitoring]
prometheus = true
```

Then point Prometheus at `https://your-server:8443/metrics`. Authentication (session cookie) is required.

## User Management

Users are stored in SQLite (not the system `/etc/passwd`). To add a user via the web UI: go to **Users** → **Add User**. To manage roles, log in as admin and edit the user.

## Upgrading

```bash
# Replace the binary
sudo systemctl stop linux-admin
sudo cp linux-admin /usr/local/bin/linux-admin
sudo systemctl start linux-admin
```

Database migrations run automatically on startup — downtime is minimal.

## Security Notes

- Runs as the `linux-admin` user (non-root) with a targeted polkit rule for systemd management
- No `sh -c` anywhere — all subprocess calls use `[]string` args to prevent command injection
- TLS 1.2+ with strong cipher suites; HTTP auto-redirects to HTTPS
- CSRF protection on all state-changing requests
- Brute-force lockout (5 failures in 60 seconds → 15-minute lockout)
- Session cookies are `HttpOnly`, `Secure`, `SameSite=Strict`
- All state-changing actions written to the audit log
- Recommended: change the default port (8443) via firewall rules; restrict access to trusted IPs

## Uninstall

```bash
sudo ./linux-admin --uninstall
```

This removes the `linux-admin` user, polkit rule, sudoers entry, and systemd unit. The data directory `/etc/linux-admin/` is left intact for manual review.

## Contributing

1. Fork and clone the repository
2. Run `make dev` for hot-reload development
3. Run `make test` and `make lint` before submitting a PR
4. See `CLAUDE.md` for architecture details and the security checklist

## License

See [LICENSE.md](LICENSE.md).
