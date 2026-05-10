package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/virajchitnis/linux-webui/internal/api"
	"github.com/virajchitnis/linux-webui/internal/auth"
	"github.com/virajchitnis/linux-webui/internal/config"
	dbuspkg "github.com/virajchitnis/linux-webui/internal/dbus"
	"github.com/virajchitnis/linux-webui/internal/distro"
	"github.com/virajchitnis/linux-webui/internal/metrics"
	"github.com/virajchitnis/linux-webui/internal/terminal"
)

// version is set at build time via -ldflags.
var version = "dev"

func main() {
	var (
		configPath = flag.String("config", "/etc/linux-admin/config.toml", "path to config file")
		install    = flag.Bool("install", false, "install linux-admin (create user, polkit, sudoers, systemd unit)")
		uninstall  = flag.Bool("uninstall", false, "uninstall linux-admin")
		showVer    = flag.Bool("version", false, "print version and exit")
	)
	flag.Parse()

	if *showVer {
		fmt.Printf("linux-admin %s\n", version)
		return
	}
	if *install {
		runInstall()
		return
	}
	if *uninstall {
		runUninstall()
		return
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	dbPath := "/etc/linux-admin/linux-admin.db"
	if _, err := os.Stat("/etc/linux-admin"); os.IsNotExist(err) {
		dbPath = "linux-admin.db"
	}
	db, err := auth.OpenDB(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	d := distro.Detect()
	log.Printf("distro: %s (%s family)", d.ID, d.FamilyString())

	// Connect to systemd via D-Bus (non-fatal if unavailable)
	var dbusClient *dbuspkg.Client
	if dc, err := dbuspkg.New(); err != nil {
		log.Printf("warning: D-Bus unavailable (%v); service manager disabled", err)
	} else {
		dbusClient = dc
		defer dbusClient.Close()
	}

	collector := metrics.NewCollector()
	collector.Start()

	// Terminal (PTY) manager — only if terminal is enabled in config and bash is available.
	var termMgr *terminal.Manager
	if cfg.Features.TerminalEnabled {
		if _, err := os.Stat("/bin/bash"); err == nil {
			idleTimeout := time.Duration(cfg.Auth.TerminalReconnectSecs) * time.Second
			if idleTimeout <= 0 {
				idleTimeout = 30 * time.Second
			}
			maxSess := cfg.Features.MaxTerminalSessions
			if maxSess <= 0 {
				maxSess = 3
			}
			termMgr = terminal.NewManager(maxSess, idleTimeout)
		}
	}

	// Feature availability detection.
	aptEnabled := d.Family == distro.FamilyDebian
	journalEnabled := false
	if _, e := os.Stat("/usr/bin/journalctl"); e == nil {
		journalEnabled = true
	}
	ufwEnabled := false
	if _, e := os.Stat("/usr/sbin/ufw"); e == nil {
		ufwEnabled = true
	}

	router := api.NewRouter(api.RouterOptions{
		DB:              db,
		Distro:          d,
		Collector:       collector,
		DBus:            dbusClient,
		TerminalManager: termMgr,
		Version:         version,
		SecureCookie:    true,
		AuthTimeout:     cfg.Auth.SessionTimeoutMinutes,
		BcryptCost:      cfg.Auth.BcryptCost,
		PrometheusOn:    cfg.Monitoring.Prometheus,
		AptEnabled:      aptEnabled,
		JournalEnabled:  journalEnabled,
		UFWEnabled:      ufwEnabled,
		FrontendHandler: frontendHandler(),
	})

	tlsCfg, err := buildTLS(cfg)
	if err != nil {
		log.Fatalf("tls: %v", err)
	}

	srv := &http.Server{
		Addr:         cfg.Server.Listen,
		Handler:      router,
		TLSConfig:    tlsCfg,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// HTTP → HTTPS redirect server
	httpRedirect := &http.Server{
		Addr:         cfg.Server.HTTPListen,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			target := "https://" + r.Host + r.RequestURI
			http.Redirect(w, r, target, http.StatusMovedPermanently)
		}),
	}

	go func() {
		log.Printf("HTTP redirect on %s", cfg.Server.HTTPListen)
		if err := httpRedirect.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("http redirect: %v", err)
		}
	}()

	log.Printf("linux-admin %s listening on %s", version, cfg.Server.Listen)
	go func() {
		if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Fatalf("https: %v", err)
		}
	}()

	// Schedule audit log pruning daily.
	go func() {
		for range time.Tick(24 * time.Hour) {
			if err := auth.PruneAuditLog(db, cfg.Audit.RetentionDays); err != nil {
				log.Printf("audit prune: %v", err)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Println("shutting down…")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = httpRedirect.Shutdown(ctx)
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}

	// Flush SQLite WAL
	_, _ = db.Exec(`PRAGMA wal_checkpoint(TRUNCATE)`)
	log.Println("stopped")
}

func buildTLS(cfg *config.Config) (*tls.Config, error) {
	cert, key := cfg.Server.TLSCert, cfg.Server.TLSKey

	// If cert/key don't exist, generate a self-signed certificate.
	if _, err := os.Stat(cert); os.IsNotExist(err) {
		if err := generateSelfSigned(cert, key); err != nil {
			return nil, fmt.Errorf("generate self-signed: %w", err)
		}
	}

	tlsCert, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
	}, nil
}

func generateSelfSigned(certPath, keyPath string) error {
	if err := os.MkdirAll(filepath.Dir(certPath), 0700); err != nil {
		return err
	}
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	hostname, _ := os.Hostname()
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"linux-admin"},
			CommonName:   hostname,
		},
		DNSNames:    []string{hostname, "localhost"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:   time.Now().Add(-time.Hour),
		NotAfter:    time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		return err
	}
	cf, err := os.OpenFile(certPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer cf.Close()
	if err := pem.Encode(cf, &pem.Block{Type: "CERTIFICATE", Bytes: der}); err != nil {
		return err
	}
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return err
	}
	kf, err := os.OpenFile(keyPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer kf.Close()
	return pem.Encode(kf, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})
}

func runInstall() {
	fmt.Println("Installation must be run via install.sh")
	fmt.Println("  sudo bash install.sh")
}

func runUninstall() {
	fmt.Println("Uninstallation must be run via install.sh --uninstall")
}
