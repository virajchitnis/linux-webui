package api

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/virajchitnis/linux-webui/internal/api/handlers"
	"github.com/virajchitnis/linux-webui/internal/api/middleware"
	dbuspkg "github.com/virajchitnis/linux-webui/internal/dbus"
	"github.com/virajchitnis/linux-webui/internal/distro"
	"github.com/virajchitnis/linux-webui/internal/files"
	"github.com/virajchitnis/linux-webui/internal/metrics"
	"github.com/virajchitnis/linux-webui/internal/ollama"
	"github.com/virajchitnis/linux-webui/internal/terminal"
)

type RouterOptions struct {
	DB              *sql.DB
	Distro          *distro.Distro
	Collector       *metrics.Collector
	DBus            *dbuspkg.Client   // nil if D-Bus unavailable
	TerminalManager *terminal.Manager // nil if bash/pty unavailable
	OllamaClient    *ollama.Client    // nil if Ollama unavailable
	FilesRoots      files.Roots       // empty → file browser disabled
	Version         string
	SecureCookie    bool
	AuthTimeout     int
	BcryptCost      int
	PrometheusOn    bool
	AptEnabled      bool // Debian-family distros with apt in PATH
	JournalEnabled  bool // journalctl in PATH
	UFWEnabled      bool // ufw in PATH
	CronEnabled     bool // crontab in PATH
	SensorsEnabled  bool // /sys/class/hwmon present
	FrontendHandler http.Handler
}

