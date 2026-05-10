package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/virajchitnis/linux-webui/internal/api/middleware"
	"github.com/virajchitnis/linux-webui/internal/auth"
	dbuspkg "github.com/virajchitnis/linux-webui/internal/dbus"
)

type ServicesHandler struct {
	DBus *dbuspkg.Client
	DB   *sql.DB
}

func (h *ServicesHandler) List(w http.ResponseWriter, r *http.Request) {
	units, err := h.DBus.ListUnits()
	if err != nil {
		http.Error(w, "failed to list units: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Filter to only service units by default
	filter := r.URL.Query().Get("type")
	if filter == "" {
		filter = "service"
	}
	var filtered []dbuspkg.UnitInfo
	for _, u := range units {
		if filter == "all" || strings.HasSuffix(u.Name, "."+filter) {
			filtered = append(filtered, u)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filtered)
}

func (h *ServicesHandler) action(w http.ResponseWriter, r *http.Request, fn func(string) error) {
	session := middleware.SessionFromContext(r.Context())
	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "unit name required", http.StatusBadRequest)
		return
	}
	if err := fn(name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if session != nil {
		auth.LogAction(h.DB, session.UserID, session.Username,
			"service_action", name, middleware.ClientIP(r))
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ServicesHandler) Start(w http.ResponseWriter, r *http.Request) {
	h.action(w, r, h.DBus.StartUnit)
}

func (h *ServicesHandler) Stop(w http.ResponseWriter, r *http.Request) {
	h.action(w, r, h.DBus.StopUnit)
}

func (h *ServicesHandler) Restart(w http.ResponseWriter, r *http.Request) {
	h.action(w, r, h.DBus.RestartUnit)
}

func (h *ServicesHandler) Enable(w http.ResponseWriter, r *http.Request) {
	h.action(w, r, h.DBus.EnableUnit)
}

func (h *ServicesHandler) Disable(w http.ResponseWriter, r *http.Request) {
	h.action(w, r, h.DBus.DisableUnit)
}
