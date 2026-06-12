package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/virajchitnis/linux-webui/internal/process"
)

type ProcessHandler struct{}

func (h *ProcessHandler) List(w http.ResponseWriter, r *http.Request) {
	procs, err := process.List()
	if err != nil {
		log.Printf("process list: %v", err)
		http.Error(w, "failed to list processes", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(procs)
}

func (h *ProcessHandler) Kill(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(chi.URLParam(r, "pid"))
	if err != nil || pid <= 0 {
		http.Error(w, "invalid pid", http.StatusBadRequest)
		return
	}
	var body struct {
		Signal string `json:"signal"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	sig := syscall.SIGTERM
	switch body.Signal {
	case "SIGKILL":
		sig = syscall.SIGKILL
	case "SIGHUP":
		sig = syscall.SIGHUP
	case "SIGSTOP":
		sig = syscall.SIGSTOP
	case "SIGCONT":
		sig = syscall.SIGCONT
	}

	if err := process.Kill(pid, sig); err != nil {
		log.Printf("kill pid %d: %v", pid, err)
		http.Error(w, "failed to send signal", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProcessHandler) Renice(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(chi.URLParam(r, "pid"))
	if err != nil || pid <= 0 {
		http.Error(w, "invalid pid", http.StatusBadRequest)
		return
	}
	var body struct {
		Priority int `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := process.Renice(pid, body.Priority); err != nil {
		log.Printf("renice pid %d: %v", pid, err)
		http.Error(w, "failed to renice process", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
