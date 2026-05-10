package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/virajchitnis/linux-webui/internal/api/middleware"
	"github.com/virajchitnis/linux-webui/internal/auth"
)

type AuthHandler struct {
	DB             *sql.DB
	TimeoutMinutes int
	BcryptCost     int
	SecureCookie   bool
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	ip := middleware.ClientIP(r)
	if auth.IsLockedOut(h.DB, ip) {
		http.Error(w, "too many failed attempts", http.StatusTooManyRequests)
		return
	}

	user, err := auth.GetUserByUsername(h.DB, body.Username)
	if err != nil || !auth.VerifyPassword(user.PasswordHash, body.Password) {
		_ = auth.RecordFailedLogin(h.DB, ip, 5, 15)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	_ = auth.ResetFailedLogins(h.DB, ip)
	sessionID, err := auth.CreateSession(h.DB, user.ID, ip, r.UserAgent())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	csrfToken, _ := auth.GenerateCSRFToken()
	auth.SetSessionCookie(w, sessionID, h.SecureCookie)
	auth.SetCSRFCookie(w, csrfToken, h.SecureCookie)
	auth.LogAction(h.DB, user.ID, user.Username, "login", "", ip)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"username": user.Username,
		"role":     user.Role,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	id := auth.SessionIDFromRequest(r)
	if id != "" {
		session := middleware.SessionFromContext(r.Context())
		if session != nil {
			auth.LogAction(h.DB, session.UserID, session.Username, "logout", "", middleware.ClientIP(r))
		}
		_ = auth.DeleteSession(h.DB, id)
	}
	auth.ClearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	session := middleware.SessionFromContext(r.Context())
	if session == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"username": session.Username,
		"role":     session.Role,
	})
}

type SetupHandler struct {
	DB         *sql.DB
	BcryptCost int
}

func (h *SetupHandler) Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"complete": auth.IsSetupComplete(h.DB),
	})
}

func (h *SetupHandler) Complete(w http.ResponseWriter, r *http.Request) {
	if auth.IsSetupComplete(h.DB) {
		http.Error(w, "already configured", http.StatusConflict)
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if len(body.Username) < 1 || len(body.Password) < 8 {
		http.Error(w, "username required; password must be at least 8 characters", http.StatusBadRequest)
		return
	}
	if err := auth.CreateUser(h.DB, body.Username, body.Password, "admin", h.BcryptCost); err != nil {
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}
	if err := auth.SetSetupComplete(h.DB); err != nil {
		http.Error(w, "failed to mark setup complete", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
