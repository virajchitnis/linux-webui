package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/virajchitnis/linux-webui/internal/api/middleware"
	"github.com/virajchitnis/linux-webui/internal/auth"
	"github.com/virajchitnis/linux-webui/internal/firewall"
)

type FirewallHandler struct {
	DB *sql.DB
}

func (h *FirewallHandler) Status(w http.ResponseWriter, r *http.Request) {
	s, err := firewall.GetStatus()
	if err != nil {
		http.Error(w, "ufw: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

func (h *FirewallHandler) AddRule(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Action string `json:"action"`
		Port   string `json:"port"`
		Proto  string `json:"proto"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if body.Proto == "" {
		body.Proto = "tcp"
	}
	if err := firewall.AddRule(body.Action, body.Port, body.Proto); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sess := middleware.SessionFromContext(r.Context())
	if sess != nil {
		auth.LogAction(h.DB, sess.UserID, sess.Username, "fw_add", body.Action+" "+body.Port+"/"+body.Proto, middleware.ClientIP(r))
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *FirewallHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	num, err := strconv.Atoi(chi.URLParam(r, "num"))
	if err != nil || num < 1 {
		http.Error(w, "invalid rule number", http.StatusBadRequest)
		return
	}
	if err := firewall.DeleteRule(num); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sess := middleware.SessionFromContext(r.Context())
	if sess != nil {
		auth.LogAction(h.DB, sess.UserID, sess.Username, "fw_delete", strconv.Itoa(num), middleware.ClientIP(r))
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *FirewallHandler) Enable(w http.ResponseWriter, r *http.Request) {
	if err := firewall.Enable(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sess := middleware.SessionFromContext(r.Context())
	if sess != nil {
		auth.LogAction(h.DB, sess.UserID, sess.Username, "fw_enable", "", middleware.ClientIP(r))
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *FirewallHandler) Disable(w http.ResponseWriter, r *http.Request) {
	if err := firewall.Disable(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sess := middleware.SessionFromContext(r.Context())
	if sess != nil {
		auth.LogAction(h.DB, sess.UserID, sess.Username, "fw_disable", "", middleware.ClientIP(r))
	}
	w.WriteHeader(http.StatusNoContent)
}
