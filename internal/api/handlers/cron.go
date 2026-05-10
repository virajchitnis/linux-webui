package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/virajchitnis/linux-webui/internal/cron"
)

// CronHandler manages crontab entries.
type CronHandler struct{}

func (h *CronHandler) List(w http.ResponseWriter, r *http.Request) {
	entries, err := cron.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func (h *CronHandler) Add(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Schedule string `json:"schedule"`
		Command  string `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := cron.Add(body.Schedule, body.Command); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *CronHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idxStr := chi.URLParam(r, "index")
	idx, err := strconv.Atoi(idxStr)
	if err != nil || idx < 0 {
		http.Error(w, "invalid index", http.StatusBadRequest)
		return
	}
	if err := cron.Delete(idx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
