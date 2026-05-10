package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/virajchitnis/linux-webui/internal/api/middleware"
	"github.com/virajchitnis/linux-webui/internal/auth"
)

// TOTPHandler manages TOTP enrollment for the authenticated user.
type TOTPHandler struct {
	DB         *sql.DB
	BcryptCost int
	Issuer     string
}

func (h *TOTPHandler) Status(w http.ResponseWriter, r *http.Request) {
	session := middleware.SessionFromContext(r.Context())
	user, err := auth.GetUserByID(h.DB, session.UserID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"enabled": user.TOTPSecret != ""})
}

// Enroll generates a new TOTP secret and returns it + the otpauth URL for QR rendering.
// The secret is NOT saved until Confirm is called.
func (h *TOTPHandler) Enroll(w http.ResponseWriter, r *http.Request) {
	session := middleware.SessionFromContext(r.Context())
	issuer := h.Issuer
	if issuer == "" {
		issuer = "linux-webui"
	}
	secret, otpauthURL, err := auth.GenerateTOTPSecret(session.Username, issuer)
	if err != nil {
		http.Error(w, "failed to generate TOTP secret", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"secret":      secret,
		"otpauth_url": otpauthURL,
	})
}

// Confirm verifies the user-supplied code against the pending secret and saves it.
func (h *TOTPHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	session := middleware.SessionFromContext(r.Context())
	var body struct {
		Secret string `json:"secret"`
		Code   string `json:"code"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if body.Secret == "" || body.Code == "" {
		http.Error(w, "secret and code required", http.StatusBadRequest)
		return
	}
	if !auth.VerifyTOTP(body.Secret, body.Code) {
		http.Error(w, "invalid TOTP code", http.StatusBadRequest)
		return
	}
	recoveryCodes, err := auth.SaveTOTPSecret(h.DB, session.UserID, body.Secret, h.BcryptCost)
	if err != nil {
		http.Error(w, "failed to save TOTP secret", http.StatusInternalServerError)
		return
	}
	auth.LogAction(h.DB, session.UserID, session.Username, "totp_enabled", "", middleware.ClientIP(r))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"recovery_codes": recoveryCodes})
}

// Revoke disables TOTP for the authenticated user.
func (h *TOTPHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	session := middleware.SessionFromContext(r.Context())
	if err := auth.RevokeTOTP(h.DB, session.UserID); err != nil {
		http.Error(w, "failed to revoke TOTP", http.StatusInternalServerError)
		return
	}
	auth.LogAction(h.DB, session.UserID, session.Username, "totp_disabled", "", middleware.ClientIP(r))
	w.WriteHeader(http.StatusNoContent)
}
