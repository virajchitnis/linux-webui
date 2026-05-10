package auth

import (
	"database/sql"
)

func LogAction(db *sql.DB, userID int64, username, action, target, ip string) {
	_, _ = db.Exec(`INSERT INTO audit_log (user_id, username, action, target, ip) VALUES (?, ?, ?, ?, ?)`,
		userID, username, action, target, ip)
}

func PruneAuditLog(db *sql.DB, retentionDays int) error {
	if retentionDays <= 0 {
		return nil
	}
	_, err := db.Exec(`DELETE FROM audit_log WHERE created_at < datetime('now', '-' || ? || ' days')`,
		retentionDays)
	return err
}

func GetAuditLog(db *sql.DB, limit int) ([]map[string]any, error) {
	rows, err := db.Query(`
		SELECT id, COALESCE(username,''), action, COALESCE(target,''), COALESCE(ip,''), created_at
		FROM audit_log ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id int64
		var username, action, target, ip, created string
		if err := rows.Scan(&id, &username, &action, &target, &ip, &created); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"id":         id,
			"username":   username,
			"action":     action,
			"target":     target,
			"ip":         ip,
			"created_at": created,
		})
	}
	return out, rows.Err()
}
