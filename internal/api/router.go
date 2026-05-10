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
	"github.com/virajchitnis/linux-webui/internal/metrics"
)

type RouterOptions struct {
	DB              *sql.DB
	Distro          *distro.Distro
	Collector       *metrics.Collector
	DBus            *dbuspkg.Client // may be nil if D-Bus unavailable
	Version         string
	SecureCookie    bool
	AuthTimeout     int
	BcryptCost      int
	PrometheusOn    bool
	FrontendHandler http.Handler // nil in dev mode
}

func NewRouter(opts RouterOptions) http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)

	caps := handlers.InitCapabilities(opts.Distro)
	rl := middleware.NewRateLimiter(20, 100)
	r.Use(rl.Limit)
	r.Use(middleware.SecurityHeaders)

	// Public endpoints
	r.Get("/health", handlers.Health)
	r.Get("/api/setup/status", (&handlers.SetupHandler{DB: opts.DB, BcryptCost: opts.BcryptCost}).Status)
	r.Post("/api/setup/complete", (&handlers.SetupHandler{DB: opts.DB, BcryptCost: opts.BcryptCost}).Complete)
	r.Post("/api/auth/login", (&handlers.AuthHandler{
		DB:             opts.DB,
		TimeoutMinutes: opts.AuthTimeout,
		BcryptCost:     opts.BcryptCost,
		SecureCookie:   opts.SecureCookie,
	}).Login)

	// Authenticated routes
	authMW := middleware.RequireAuth(middleware.Options{DB: opts.DB, TimeoutMinutes: opts.AuthTimeout})
	r.Group(func(r chi.Router) {
		r.Use(authMW)
		r.Use(middleware.RequireCSRF)

		r.Post("/api/auth/logout", (&handlers.AuthHandler{DB: opts.DB}).Logout)
		r.Get("/api/auth/me", (&handlers.AuthHandler{DB: opts.DB}).Me)

		sysHandler := &handlers.SystemHandler{Collector: opts.Collector, Version: opts.Version}
		r.Get("/api/system/info", sysHandler.Info)
		r.Get("/api/system/metrics", sysHandler.Metrics)
		r.Get("/api/capabilities", caps.Handler())

		// WebSocket: metrics stream
		r.Get("/ws/metrics", (&handlers.MetricsWSHandler{Collector: opts.Collector}).ServeHTTP)

		// Services (systemd via D-Bus)
		if opts.DBus != nil {
			svc := &handlers.ServicesHandler{DBus: opts.DBus, DB: opts.DB}
			r.Get("/api/services", svc.List)
			r.Post("/api/services/{name}/start", svc.Start)
			r.Post("/api/services/{name}/stop", svc.Stop)
			r.Post("/api/services/{name}/restart", svc.Restart)
			r.Post("/api/services/{name}/enable", svc.Enable)
			r.Post("/api/services/{name}/disable", svc.Disable)
		}
	})

	// Serve embedded frontend for all unmatched routes (SPA fallback)
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

// spaHandler serves static files and falls back to index.html for SPA routing.
func spaHandler(static http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Let the file server handle actual assets; for everything else serve index.html
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
		// Basic metrics in Prometheus text format
		fmt.Fprintf(w, "# HELP linux_admin_cpu_percent CPU usage percent\n")
		fmt.Fprintf(w, "linux_admin_cpu_percent %.2f\n", snap.CPUPercent)
		fmt.Fprintf(w, "linux_admin_mem_used_kb %d\n", snap.MemUsed)
		fmt.Fprintf(w, "linux_admin_mem_total_kb %d\n", snap.MemTotal)
		fmt.Fprintf(w, "linux_admin_load1 %.2f\n", snap.LoadAvg1)
		fmt.Fprintf(w, "linux_admin_uptime_seconds %.0f\n", snap.Uptime)
	}
}
