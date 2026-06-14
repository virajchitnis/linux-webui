package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
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
		log.Printf("list users: %v", err)
		http.Error(w, "failed to list users", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *UsersHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	list, err := users.ListGroups()
	if err != nil {
		log.Printf("list groups: %v", err)
		http.Error(w, "failed to list groups", http.StatusInternalServerError)
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
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := users.CreateUser(body.Username); err != nil {
		log.Printf("create user %q: %v", body.Username, err)
		http.Error(w, "failed to create user", http.StatusBadRequest)
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
		log.Printf("delete user %q: %v", username, err)
		http.Error(w, "failed to delete user", http.StatusBadRequest)
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
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := users.SetPassword(username, body.Password); err != nil {
		log.Printf("set password for %q: %v", username, err)
		http.Error(w, "failed to set password", http.StatusBadRequest)
		return
	}
	sess := middleware.SessionFromContext(r.Context())
	if sess != nil {
		auth.LogAction(h.DB, sess.UserID, sess.Username, "set_password", username, middleware.ClientIP(r))
	}
	w.WriteHeader(http.StatusNoContent)
}
