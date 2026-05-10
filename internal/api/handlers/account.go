package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/virajchitnis/linux-webui/internal/api/middleware"
	"github.com/virajchitnis/linux-webui/internal/auth"
)

// AccountHandler handles the current user's own account operations.
type AccountHandler struct {
	DB         *sql.DB
	BcryptCost int
}

// ChangeSelfPassword changes the currently-authenticated user's password.
func (h *AccountHandler) ChangeSelfPassword(w http.ResponseWriter, r *http.Request) {
	sess := middleware.SessionFromContext(r.Context())
	if sess == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if len(body.NewPassword) < 8 {
		http.Error(w, "new password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	u, err := auth.GetUserByID(h.DB, sess.UserID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	if !auth.VerifyPassword(u.PasswordHash, body.CurrentPassword) {
		http.Error(w, "current password incorrect", http.StatusForbidden)
		return
	}
	if err := auth.ChangePassword(h.DB, sess.UserID, body.NewPassword, h.BcryptCost); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	auth.LogAction(h.DB, sess.UserID, sess.Username, "change_password", "", middleware.ClientIP(r))
	w.WriteHeader(http.StatusNoContent)
}

// ListSelfSessions returns the current user's active sessions.
func (h *AccountHandler) ListSelfSessions(w http.ResponseWriter, r *http.Request) {
	sess := middleware.SessionFromContext(r.Context())
	if sess == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	sessions, err := auth.ListSessions(h.DB, sess.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

// RevokeSelfSession deletes one of the current user's own sessions.
func (h *AccountHandler) RevokeSelfSession(w http.ResponseWriter, r *http.Request) {
	sess := middleware.SessionFromContext(r.Context())
	if sess == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	// Verify session belongs to this user before deleting.
	sessions, err := auth.ListSessions(h.DB, sess.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	found := false
	for _, s := range sessions {
		if sid, _ := s["id"].(string); sid == id {
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}
	if err := auth.DeleteSession(h.DB, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AdminListSessions returns all sessions across all users (admin-only).
type AdminSessionsHandler struct {
	DB *sql.DB
}

func (h *AdminSessionsHandler) List(w http.ResponseWriter, r *http.Request) {
	sessions, err := auth.ListAllSessions(h.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

func (h *AdminSessionsHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	sess := middleware.SessionFromContext(r.Context())
	id := chi.URLParam(r, "id")
	if err := auth.DeleteSession(h.DB, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if sess != nil {
		auth.LogAction(h.DB, sess.UserID, sess.Username, "revoke_session", id, middleware.ClientIP(r))
	}
	w.WriteHeader(http.StatusNoContent)
}
