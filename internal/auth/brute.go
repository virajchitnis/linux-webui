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
	t, err := time.Parse("2006-01-02 15:04:05", lockedUntil.String)
	if err != nil {
		return false
	}
	return time.Now().Before(t)
}

func ResetFailedLogins(db *sql.DB, ip string) error {
	_, err := db.Exec(`DELETE FROM brute_force WHERE ip = ?`, ip)
	return err
}
