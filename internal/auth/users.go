package auth

import (
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var ErrUserNotFound = errors.New("user not found")
var ErrInvalidCredentials = errors.New("invalid credentials")

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Role         string
	TOTPSecret   string
}

func CreateUser(db *sql.DB, username, password, role string, cost int) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	_, err = db.Exec(`INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)`,
		username, string(hash), role)
	return err
}

func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	var u User
	var totpSecret sql.NullString
	err := db.QueryRow(`SELECT id, username, password_hash, role, totp_secret FROM users WHERE username = ?`,
		username).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &totpSecret)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	if totpSecret.Valid {
		u.TOTPSecret = totpSecret.String
	}
	return &u, nil
}

func GetUserByID(db *sql.DB, id int64) (*User, error) {
	var u User
	var totpSecret sql.NullString
	err := db.QueryRow(`SELECT id, username, password_hash, role, totp_secret FROM users WHERE id = ?`,
		id).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &totpSecret)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	if totpSecret.Valid {
		u.TOTPSecret = totpSecret.String
	}
	return &u, nil
}

func VerifyPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func ChangePassword(db *sql.DB, userID int64, newPassword string, cost int) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), cost)
	if err != nil {
		return err
	}
	_, err = db.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, string(hash), userID)
	return err
}

func ListUsers(db *sql.DB) ([]map[string]any, error) {
	rows, err := db.Query(`SELECT id, username, role, created_at FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id int64
		var username, role, created string
		if err := rows.Scan(&id, &username, &role, &created); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"id":         id,
			"username":   username,
			"role":       role,
			"created_at": created,
		})
	}
	return out, rows.Err()
}

func IsSetupComplete(db *sql.DB) bool {
	var v string
	err := db.QueryRow(`SELECT value FROM meta WHERE key = 'setup_complete'`).Scan(&v)
	return err == nil && v == "1"
}

func SetSetupComplete(db *sql.DB) error {
	_, err := db.Exec(`INSERT OR REPLACE INTO meta (key, value) VALUES ('setup_complete', '1')`)
	return err
}
