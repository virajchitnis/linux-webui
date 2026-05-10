# CLAUDE.md — linux-admin Codebase Guide

## Project Overview

linux-admin is a modern Go + React web panel for Linux server administration targeting Ubuntu 24.04/26.04 with systemd. It replaces a PHP/Apache Gentoo-specific panel. The backend is a single self-contained Go binary that embeds the React frontend via `go:embed`. It communicates with systemd via D-Bus (no sudo for service management), reads metrics directly from `/proc`, and exposes both a REST API and WebSocket channels for real-time data.

## Build & Run

```bash
make build          # vite build → ui/dist + go build -tags production → ./linux-admin
make dev            # vite dev server (:5173) + go run ./cmd/linux-admin (LINUX_ADMIN_DEV=1)
make test           # go test ./... + go vet ./...
make lint           # golangci-lint run
```

The `LINUX_ADMIN_DEV=1` env var makes Vite proxy `/api/*` and `/ws/*` to the Go server at `:8443`. The `-tags production` build tag compiles out the dev proxy code via `//go:build !production`.

## Key Design Decisions

**No `sh -c` ever.** Every `os/exec` call uses `[]string` args. Never interpolate user input into a string that becomes a shell command. See `internal/firewall/ufw.go` and `internal/users/users.go` for examples of how to construct commands with validated components.

**D-Bus for systemd (not sudo).** The polkit rule at `/etc/polkit-1/rules.d/50-linux-admin.rules` grants `linux-admin` user `org.freedesktop.systemd1.manage-units` without password. `internal/dbus/` wraps `godbus/dbus/v5`. This means no shell process is spawned to start/stop services.

**SQLite for all transient state.** Sessions, audit log, brute-force counters, TOTP recovery codes — all in `internal/auth/` using `modernc.org/sqlite` (pure Go, no CGO). WAL mode, single writer (`SetMaxOpenConns(1)`).

**Frontend embedded at compile time.** `ui/embed.go` uses `//go:embed all:dist` to bundle the Vite output. In dev mode the file is excluded by build tag and replaced with a reverse proxy to Vite.

**Capability detection at runtime.** `internal/api/handlers/capabilities.go` probes for binaries and features every 5 minutes. The frontend reads `/api/capabilities` once on load and hides nav items for unavailable features. Never hardcode which features are present.

**PTY read deadline pattern.** `internal/terminal/manager.go` sets a 300ms read deadline before every PTY read so that context cancellation exits the read loop promptly without goroutine leaks.

## Package Map

| Package | Description |
|---|---|
| `cmd/linux-admin` | Entry point: config loading, feature detection, server startup, graceful shutdown |
| `internal/api` | chi router (`router.go`), `RouterOptions` struct |
| `internal/api/handlers` | One file per route group (auth, services, apt, journal, terminal, process, users, firewall, account, totp, files, cron, netif, sensors, ai, auditlog, capabilities) |
| `internal/api/middleware` | Auth session middleware, CSRF validation, rate limiter, security headers, ClientIP |
| `internal/auth` | SQLite schema+migrations, sessions, users, bcrypt, CSRF tokens, brute-force, TOTP/2FA, audit log |
| `internal/config` | TOML config loader with defaults |
| `internal/distro` | `/etc/os-release` parser → `Distro` struct + `PackageFamily` enum |
| `internal/dbus` | D-Bus client wrapping `godbus/dbus/v5` for systemd unit management |
| `internal/exec` | `Resolve(name)` — `exec.LookPath` wrapper returning `(bool, string)` |
| `internal/metrics` | `/proc` collector goroutine, `Snapshot` struct, ring buffer, subscriber pattern |
| `internal/journal` | `journalctl` streaming to channel; JSON line parser handles byte-array MESSAGE field |
| `internal/terminal` | PTY session lifecycle, reaper goroutine, 300ms read deadline pattern |
| `internal/apt` | `apt-get` wrapper with single-job mutex; output streamed to `chan<- string` |
| `internal/firewall` | UFW status parse, add/delete rule, enable/disable |
| `internal/users` | `/etc/passwd` + `/etc/group` parser; `useradd`/`userdel`/`chpasswd` wrappers |
| `internal/process` | `/proc` enumeration, CPU% calculation, `Kill`/`Renice` with pid validation |
| `internal/files` | Read-only file browser, `Roots.Jail()` path escape prevention via `EvalSymlinks` |
| `internal/cron` | `crontab -l/-` parser/writer, schedule + command validation |
| `internal/netif` | Network interface stats from `/proc/net/dev` + `net.Interfaces()` |
| `internal/sensors` | hwmon sysfs reader for CPU temps (millidegrees → °C) and fan RPMs |
| `internal/ollama` | Streaming Ollama REST client, system prompt builder with live server context |
| `internal/ws` | (reserved for typed envelope hub — not yet extracted) |
| `ui/embed.go` | `go:embed all:dist` — serves React frontend |

