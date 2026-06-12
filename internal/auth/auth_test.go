package auth_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/virajchitnis/linux-webui/internal/auth"
)

// openTestDB creates a temp-file SQLite DB for tests and registers cleanup.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	f, err := os.CreateTemp("", "lwui-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })

	db, err := auth.OpenDB(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// ── CSRF ─────────────────────────────────────────────────────────────────────

func TestGenerateCSRFToken(t *testing.T) {
	tok, err := auth.GenerateCSRFToken()
	if err != nil {
		t.Fatal(err)
	}
	if len(tok) != 64 {
		t.Fatalf("expected 64-char hex token, got %d chars", len(tok))
	}

	tok2, _ := auth.GenerateCSRFToken()
	if tok == tok2 {
		t.Fatal("two tokens should not be equal")
	}
}

func TestValidateCSRF(t *testing.T) {
	tok, _ := auth.GenerateCSRFToken()

	buildReq := func(cookieVal, headerVal string) *http.Request {
		r := httptest.NewRequest(http.MethodPost, "/", nil)
		if cookieVal != "" {
			r.AddCookie(&http.Cookie{Name: "linux_webui_csrf", Value: cookieVal})
		}
		if headerVal != "" {
			r.Header.Set("X-CSRF-Token", headerVal)
		}
		return r
	}

	if !auth.ValidateCSRF(buildReq(tok, tok)) {
		t.Error("matching token/header should be valid")
	}
	if auth.ValidateCSRF(buildReq(tok, "wrong")) {
		t.Error("mismatched header should be invalid")
	}
	if auth.ValidateCSRF(buildReq("", tok)) {
		t.Error("missing cookie should be invalid")
	}
	if auth.ValidateCSRF(buildReq(tok, "")) {
		t.Error("missing header should be invalid")
	}
	if auth.ValidateCSRF(buildReq("", "")) {
		t.Error("both empty should be invalid")
	}
}

// ── Users ─────────────────────────────────────────────────────────────────────

func TestCreateAndGetUser(t *testing.T) {
	db := openTestDB(t)

	if err := auth.CreateUser(db, "alice", "password123", "admin", 4); err != nil {
		t.Fatal(err)
	}

	u, err := auth.GetUserByUsername(db, "alice")
	if err != nil {
		t.Fatal(err)
	}
	if u.Username != "alice" || u.Role != "admin" {
		t.Fatalf("unexpected user: %+v", u)
	}

	u2, err := auth.GetUserByID(db, u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if u2.Username != "alice" {
		t.Fatal("GetUserByID returned wrong user")
	}
}

func TestGetUserByUsername_NotFound(t *testing.T) {
	db := openTestDB(t)
	_, err := auth.GetUserByUsername(db, "nobody")
	if err != auth.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestVerifyPassword(t *testing.T) {
	db := openTestDB(t)
	_ = auth.CreateUser(db, "bob", "secret123", "readonly", 4)

	u, _ := auth.GetUserByUsername(db, "bob")
	if !auth.VerifyPassword(u.PasswordHash, "secret123") {
		t.Error("correct password should verify")
	}
	if auth.VerifyPassword(u.PasswordHash, "wrong") {
		t.Error("wrong password should not verify")
	}
}

func TestChangePassword(t *testing.T) {
	db := openTestDB(t)
	_ = auth.CreateUser(db, "carol", "oldpass12", "admin", 4)
	u, _ := auth.GetUserByUsername(db, "carol")

	if err := auth.ChangePassword(db, u.ID, "newpass99", 4); err != nil {
		t.Fatal(err)
	}
	u2, _ := auth.GetUserByUsername(db, "carol")
	if !auth.VerifyPassword(u2.PasswordHash, "newpass99") {
		t.Error("new password should verify after change")
	}
}

func TestDuplicateUser(t *testing.T) {
	db := openTestDB(t)
	_ = auth.CreateUser(db, "dave", "pass1234", "admin", 4)
	err := auth.CreateUser(db, "dave", "pass5678", "admin", 4)
	if err == nil {
		t.Error("creating duplicate user should fail")
	}
}

// ── Setup ─────────────────────────────────────────────────────────────────────

func TestSetupComplete(t *testing.T) {
	db := openTestDB(t)

	if auth.IsSetupComplete(db) {
		t.Error("fresh DB should not be setup-complete")
	}
	if err := auth.SetSetupComplete(db); err != nil {
		t.Fatal(err)
	}
	if !auth.IsSetupComplete(db) {
		t.Error("DB should be setup-complete after SetSetupComplete")
	}
	// Idempotent
	if err := auth.SetSetupComplete(db); err != nil {
		t.Fatal("SetSetupComplete should be idempotent")
	}
}

// ── Sessions ──────────────────────────────────────────────────────────────────

func TestSessionLifecycle(t *testing.T) {
	db := openTestDB(t)
	_ = auth.CreateUser(db, "eve", "pass1234", "admin", 4)
	u, _ := auth.GetUserByUsername(db, "eve")

	sid, err := auth.CreateSession(db, u.ID, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatal(err)
	}
	if len(sid) != 64 {
		t.Fatalf("expected 64-char session ID, got %d", len(sid))
	}

	sess, err := auth.GetSession(db, sid, 30)
	if err != nil {
		t.Fatal(err)
	}
	if sess.Username != "eve" || sess.Role != "admin" {
		t.Fatalf("unexpected session: %+v", sess)
	}

	if err := auth.DeleteSession(db, sid); err != nil {
		t.Fatal(err)
	}
	_, err = auth.GetSession(db, sid, 30)
	if err != auth.ErrNoSession {
		t.Fatalf("expected ErrNoSession after delete, got %v", err)
	}
}

func TestSessionNotFound(t *testing.T) {
	db := openTestDB(t)
	_, err := auth.GetSession(db, "nonexistent", 30)
	if err != auth.ErrNoSession {
		t.Fatalf("expected ErrNoSession, got %v", err)
	}
}

func TestListSessions(t *testing.T) {
	db := openTestDB(t)
	_ = auth.CreateUser(db, "frank", "pass1234", "admin", 4)
	u, _ := auth.GetUserByUsername(db, "frank")

	sid1, _ := auth.CreateSession(db, u.ID, "1.2.3.4", "agent-a")
	sid2, _ := auth.CreateSession(db, u.ID, "5.6.7.8", "agent-b")

	sessions, err := auth.ListSessions(db, u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}

	_ = auth.DeleteSession(db, sid1)
	_ = auth.DeleteSession(db, sid2)
}

// ── Brute force ───────────────────────────────────────────────────────────────

func TestBruteForceIP(t *testing.T) {
	db := openTestDB(t)
	ip := "192.0.2.1"

	if auth.IsLockedOut(db, ip) {
		t.Error("fresh IP should not be locked out")
	}

	for i := 0; i < 4; i++ {
		_ = auth.RecordFailedLogin(db, ip, 5, 15)
	}
	if auth.IsLockedOut(db, ip) {
		t.Error("should not be locked out at 4 failures (threshold=5)")
	}

	_ = auth.RecordFailedLogin(db, ip, 5, 15)
	if !auth.IsLockedOut(db, ip) {
		t.Error("should be locked out at 5 failures")
	}

	_ = auth.ResetFailedLogins(db, ip)
	if auth.IsLockedOut(db, ip) {
		t.Error("should not be locked out after reset")
	}
}

func TestBruteForceUsername(t *testing.T) {
	db := openTestDB(t)
	username := "targetuser"

	if auth.IsLockedOutUsername(db, username) {
		t.Error("fresh username should not be locked out")
	}

	for i := 0; i < 5; i++ {
		_ = auth.RecordFailedLoginUsername(db, username, 5, 15)
	}
	if !auth.IsLockedOutUsername(db, username) {
		t.Error("should be locked out at 5 failures")
	}

	_ = auth.ResetFailedLoginsUsername(db, username)
	if auth.IsLockedOutUsername(db, username) {
		t.Error("should not be locked out after reset")
	}
}

func TestBruteForceIPAndUsername_Independent(t *testing.T) {
	db := openTestDB(t)
	ip := "192.0.2.2"
	username := "victim"

	for i := 0; i < 5; i++ {
		_ = auth.RecordFailedLoginUsername(db, username, 5, 15)
	}

	// IP should not be locked even though username is
	if auth.IsLockedOut(db, ip) {
		t.Error("IP should not be locked when only username threshold reached")
	}
	if !auth.IsLockedOutUsername(db, username) {
		t.Error("username should be locked")
	}
}

// ── TOTP ──────────────────────────────────────────────────────────────────────

func TestGenerateTOTPSecret(t *testing.T) {
	secret, url, err := auth.GenerateTOTPSecret("alice", "linux-webui")
	if err != nil {
		t.Fatal(err)
	}
	if secret == "" {
		t.Error("expected non-empty secret")
	}
	if !strings.HasPrefix(url, "otpauth://totp/") {
		t.Errorf("expected otpauth URL, got %q", url)
	}
}

func TestVerifyTOTP_Invalid(t *testing.T) {
	// Non-numeric and too-short codes must fail.
	if auth.VerifyTOTP("JBSWY3DPEHPK3PXP", "abc") {
		t.Error("non-numeric code should be invalid")
	}
	if auth.VerifyTOTP("JBSWY3DPEHPK3PXP", "") {
		t.Error("empty code should be invalid")
	}
}

func TestSaveTOTPSecret(t *testing.T) {
	db := openTestDB(t)
	_ = auth.CreateUser(db, "henry", "pass1234", "admin", 4)
	u, _ := auth.GetUserByUsername(db, "henry")

	secret, _, _ := auth.GenerateTOTPSecret("henry", "linux-webui")
	codes, err := auth.SaveTOTPSecret(db, u.ID, secret, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(codes) != 8 {
		t.Fatalf("expected 8 recovery codes, got %d", len(codes))
	}
	for _, c := range codes {
		if len(c) == 0 {
			t.Error("recovery code should be non-empty")
		}
		if !strings.Contains(c, "-") {
			t.Errorf("recovery code should contain dash, got %q", c)
		}
	}

	u2, _ := auth.GetUserByUsername(db, "henry")
	if u2.TOTPSecret != secret {
		t.Error("TOTP secret not persisted")
	}
}

func TestUseRecoveryCode(t *testing.T) {
	db := openTestDB(t)
	_ = auth.CreateUser(db, "iris", "pass1234", "admin", 4)
	u, _ := auth.GetUserByUsername(db, "iris")

	secret, _, _ := auth.GenerateTOTPSecret("iris", "linux-webui")
	codes, _ := auth.SaveTOTPSecret(db, u.ID, secret, 4)

	// First use succeeds.
	if err := auth.UseRecoveryCode(db, u.ID, codes[0]); err != nil {
		t.Fatalf("UseRecoveryCode failed: %v", err)
	}
	// Same code rejected on reuse.
	if err := auth.UseRecoveryCode(db, u.ID, codes[0]); err != auth.ErrInvalidTOTP {
		t.Errorf("reused recovery code should return ErrInvalidTOTP, got %v", err)
	}
	// Wrong code rejected.
	if err := auth.UseRecoveryCode(db, u.ID, "INVALID-CODE"); err != auth.ErrInvalidTOTP {
		t.Errorf("invalid recovery code should return ErrInvalidTOTP, got %v", err)
	}
}

func TestRevokeTOTP(t *testing.T) {
	db := openTestDB(t)
	_ = auth.CreateUser(db, "james", "pass1234", "admin", 4)
	u, _ := auth.GetUserByUsername(db, "james")

	secret, _, _ := auth.GenerateTOTPSecret("james", "linux-webui")
	_, _ = auth.SaveTOTPSecret(db, u.ID, secret, 4)

	if err := auth.RevokeTOTP(db, u.ID); err != nil {
		t.Fatal(err)
	}
	u2, _ := auth.GetUserByUsername(db, "james")
	if u2.TOTPSecret != "" {
		t.Error("TOTP secret should be empty after revoke")
	}
}

// ── Additional SQL coverage ────────────────────────────────────────────────────

func TestListUsers(t *testing.T) {
	db := openTestDB(t)
	_ = auth.CreateUser(db, "kate", "pass1234", "admin", 4)
	_ = auth.CreateUser(db, "leo", "pass1234", "readonly", 4)

	users, err := auth.ListUsers(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(users) < 2 {
		t.Fatalf("expected at least 2 users, got %d", len(users))
	}
}

func TestListAllSessions(t *testing.T) {
	db := openTestDB(t)
	_ = auth.CreateUser(db, "mary", "pass1234", "admin", 4)
	u, _ := auth.GetUserByUsername(db, "mary")

	_, _ = auth.CreateSession(db, u.ID, "1.2.3.4", "agent-a")
	_, _ = auth.CreateSession(db, u.ID, "5.6.7.8", "agent-b")

	sessions, err := auth.ListAllSessions(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) < 2 {
		t.Fatalf("expected at least 2 sessions, got %d", len(sessions))
	}
	if _, ok := sessions[0]["username"]; !ok {
		t.Error("session entry should include username field")
	}
}

func TestDeleteUserSessions(t *testing.T) {
	db := openTestDB(t)
	_ = auth.CreateUser(db, "nick", "pass1234", "admin", 4)
	u, _ := auth.GetUserByUsername(db, "nick")

	_, _ = auth.CreateSession(db, u.ID, "1.2.3.4", "agent")
	_, _ = auth.CreateSession(db, u.ID, "5.6.7.8", "agent")

	if err := auth.DeleteUserSessions(db, u.ID); err != nil {
		t.Fatal(err)
	}
	sessions, _ := auth.ListSessions(db, u.ID)
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions after DeleteUserSessions, got %d", len(sessions))
	}
}

func TestPruneAuditLog(t *testing.T) {
	db := openTestDB(t)
	_ = auth.CreateUser(db, "olive", "pass1234", "admin", 4)
	u, _ := auth.GetUserByUsername(db, "olive")
	auth.LogAction(db, u.ID, "olive", "test_action", "", "127.0.0.1")

	// Zero days = no-op.
	if err := auth.PruneAuditLog(db, 0); err != nil {
		t.Fatal(err)
	}
	entries, _ := auth.GetAuditLog(db, 100)
	if len(entries) == 0 {
		t.Error("entries should survive a 0-day prune")
	}

	// 30-day retention won't delete a just-created entry.
	if err := auth.PruneAuditLog(db, 30); err != nil {
		t.Fatal(err)
	}
	entries2, _ := auth.GetAuditLog(db, 100)
	if len(entries2) == 0 {
		t.Error("recent entries should survive 30-day prune")
	}
}

// ── Audit log ─────────────────────────────────────────────────────────────────

func TestAuditLog(t *testing.T) {
	db := openTestDB(t)
	_ = auth.CreateUser(db, "grace", "pass1234", "admin", 4)
	u, _ := auth.GetUserByUsername(db, "grace")

	auth.LogAction(db, u.ID, "grace", "login", "", "127.0.0.1")
	auth.LogAction(db, u.ID, "grace", "logout", "", "127.0.0.1")

	entries, err := auth.GetAuditLog(db, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) < 2 {
		t.Fatalf("expected at least 2 audit entries, got %d", len(entries))
	}
}
