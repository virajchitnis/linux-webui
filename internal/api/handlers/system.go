package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/virajchitnis/linux-webui/internal/api/middleware"
	"github.com/virajchitnis/linux-webui/internal/auth"
	"github.com/virajchitnis/linux-webui/internal/metrics"
)

type SystemHandler struct {
	Collector *metrics.Collector
	Version   string
	DB        *sql.DB
}

func (h *SystemHandler) Info(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	kernelVersion := readKernelVersion()
	var snap *metrics.Snapshot
	if h.Collector != nil {
		snap = h.Collector.Latest()
	}

	info := map[string]any{
		"hostname":        hostname,
		"kernel":          kernelVersion,
		"os":              runtime.GOOS,
		"arch":            runtime.GOARCH,
		"app_version":     h.Version,
	}
	if snap != nil {
		info["uptime_seconds"] = snap.Uptime
		info["load1"] = snap.LoadAvg1
		info["load5"] = snap.LoadAvg5
		info["load15"] = snap.LoadAvg15
		info["mem_total_kb"] = snap.MemTotal
		info["mem_used_kb"] = snap.MemUsed
		info["swap_total_kb"] = snap.SwapTotal
		info["swap_used_kb"] = snap.SwapUsed
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func (h *SystemHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	if h.Collector == nil {
		http.Error(w, "metrics unavailable", http.StatusServiceUnavailable)
		return
	}
	snap := h.Collector.Latest()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snap)
}

func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Reboot triggers an immediate system reboot via sudo /sbin/reboot.
func (h *SystemHandler) Reboot(w http.ResponseWriter, r *http.Request) {
	session := middleware.SessionFromContext(r.Context())
	if session != nil {
		auth.LogAction(h.DB, session.UserID, session.Username, "reboot", "", middleware.ClientIP(r))
	}
	cmd := exec.Command("/usr/bin/sudo", "/sbin/reboot")
	if err := cmd.Run(); err != nil {
		log.Printf("reboot: %v", err)
		http.Error(w, "reboot failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func readKernelVersion() string {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return "unknown"
	}
	line := strings.TrimSpace(string(data))
	parts := strings.Fields(line)
	if len(parts) >= 3 {
		return parts[2]
	}
	return line
}
