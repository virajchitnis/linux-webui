package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const csrfCookieName = "linux_admin_csrf"
const csrfHeaderName = "X-CSRF-Token"

func GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func SetCSRFCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false, // JavaScript must read this to send in header
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400,
	})
}

func ValidateCSRF(r *http.Request) bool {
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil {
		return false
	}
	header := r.Header.Get(csrfHeaderName)
	return cookie.Value != "" && cookie.Value == header
}
