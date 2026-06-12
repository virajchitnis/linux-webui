package auth

import (
	"database/sql"
	"time"
)

func RecordFailedLogin(db *sql.DB, ip string, maxAttempts, lockoutMinutes int) error {
	_, err := db.Exec(`
		INSERT INTO brute_force (ip, attempts, locked_until)
		VALUES (?, 1, NULL)
		ON CONFLICT(ip) DO UPDATE SET
			attempts = attempts + 1,
			locked_until = CASE
				WHEN attempts + 1 >= ? THEN datetime('now', '+' || ? || ' minutes')
				ELSE locked_until
			END`,
		ip, maxAttempts, lockoutMinutes)
	return err
}

func IsLockedOut(db *sql.DB, ip string) bool {
	var lockedUntil sql.NullString
	err := db.QueryRow(`SELECT locked_until FROM brute_force WHERE ip = ?`, ip).Scan(&lockedUntil)
	if err != nil || !lockedUntil.Valid {
		return false
	}
	return parseAndCheck(lockedUntil.String)
}

func ResetFailedLogins(db *sql.DB, ip string) error {
	_, err := db.Exec(`DELETE FROM brute_force WHERE ip = ?`, ip)
	return err
}

func RecordFailedLoginUsername(db *sql.DB, username string, maxAttempts, lockoutMinutes int) error {
	_, err := db.Exec(`
		INSERT INTO brute_force_user (username, attempts, locked_until)
		VALUES (?, 1, NULL)
		ON CONFLICT(username) DO UPDATE SET
			attempts = attempts + 1,
			locked_until = CASE
				WHEN attempts + 1 >= ? THEN datetime('now', '+' || ? || ' minutes')
				ELSE locked_until
			END`,
		username, maxAttempts, lockoutMinutes)
	return err
}

func IsLockedOutUsername(db *sql.DB, username string) bool {
	var lockedUntil sql.NullString
	err := db.QueryRow(`SELECT locked_until FROM brute_force_user WHERE username = ?`, username).Scan(&lockedUntil)
	if err != nil || !lockedUntil.Valid {
		return false
	}
	return parseAndCheck(lockedUntil.String)
}

// parseAndCheck parses a datetime string (SQLite may return RFC 3339 or space-separated)
// and returns true if the time is in the future.
func parseAndCheck(s string) bool {
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil {
			return time.Now().UTC().Before(t.UTC())
		}
	}
	return false
}

func ResetFailedLoginsUsername(db *sql.DB, username string) error {
	_, err := db.Exec(`DELETE FROM brute_force_user WHERE username = ?`, username)
	return err
}