## Adding a New Module

1. Create `internal/<module>/<module>.go`
2. Write handler(s) in `internal/api/handlers/<module>.go`
3. Register routes in `internal/api/router.go` inside the auth group; gate mutations with `r.With(adminMW)`
4. Add capability probe in `internal/api/handlers/capabilities.go` (new key in `caps` map)
5. Add frontend page at `web/src/pages/<Module>.tsx`
6. Add nav item in `web/src/components/Sidebar.tsx` with `enabled: caps?.<key> ?? false`
7. Add route in `web/src/App.tsx`
8. Add E2E test at `e2e/<module>.spec.ts`

## Auth & RBAC

Sessions are 256-bit random hex IDs in the SQLite `sessions` table. `middleware.RequireAuth` validates the `linux_admin_session` cookie on every request and stores the `*auth.Session` in the request context. `middleware.SessionFromContext(ctx)` retrieves it.

Roles: `admin` (full access) and `readonly` (GET endpoints only). `middleware.RequireAdmin` returns 403 if the session role is not `admin`. Wrap mutation routes with `r.With(adminMW)`.

CSRF: every state-changing request (POST/PUT/DELETE) must include the `X-CSRF-Token` header matching the `linux_admin_csrf` cookie. Validated in `middleware.RequireCSRF`. The cookie is set on login and is readable by JS (not HttpOnly).

TOTP: if `users.totp_secret` is non-empty, the login handler requires a `totp_code` field. Recovery codes are bcrypt-hashed in `totp_recovery` and consumed one-at-a-time. Enrollment is a two-step HTTP exchange: `POST /api/account/totp/enroll` (get secret+QR URL) → `POST /api/account/totp/confirm` (verify code → save secret + generate recovery codes).

## WebSocket Protocol

All WS endpoints send/receive JSON. The `coder/websocket` library is used throughout (gorilla/websocket is unmaintained). Example pattern (metrics):

```go
conn, err := websocket.Accept(w, r, nil)
// ...
ch := collector.Subscribe()
defer collector.Unsubscribe(ch)
for snap := range ch {
    wsjson.Write(ctx, conn, snap)
}
```

For bidirectional channels (terminal, AI), use `conn.CloseRead(ctx)` to get a context that cancels on WS close, then use `wsjson.Read` in a loop for inbound and `wsjson.Write` for outbound. See `internal/api/handlers/terminal.go` and `internal/api/handlers/ai.go`.

## Security Checklist (for every PR)

1. No `sh -c` or string interpolation into exec args
2. All user-supplied identifiers (package names, usernames, pids, indexes) validated before use
3. File paths run through `files.Roots.Jail()` before open/read
4. Mutation routes wrapped with `adminMW` where required
5. WS upgrade validates session cookie (done by `authMW` group)
6. No internal paths, stack traces, or session IDs logged or returned to client
7. Body size limited via `http.MaxBytesReader` at every JSON decode site
8. No new binary executed with `shell=true` or equivalent
9. New audit log entries for state-changing actions (`auth.LogAction`)
10. CSRF header required for all POST/DELETE endpoints (enforced by `RequireCSRF`)

## Common Pitfalls

- **Never use `filepath.Join` for URL paths** — use `path.Join` or string concatenation
- **Don't return `err.Error()` from internal packages to HTTP clients** — it can leak paths, SQL, and internal state. Return a generic message and log internally.
- **Never log session IDs** — they are equivalent to passwords
- **SQLite single writer** — `db.SetMaxOpenConns(1)` is set; don't open a second `sql.DB` on the same file
- **The PTY read deadline must be reset before every read** — not just once. See `terminal/manager.go`
- **`EvalSymlinks` returns an error if the path doesn't exist** — in `files.Jail()` we allow non-existent paths for listing (the OS will return a proper error on open), but only after passing the prefix check on the cleaned (unresolved) path
- **D-Bus unit names are validated against the live unit list**, not a static allowlist — this handles services with special characters

## Running Tests

```bash
go test ./...           # all unit tests
go vet ./...            # static analysis
cd web && npm run build # TypeScript type check + Vite build
```

E2E tests (Playwright) require a running instance at `https://localhost:8443`:
```bash
make e2e
```
