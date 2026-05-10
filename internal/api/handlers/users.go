package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/virajchitnis/linux-webui/internal/api/middleware"
	"github.com/virajchitnis/linux-webui/internal/auth"
	"github.com/virajchitnis/linux-webui/internal/users"
)

type UsersHandler struct {
	DB *sql.DB
}

func (h *UsersHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	list, err := users.ListUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *UsersHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	list, err := users.ListGroups()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *UsersHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := users.CreateUser(body.Username); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if body.Password != "" {
		_ = users.SetPassword(body.Username, body.Password)
	}
	sess := middleware.SessionFromContext(r.Context())
	if sess != nil {
		auth.LogAction(h.DB, sess.UserID, sess.Username, "create_user", body.Username, middleware.ClientIP(r))
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *UsersHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if err := users.DeleteUser(username); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sess := middleware.SessionFromContext(r.Context())
	if sess != nil {
		auth.LogAction(h.DB, sess.UserID, sess.Username, "delete_user", username, middleware.ClientIP(r))
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UsersHandler) SetPassword(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := users.SetPassword(username, body.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sess := middleware.SessionFromContext(r.Context())
	if sess != nil {
		auth.LogAction(h.DB, sess.UserID, sess.Username, "set_password", username, middleware.ClientIP(r))
	}
	w.WriteHeader(http.StatusNoContent)
}
