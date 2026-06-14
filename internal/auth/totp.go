package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

const recoveryCodeCount = 8

var ErrInvalidTOTP = errors.New("invalid TOTP code")

// GenerateTOTPSecret creates a new TOTP key for the given username/issuer.
// Returns the secret (base32) and the otpauth URL for QR code generation.
func GenerateTOTPSecret(username, issuer string) (secret, otpauthURL string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: username,
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
		Period:      30,
	})
	if err != nil {
		return "", "", fmt.Errorf("totp generate: %w", err)
	}
	return key.Secret(), key.URL(), nil
}

// VerifyTOTP checks a 6-digit code against the stored secret.
func VerifyTOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}

// SaveTOTPSecret stores the confirmed TOTP secret for the user and
// generates + stores recovery codes, returning the plaintext codes.
func SaveTOTPSecret(db *sql.DB, userID int64, secret string, cost int) ([]string, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE users SET totp_secret = ? WHERE id = ?`, secret, userID); err != nil {
		return nil, err
	}

	// Remove any existing recovery codes.
	if _, err := tx.Exec(`DELETE FROM totp_recovery WHERE user_id = ?`, userID); err != nil {
		return nil, err
	}

	codes := make([]string, recoveryCodeCount)
	for i := range codes {
		raw, err := generateRecoveryCode()
		if err != nil {
			return nil, err
		}
		codes[i] = raw
		hash, err := bcrypt.GenerateFromPassword([]byte(raw), cost)
		if err != nil {
			return nil, err
		}
		if _, err := tx.Exec(`INSERT INTO totp_recovery (user_id, hash) VALUES (?, ?)`, userID, string(hash)); err != nil {
			return nil, err
		}
	}

	return codes, tx.Commit()
}

// RevokeTOTP removes the TOTP secret and all recovery codes.
func RevokeTOTP(db *sql.DB, userID int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE users SET totp_secret = NULL WHERE id = ?`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM totp_recovery WHERE user_id = ?`, userID); err != nil {
		return err
	}
	return tx.Commit()
}

// UseRecoveryCode verifies and consumes a single-use recovery code.
func UseRecoveryCode(db *sql.DB, userID int64, code string) error {
	rows, err := db.Query(`SELECT id, hash FROM totp_recovery WHERE user_id = ?`, userID)
	if err != nil {
		return err
	}
	defer rows.Close()
	var matchID int64
	for rows.Next() {
		var id int64
		var hash string
		if err := rows.Scan(&id, &hash); err != nil {
			return err
		}
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(code)) == nil {
			matchID = id
			break
		}
	}
	rows.Close()
	if matchID == 0 {
		return ErrInvalidTOTP
	}
	_, err = db.Exec(`DELETE FROM totp_recovery WHERE id = ?`, matchID)
	return err
}

func generateRecoveryCode() (string, error) {
	b := make([]byte, 10)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	raw := strings.ToUpper(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b))
	// Format as XXXXX-XXXXX
	return raw[:8] + "-" + raw[8:], nil
}