func NewRouter(opts RouterOptions) http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)

	caps := handlers.InitCapabilities(opts.Distro)
	rl := middleware.NewRateLimiter(20, 100)
	r.Use(rl.Limit)
	r.Use(middleware.SecurityHeaders)

	// ── Public ────────────────────────────────────────────────────────────────
	r.Get("/health", handlers.Health)
	r.Get("/api/setup/status", (&handlers.SetupHandler{DB: opts.DB, BcryptCost: opts.BcryptCost}).Status)
	r.Post("/api/setup/complete", (&handlers.SetupHandler{DB: opts.DB, BcryptCost: opts.BcryptCost}).Complete)
	r.Post("/api/auth/login", (&handlers.AuthHandler{
		DB:             opts.DB,
		TimeoutMinutes: opts.AuthTimeout,
		BcryptCost:     opts.BcryptCost,
		SecureCookie:   opts.SecureCookie,
	}).Login)

	// ── Authenticated ─────────────────────────────────────────────────────────
	authMW := middleware.RequireAuth(middleware.Options{DB: opts.DB, TimeoutMinutes: opts.AuthTimeout})
	adminMW := middleware.RequireAdmin

	r.Group(func(r chi.Router) {
		r.Use(authMW)
		r.Use(middleware.RequireCSRF)

		r.Post("/api/auth/logout", (&handlers.AuthHandler{DB: opts.DB}).Logout)
		r.Get("/api/auth/me", (&handlers.AuthHandler{DB: opts.DB}).Me)

		sysHandler := &handlers.SystemHandler{Collector: opts.Collector, Version: opts.Version, DB: opts.DB}
		r.Get("/api/system/info", sysHandler.Info)
		r.Get("/api/system/metrics", sysHandler.Metrics)
		r.Post("/api/system/reboot", sysHandler.Reboot)
		r.Get("/api/capabilities", caps.Handler())

		// WebSocket: metrics stream
		r.Get("/ws/metrics", (&handlers.MetricsWSHandler{Collector: opts.Collector}).ServeHTTP)

		// Services (systemd via D-Bus)
		if opts.DBus != nil {
			svc := &handlers.ServicesHandler{DBus: opts.DBus, DB: opts.DB}
			r.Get("/api/services", svc.List)
			r.With(adminMW).Post("/api/services/{name}/start", svc.Start)
			r.With(adminMW).Post("/api/services/{name}/stop", svc.Stop)
			r.With(adminMW).Post("/api/services/{name}/restart", svc.Restart)
			r.With(adminMW).Post("/api/services/{name}/enable", svc.Enable)
			r.With(adminMW).Post("/api/services/{name}/disable", svc.Disable)
		}

		// APT package manager
		if opts.AptEnabled {
			apt := &handlers.AptHandler{DB: opts.DB}
			r.Get("/api/packages/upgradable", apt.Upgradable)
			r.Get("/api/packages/status", apt.Status)
			r.With(adminMW).Get("/ws/apt", apt.Stream)
		}

		// Journal log viewer
		if opts.JournalEnabled {
			r.Get("/ws/logs", (&handlers.JournalWSHandler{}).ServeHTTP)
		}

		// Web terminal (PTY)
		if opts.TerminalManager != nil {
			th := &handlers.TerminalHandler{Manager: opts.TerminalManager, DB: opts.DB}
			r.With(adminMW).Post("/api/terminal/new", th.New)
			r.With(adminMW).Get("/ws/terminal/{id}", th.Connect)
		}

		// Processes
		ph := &handlers.ProcessHandler{}
		r.Get("/api/processes", ph.List)
		r.With(adminMW).Post("/api/processes/{pid}/kill", ph.Kill)
		r.With(adminMW).Post("/api/processes/{pid}/renice", ph.Renice)

		// User management (admin only)
		uh := &handlers.UsersHandler{DB: opts.DB}
		r.With(adminMW).Get("/api/users", uh.ListUsers)
		r.With(adminMW).Get("/api/groups", uh.ListGroups)
		r.With(adminMW).Post("/api/users", uh.CreateUser)
		r.With(adminMW).Delete("/api/users/{username}", uh.DeleteUser)
		r.With(adminMW).Post("/api/users/{username}/password", uh.SetPassword)

		// Firewall (UFW)
		if opts.UFWEnabled {
			fh := &handlers.FirewallHandler{DB: opts.DB}
			r.Get("/api/firewall/status", fh.Status)
			r.With(adminMW).Post("/api/firewall/rules", fh.AddRule)
			r.With(adminMW).Delete("/api/firewall/rules/{num}", fh.DeleteRule)
			r.With(adminMW).Post("/api/firewall/enable", fh.Enable)
			r.With(adminMW).Post("/api/firewall/disable", fh.Disable)
		}

		// Account settings (own user)
		acc := &handlers.AccountHandler{DB: opts.DB, BcryptCost: opts.BcryptCost}
		r.Post("/api/account/password", acc.ChangeSelfPassword)
		r.Get("/api/account/sessions", acc.ListSelfSessions)
		r.Delete("/api/account/sessions/{id}", acc.RevokeSelfSession)

		// Admin session management
		adm := &handlers.AdminSessionsHandler{DB: opts.DB}
		r.With(adminMW).Get("/api/admin/sessions", adm.List)
		r.With(adminMW).Delete("/api/admin/sessions/{id}", adm.Revoke)

		// TOTP/2FA enrollment (self)
		th := &handlers.TOTPHandler{DB: opts.DB, BcryptCost: opts.BcryptCost, Issuer: "linux-admin"}
		r.Get("/api/account/totp/status", th.Status)
		r.Post("/api/account/totp/enroll", th.Enroll)
		r.Post("/api/account/totp/confirm", th.Confirm)
		r.Delete("/api/account/totp", th.Revoke)

		// Audit log (admin only)
		r.With(adminMW).Get("/api/admin/audit", (&handlers.AuditLogHandler{DB: opts.DB}).List)

		// File browser (read-only)
		if len(opts.FilesRoots) > 0 {
			fh := &handlers.FilesHandler{AllowedRoots: opts.FilesRoots}
			r.Get("/api/files/list", fh.List)
			r.Get("/api/files/read", fh.Read)
			r.Get("/api/files/roots", fh.GetRoots)
		}

		// Cron job editor
		if opts.CronEnabled {
			ch := &handlers.CronHandler{}
			r.Get("/api/cron", ch.List)
			r.With(adminMW).Post("/api/cron", ch.Add)
			r.With(adminMW).Delete("/api/cron/{index}", ch.Delete)
		}

		// Network interfaces
		r.Get("/api/network/interfaces", (&handlers.NetifHandler{}).List)

		// Hardware sensors
		if opts.SensorsEnabled {
			r.Get("/api/sensors", (&handlers.SensorsHandler{}).Read)
		}

		// AI assistant (Ollama)
		if opts.OllamaClient != nil {
			ai := &handlers.AIHandler{Client: opts.OllamaClient, Metrics: opts.Collector}
			r.Get("/ws/ai", ai.ServeHTTP)
			r.Get("/api/ai/models", ai.Models)
		}
	})

	// Serve embedded frontend (SPA fallback)
	if opts.FrontendHandler != nil {
		r.Handle("/*", spaHandler(opts.FrontendHandler))
	}

	if opts.PrometheusOn {
		r.Group(func(r chi.Router) {
			r.Use(authMW)
			r.Get("/metrics", prometheusHandler(opts.Collector))
		})
	}

	return r
}

func spaHandler(static http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		static.ServeHTTP(w, r)
	}
}

func prometheusHandler(c *metrics.Collector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := c.Latest()
		if snap == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "# HELP linux_admin_cpu_percent CPU usage percent\n")
		fmt.Fprintf(w, "linux_admin_cpu_percent %.2f\n", snap.CPUPercent)
		fmt.Fprintf(w, "linux_admin_mem_used_kb %d\n", snap.MemUsed)
		fmt.Fprintf(w, "linux_admin_mem_total_kb %d\n", snap.MemTotal)
		fmt.Fprintf(w, "linux_admin_load1 %.2f\n", snap.LoadAvg1)
		fmt.Fprintf(w, "linux_admin_uptime_seconds %.0f\n", snap.Uptime)
	}
}
