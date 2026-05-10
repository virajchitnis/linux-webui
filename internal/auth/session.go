package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const sessionCookieName = "linux_admin_session"

var ErrNoSession = errors.New("no session")
var ErrSessionExpired = errors.New("session expired")

type Session struct {
	ID        string
	UserID    int64
	Username  string
	Role      string
	IP        string
	UserAgent string
	LastSeen  time.Time
}

func generateID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func CreateSession(db *sql.DB, userID int64, ip, userAgent string) (string, error) {
	id, err := generateID()
	if err != nil {
		return "", err
	}
	_, err = db.Exec(`INSERT INTO sessions (id, user_id, ip, user_agent) VALUES (?, ?, ?, ?)`,
		id, userID, ip, userAgent)
	if err != nil {
		return "", fmt.Errorf("insert session: %w", err)
	}
	return id, nil
}

func GetSession(db *sql.DB, id string, timeoutMinutes int) (*Session, error) {
	var s Session
	var lastSeen time.Time
	err := db.QueryRow(`
		SELECT s.id, s.user_id, u.username, u.role, s.ip, s.user_agent, s.last_seen
		FROM sessions s JOIN users u ON u.id = s.user_id
		WHERE s.id = ?`, id).Scan(
		&s.ID, &s.UserID, &s.Username, &s.Role, &s.IP, &s.UserAgent, &lastSeen)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoSession
	}
	if err != nil {
		return nil, err
	}
	if time.Since(lastSeen) > time.Duration(timeoutMinutes)*time.Minute {
		_ = DeleteSession(db, id)
		return nil, ErrSessionExpired
	}
	s.LastSeen = lastSeen
	_, _ = db.Exec(`UPDATE sessions SET last_seen = CURRENT_TIMESTAMP WHERE id = ?`, id)
	return &s, nil
}

func DeleteSession(db *sql.DB, id string) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

func DeleteUserSessions(db *sql.DB, userID int64) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID)
	return err
}

func SetSessionCookie(w http.ResponseWriter, sessionID string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400 * 7,
	})
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

func SessionIDFromRequest(r *http.Request) string {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return ""
	}
	return c.Value
}

func ListSessions(db *sql.DB, userID int64) ([]map[string]any, error) {
	rows, err := db.Query(`
		SELECT s.id, s.ip, s.user_agent, s.created_at, s.last_seen
		FROM sessions s WHERE s.user_id = ?
		ORDER BY s.last_seen DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, ip, ua, created, lastSeen string
		if err := rows.Scan(&id, &ip, &ua, &created, &lastSeen); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"id":         id,
			"ip":         ip,
			"user_agent": ua,
			"created_at": created,
			"last_seen":  lastSeen,
		})
	}
	return out, rows.Err()
}

func ListAllSessions(db *sql.DB) ([]map[string]any, error) {
	rows, err := db.Query(`
		SELECT s.id, u.username, s.ip, s.user_agent, s.created_at, s.last_seen
		FROM sessions s JOIN users u ON u.id = s.user_id
		ORDER BY s.last_seen DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, username, ip, ua, created, lastSeen string
		if err := rows.Scan(&id, &username, &ip, &ua, &created, &lastSeen); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"id":         id,
			"username":   username,
			"ip":         ip,
			"user_agent": ua,
			"created_at": created,
			"last_seen":  lastSeen,
		})
	}
	return out, rows.Err()
}
