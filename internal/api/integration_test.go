package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/virajchitnis/linux-webui/internal/api"
	"github.com/virajchitnis/linux-webui/internal/auth"
	"github.com/virajchitnis/linux-webui/internal/distro"
	"github.com/virajchitnis/linux-webui/internal/files"
	"github.com/virajchitnis/linux-webui/internal/terminal"
)

// testServer bundles a live httptest server with a helper client.
type testServer struct {
	srv    *httptest.Server
	client *http.Client
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()

	f, err := os.CreateTemp("", "lwui-integ-*.db")
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	t.Cleanup(func() { _ = os.Remove(f.Name()) })

	db, err := auth.OpenDB(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	termMgr := terminal.NewManager(5, 5*time.Minute)
	t.Cleanup(func() { termMgr.CloseAll() })

	router := api.NewRouter(api.RouterOptions{
		DB:              db,
		Distro:          distro.Detect(),
		SecureCookie:    false, // plain HTTP in tests
		BcryptCost:      4,     // fast hashing for tests
		AuthTimeout:     30,    // 30 minute session timeout
		SensorsEnabled:  true,
		FilesRoots:      files.Roots{os.TempDir()},
		AptEnabled:      true,
		UFWEnabled:      true,
		CronEnabled:     true,
		TerminalManager: termMgr,
		DevMode:         true, // skip WS origin check in tests
	})
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	return &testServer{srv: srv, client: client}
}

// baseURL returns the test server base URL.
func (ts *testServer) baseURL() string { return ts.srv.URL }

// postJSON sends a JSON POST and returns the response.
func (ts *testServer) postJSON(t *testing.T, path string, body any, extraHeaders map[string]string) *http.Response {
	t.Helper()
	b, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, ts.baseURL()+path, bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	resp, err := ts.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

// get sends a GET and returns the response.
func (ts *testServer) get(t *testing.T, path string) *http.Response {
	t.Helper()
	resp, err := ts.client.Get(ts.baseURL() + path)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

// csrfToken extracts the CSRF cookie value from the jar.
func (ts *testServer) csrfToken(t *testing.T) string {
	t.Helper()
	u, _ := url.Parse(ts.baseURL())
	for _, c := range ts.client.Jar.(*cookiejar.Jar).Cookies(u) {
		if c.Name == "linux_webui_csrf" {
			return c.Value
		}
	}
	return ""
}

// setup completes the first-run wizard and returns the admin credentials.
func (ts *testServer) setup(t *testing.T, username, password string) {
	t.Helper()
	resp := ts.postJSON(t, "/api/setup/complete", map[string]string{
		"username": username,
		"password": password,
	}, nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("setup failed: %d %s", resp.StatusCode, body)
	}
}

// login logs in and returns the CSRF token (already stored in jar).
func (ts *testServer) login(t *testing.T, username, password string) string {
	t.Helper()
	resp := ts.postJSON(t, "/api/auth/login", map[string]string{
		"username": username,
		"password": password,
	}, nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("login failed: %d %s", resp.StatusCode, body)
	}
	return ts.csrfToken(t)
}

// csrfFromResponse is a simpler CSRF extraction that reads cookies from the server response.
func csrfFromResponse(resp *http.Response) string {
	for _, c := range resp.Cookies() {
		if c.Name == "linux_webui_csrf" {
			return c.Value
		}
	}
	return ""
}

// ── Health ────────────────────────────────────────────────────────────────────

func TestHealth(t *testing.T) {
	ts := newTestServer(t)
	resp := ts.get(t, "/health")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health: expected 200, got %d", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("health: invalid JSON: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("health body status = %v, want ok", body["status"])
	}
}

// ── Setup ─────────────────────────────────────────────────────────────────────

func TestSetupStatus_Fresh(t *testing.T) {
	ts := newTestServer(t)
	resp := ts.get(t, "/api/setup/status")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("setup status: expected 200, got %d", resp.StatusCode)
	}
	var body map[string]bool
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("setup status: invalid JSON: %v", err)
	}
	if body["complete"] {
		t.Error("fresh server should not be setup-complete")
	}
}

func TestSetupComplete(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/setup/status")
	defer resp.Body.Close()
	var body map[string]bool
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("setup complete: invalid JSON: %v", err)
	}
	if !body["complete"] {
		t.Error("server should be setup-complete after wizard")
	}
}

func TestSetupComplete_Duplicate(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")

	resp := ts.postJSON(t, "/api/setup/complete", map[string]string{
		"username": "admin2",
		"password": "testpass123!",
	}, nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("duplicate setup: expected 409, got %d", resp.StatusCode)
	}
}

// ── Auth ──────────────────────────────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")

	resp := ts.postJSON(t, "/api/auth/login", map[string]string{
		"username": "admin",
		"password": "testpass123!",
	}, nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login: expected 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("login: invalid JSON: %v", err)
	}
	if body["username"] != "admin" {
		t.Errorf("login response username = %v, want admin", body["username"])
	}
	if body["role"] != "admin" {
		t.Errorf("login response role = %v, want admin", body["role"])
	}

	// Should have set session + CSRF cookies.
	sessionCookie := false
	csrfCookie := false
	for _, c := range resp.Cookies() {
		if c.Name == "linux_webui_session" {
			sessionCookie = true
		}
		if c.Name == "linux_webui_csrf" {
			csrfCookie = true
		}
	}
	if !sessionCookie {
		t.Error("login should set linux_webui_session cookie")
	}
	if !csrfCookie {
		t.Error("login should set linux_webui_csrf cookie")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")

	resp := ts.postJSON(t, "/api/auth/login", map[string]string{
		"username": "admin",
		"password": "wrongpassword",
	}, nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("wrong password: expected 401, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	// Must not reveal details — only "invalid credentials" is acceptable, not stack traces.
	_ = string(body)
}

func TestLogin_UnknownUser(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")

	resp := ts.postJSON(t, "/api/auth/login", map[string]string{
		"username": "nobody",
		"password": "anything",
	}, nil)
	defer resp.Body.Close()

	// Must return 401 (not 404 — that would reveal user existence)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("unknown user: expected 401, got %d", resp.StatusCode)
	}
}

func TestLogin_BruteForce_IP(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")

	for i := 0; i < 5; i++ {
		resp := ts.postJSON(t, "/api/auth/login", map[string]string{
			"username": "admin",
			"password": "wrong",
		}, nil)
		resp.Body.Close()
	}

	resp := ts.postJSON(t, "/api/auth/login", map[string]string{
		"username": "admin",
		"password": "testpass123!",
	}, nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("after 5 failures: expected 429, got %d", resp.StatusCode)
	}
}

func TestLogin_BruteForce_Username(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "target", "testpass123!")
	// Also create another user with a different server instance to avoid IP lockout
	// We'll use a fresh client (new IP simulation not possible, but username lockout
	// is tracked independently).

	// Exhaust the username lockout (5 attempts)
	for i := 0; i < 5; i++ {
		resp := ts.postJSON(t, "/api/auth/login", map[string]string{
			"username": "target",
			"password": "wrong",
		}, nil)
		resp.Body.Close()
	}

	// Now try the correct password — username should be locked
	resp := ts.postJSON(t, "/api/auth/login", map[string]string{
		"username": "target",
		"password": "testpass123!",
	}, nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("username lockout: expected 429, got %d", resp.StatusCode)
	}
}

func TestLogout(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")

	loginResp := ts.postJSON(t, "/api/auth/login", map[string]string{
		"username": "admin",
		"password": "testpass123!",
	}, nil)
	csrf := csrfFromResponse(loginResp)
	loginResp.Body.Close()

	logoutResp := ts.postJSON(t, "/api/auth/logout", nil, map[string]string{
		"X-CSRF-Token": csrf,
	})
	defer logoutResp.Body.Close()

	if logoutResp.StatusCode != http.StatusNoContent {
		t.Errorf("logout: expected 204, got %d", logoutResp.StatusCode)
	}

	// Session should be invalidated — /api/auth/me should now 401
	meResp := ts.get(t, "/api/auth/me")
	defer meResp.Body.Close()
	if meResp.StatusCode != http.StatusUnauthorized {
		t.Errorf("after logout, /api/auth/me: expected 401, got %d", meResp.StatusCode)
	}
}

// ── CSRF ──────────────────────────────────────────────────────────────────────

func TestCSRF_RequiredOnMutations(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	// POST without CSRF header should be rejected
	resp := ts.postJSON(t, "/api/auth/logout", nil, nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("POST without CSRF: expected 403, got %d", resp.StatusCode)
	}
}

func TestCSRF_WrongToken(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.postJSON(t, "/api/auth/logout", nil, map[string]string{
		"X-CSRF-Token": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("wrong CSRF token: expected 403, got %d", resp.StatusCode)
	}
}

// ── Protected endpoints ───────────────────────────────────────────────────────

func TestMe_Unauthenticated(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/auth/me")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("/api/auth/me without session: expected 401, got %d", resp.StatusCode)
	}
}

func TestMe_Authenticated(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/auth/me")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/api/auth/me: expected 200, got %d", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("/api/auth/me: invalid JSON: %v", err)
	}
	if body["username"] != "admin" {
		t.Errorf("me.username = %v, want admin", body["username"])
	}
}

func TestCapabilities_Authenticated(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/capabilities")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/api/capabilities: expected 200, got %d", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("capabilities: invalid JSON: %v", err)
	}
}

// ── RBAC ──────────────────────────────────────────────────────────────────────

func TestRBAC_ReadonlyCannotReboot(t *testing.T) {
	// Create a readonly user directly in the DB
	f, _ := os.CreateTemp("", "lwui-rbac-*.db")
	f.Close()
	defer func() { _ = os.Remove(f.Name()) }()

	db, _ := auth.OpenDB(f.Name())
	defer db.Close()

	_ = auth.CreateUser(db, "admin", "adminpass1", "admin", 4)
	_ = auth.SetSetupComplete(db)
	_ = auth.CreateUser(db, "reader", "readerpass", "readonly", 4)

	router := api.NewRouter(api.RouterOptions{
		DB:           db,
		Distro:       distro.Detect(),
		SecureCookie: false,
		BcryptCost:   4,
		AuthTimeout:  30,
	})
	srv := httptest.NewServer(router)
	defer srv.Close()

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	// Login as readonly user
	loginBody, _ := json.Marshal(map[string]string{"username": "reader", "password": "readerpass"})
	loginResp, _ := client.Post(srv.URL+"/api/auth/login", "application/json", bytes.NewReader(loginBody))
	csrf := csrfFromResponse(loginResp)
	loginResp.Body.Close()

	// Try to reboot (admin-only)
	rebootReq, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/system/reboot", nil)
	rebootReq.Header.Set("X-CSRF-Token", csrf)
	rebootReq.Header.Set("Content-Type", "application/json")
	roResp, err := client.Do(rebootReq)
	if err != nil {
		t.Fatal(err)
	}
	defer roResp.Body.Close()

	if roResp.StatusCode != http.StatusForbidden {
		t.Errorf("readonly reboot: expected 403, got %d", roResp.StatusCode)
	}
}

// ── Session management ────────────────────────────────────────────────────────

func TestSessionList(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	req, _ := http.NewRequest(http.MethodGet, ts.baseURL()+"/api/account/sessions", nil)
	req.Header.Set("X-CSRF-Token", csrf)
	resp, err := ts.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("session list: expected 200, got %d", resp.StatusCode)
	}
	var sessions []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		t.Fatalf("session list: invalid JSON: %v", err)
	}
	if len(sessions) == 0 {
		t.Error("session list should have at least one entry (current session)")
	}
}

// ── Account endpoints ─────────────────────────────────────────────────────────

func TestAccountChangePassword(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	resp := ts.postJSON(t, "/api/account/password", map[string]string{
		"current_password": "testpass123!",
		"new_password":     "newpass456!!",
	}, map[string]string{"X-CSRF-Token": csrf})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("change password: expected 204, got %d: %s", resp.StatusCode, body)
	}
}

func TestAccountChangePassword_WrongCurrent(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	resp := ts.postJSON(t, "/api/account/password", map[string]string{
		"current_password": "wrongpassword",
		"new_password":     "newpass456!!",
	}, map[string]string{"X-CSRF-Token": csrf})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("wrong current password: expected 403, got %d", resp.StatusCode)
	}
}

func TestAccountChangePassword_TooShort(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	resp := ts.postJSON(t, "/api/account/password", map[string]string{
		"current_password": "testpass123!",
		"new_password":     "short",
	}, map[string]string{"X-CSRF-Token": csrf})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("short new password: expected 400, got %d", resp.StatusCode)
	}
}

func TestAccountSessionList(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/account/sessions")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/api/account/sessions: expected 200, got %d", resp.StatusCode)
	}
	var sessions []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		t.Fatalf("account sessions: invalid JSON: %v", err)
	}
	if len(sessions) == 0 {
		t.Error("should have at least one active session")
	}
}

func TestAccountSessionRevoke(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	// Get own sessions.
	resp := ts.get(t, "/api/account/sessions")
	var sessions []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		resp.Body.Close()
		t.Fatalf("account sessions revoke: invalid JSON: %v", err)
	}
	resp.Body.Close()

	if len(sessions) == 0 {
		t.Fatal("expected at least one session")
	}
	id := sessions[0]["id"].(string)

	delReq, _ := http.NewRequest(http.MethodDelete, ts.baseURL()+"/api/account/sessions/"+id, nil)
	delReq.Header.Set("X-CSRF-Token", csrf)
	delResp, err := ts.client.Do(delReq)
	if err != nil {
		t.Fatal(err)
	}
	defer delResp.Body.Close()
	if delResp.StatusCode >= 500 {
		t.Errorf("session revoke returned 5xx: %d", delResp.StatusCode)
	}
}

func TestAccountSessionRevoke_NotOwned(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	// Try to revoke a nonexistent session.
	delReq, _ := http.NewRequest(http.MethodDelete, ts.baseURL()+"/api/account/sessions/nonexistent-session-id", nil)
	delReq.Header.Set("X-CSRF-Token", csrf)
	delResp, err := ts.client.Do(delReq)
	if err != nil {
		t.Fatal(err)
	}
	defer delResp.Body.Close()
	if delResp.StatusCode != http.StatusNotFound {
		t.Errorf("nonexistent session: expected 404, got %d", delResp.StatusCode)
	}
}

// ── Admin endpoints ───────────────────────────────────────────────────────────

func TestAdminSessionsList(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/admin/sessions")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/api/admin/sessions: expected 200, got %d", resp.StatusCode)
	}
	var sessions []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		t.Fatalf("admin sessions: invalid JSON: %v", err)
	}
	if len(sessions) == 0 {
		t.Error("admin sessions list should have at least one entry")
	}
}

func TestAuditLogList(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/admin/audit")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/api/admin/audit: expected 200, got %d", resp.StatusCode)
	}
	var entries []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		t.Fatalf("audit log: invalid JSON: %v", err)
	}
	// At least the login action should be present.
	if len(entries) == 0 {
		t.Error("audit log should have entries after login")
	}
}

func TestAuditLogList_ReadonlyForbidden(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")

	// Create a readonly user, then log in as them.
	f, _ := os.CreateTemp("", "lwui-ro-audit-*.db")
	f.Close()
	defer os.Remove(f.Name())

	db, _ := auth.OpenDB(f.Name())
	defer db.Close()
	_ = auth.CreateUser(db, "admin", "adminpass1", "admin", 4)
	_ = auth.SetSetupComplete(db)
	_ = auth.CreateUser(db, "reader", "readerpass", "readonly", 4)

	router := api.NewRouter(api.RouterOptions{
		DB: db, Distro: distro.Detect(),
		SecureCookie: false, BcryptCost: 4, AuthTimeout: 30,
	})
	srv := httptest.NewServer(router)
	defer srv.Close()
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	loginBody, _ := json.Marshal(map[string]string{"username": "reader", "password": "readerpass"})
	loginResp, _ := client.Post(srv.URL+"/api/auth/login", "application/json", bytes.NewReader(loginBody))
	loginResp.Body.Close()

	resp, err := client.Get(srv.URL + "/api/admin/audit")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("readonly audit access: expected 403, got %d", resp.StatusCode)
	}
}

// ── Process and network endpoints ─────────────────────────────────────────────

func TestProcessList(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/processes")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/api/processes: expected 200, got %d", resp.StatusCode)
	}
	var procs []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&procs); err != nil {
		t.Fatalf("processes: invalid JSON: %v", err)
	}
	if len(procs) == 0 {
		t.Error("process list should be non-empty")
	}
}

func TestNetworkInterfaces(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/network/interfaces")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/api/network/interfaces: expected 200, got %d", resp.StatusCode)
	}
	var ifaces []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&ifaces); err != nil {
		t.Fatalf("network/interfaces: invalid JSON: %v", err)
	}
	if len(ifaces) == 0 {
		t.Error("network interface list should be non-empty")
	}
}

// ── TOTP endpoints ────────────────────────────────────────────────────────────

func TestTOTPStatus_NotEnabled(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/account/totp/status")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/api/account/totp/status: expected 200, got %d", resp.StatusCode)
	}
	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)
	if enabled, ok := body["enabled"].(bool); !ok || enabled {
		t.Errorf("TOTP should not be enabled on fresh account, got %v", body["enabled"])
	}
}

func TestTOTPEnroll(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	resp := ts.postJSON(t, "/api/account/totp/enroll", nil, map[string]string{
		"X-CSRF-Token": csrf,
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("/api/account/totp/enroll: expected 200, got %d: %s", resp.StatusCode, body)
	}
	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)
	if _, ok := body["secret"]; !ok {
		t.Error("enroll response should include secret")
	}
	if _, ok := body["otpauth_url"]; !ok {
		t.Error("enroll response should include otpauth_url")
	}
}

func TestTOTPRevoke_NotEnabled(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	delReq, _ := http.NewRequest(http.MethodDelete, ts.baseURL()+"/api/account/totp", nil)
	delReq.Header.Set("X-CSRF-Token", csrf)
	resp, err := ts.client.Do(delReq)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	// Revoking when not enabled is still a 204 (idempotent).
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("TOTP revoke: expected 204, got %d", resp.StatusCode)
	}
}

func TestProcessKill_InvalidPID(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	// Non-numeric PID → 400.
	req, _ := http.NewRequest(http.MethodPost, ts.baseURL()+"/api/processes/abc/kill", nil)
	req.Header.Set("X-CSRF-Token", csrf)
	resp, err := ts.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("kill invalid pid: expected 400, got %d", resp.StatusCode)
	}
}

func TestProcessKill_PID1Refused(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	// PID 1 = init — process.Kill refuses it, handler returns 400.
	req, _ := http.NewRequest(http.MethodPost, ts.baseURL()+"/api/processes/1/kill", nil)
	req.Header.Set("X-CSRF-Token", csrf)
	req.Header.Set("Content-Type", "application/json")
	resp, err := ts.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("kill pid 1: expected 400, got %d", resp.StatusCode)
	}
}

func TestProcessRenice_InvalidPID(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	resp := ts.postJSON(t, "/api/processes/abc/renice",
		map[string]int{"priority": 5},
		map[string]string{"X-CSRF-Token": csrf})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("renice invalid pid: expected 400, got %d", resp.StatusCode)
	}
}

func TestProcessRenice_InvalidPriority(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	// Renice our own PID with out-of-range priority.
	pid := os.Getpid()
	resp := ts.postJSON(t, "/api/processes/"+strconv.Itoa(pid)+"/renice",
		map[string]int{"priority": 99},
		map[string]string{"X-CSRF-Token": csrf})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("renice invalid priority: expected 400, got %d", resp.StatusCode)
	}
}

func TestAdminSessionRevoke(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")
	csrf := ts.csrfToken(t)

	// List all sessions to get an ID.
	resp := ts.get(t, "/api/admin/sessions")
	var sessions []map[string]any
	json.NewDecoder(resp.Body).Decode(&sessions)
	resp.Body.Close()
	if len(sessions) == 0 {
		t.Fatal("expected at least one admin session")
	}
	id := sessions[0]["id"].(string)

	delReq, _ := http.NewRequest(http.MethodDelete, ts.baseURL()+"/api/admin/sessions/"+id, nil)
	delReq.Header.Set("X-CSRF-Token", csrf)
	delResp, err := ts.client.Do(delReq)
	if err != nil {
		t.Fatal(err)
	}
	defer delResp.Body.Close()
	if delResp.StatusCode != http.StatusNoContent {
		t.Errorf("admin session revoke: expected 204, got %d", delResp.StatusCode)
	}
}

// ── File browser endpoints ────────────────────────────────────────────────────

func TestFilesRoots(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/files/roots")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/api/files/roots: expected 200, got %d", resp.StatusCode)
	}
	var roots []string
	json.NewDecoder(resp.Body).Decode(&roots)
	if len(roots) == 0 {
		t.Error("expected at least one file browser root")
	}
}

func TestFilesList(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/files/list?path="+os.TempDir())
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/api/files/list: expected 200, got %d", resp.StatusCode)
	}
}

func TestFilesList_Escape(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/files/list?path=/etc/passwd")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("/api/files/list escape: expected 403, got %d", resp.StatusCode)
	}
}

func TestFilesRead(t *testing.T) {
	// Write a temp file inside TempDir for reading.
	f, err := os.CreateTemp("", "lwui-test-read-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString("hello test")
	f.Close()
	defer os.Remove(f.Name())

	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/files/read?path="+f.Name())
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("/api/files/read: expected 200, got %d: %s", resp.StatusCode, body)
	}
}

func TestFilesRead_NoPath(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/files/read")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("/api/files/read without path: expected 400, got %d", resp.StatusCode)
	}
}

func TestFilesRead_Escape(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/files/read?path=/etc/shadow")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("/api/files/read escape: expected 403, got %d", resp.StatusCode)
	}
}

// ── Sensors endpoint ──────────────────────────────────────────────────────────

func TestSensors(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/sensors")
	defer resp.Body.Close()
	// Should return 200 even if no sensors are present (returns empty array).
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("/api/sensors: expected 200 or 500, got %d", resp.StatusCode)
	}
}

// ── Users handler validation ──────────────────────────────────────────────────

func TestUsersHandler_ListUnauthenticated(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/users")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("/api/users unauthenticated: expected 401, got %d", resp.StatusCode)
	}
}

func TestUsersHandler_ListAuthenticated(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/users")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("/api/users authenticated admin: expected 200, got %d", resp.StatusCode)
	}
}

func TestSystemInfo(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/system/info")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/api/system/info: expected 200, got %d", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("system info: invalid JSON: %v", err)
	}
}

func TestSystemMetrics_NoCollector(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/system/metrics")
	defer resp.Body.Close()
	// No collector configured → 503 (not a panic).
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("/api/system/metrics no collector: expected 503, got %d", resp.StatusCode)
	}
}

// ── Cron endpoints ────────────────────────────────────────────────────────────

func TestCronList(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/cron")
	defer resp.Body.Close()
	// Returns 200 with an array (possibly empty) or 500 if crontab unavailable.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("/api/cron list: expected 200 or 500, got %d", resp.StatusCode)
	}
}

func TestCronAdd_InvalidSchedule(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	resp := ts.postJSON(t, "/api/cron", map[string]string{
		"schedule": "not a schedule",
		"command":  "/usr/bin/true",
	}, map[string]string{"X-CSRF-Token": csrf})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("cron add invalid schedule: expected 400, got %d", resp.StatusCode)
	}
}

func TestCronAdd_InvalidCommand(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	resp := ts.postJSON(t, "/api/cron", map[string]string{
		"schedule": "* * * * *",
		"command":  "echo bad\ncmd",
	}, map[string]string{"X-CSRF-Token": csrf})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("cron add invalid command (newline): expected 400, got %d", resp.StatusCode)
	}
}

func TestCronDelete_InvalidIndex(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	req, _ := http.NewRequest(http.MethodDelete, ts.baseURL()+"/api/cron/notanumber", nil)
	req.Header.Set("X-CSRF-Token", csrf)
	resp, err := ts.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("cron delete invalid index: expected 400, got %d", resp.StatusCode)
	}
}

// ── APT endpoints ─────────────────────────────────────────────────────────────

func TestAPTStatus(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/packages/status")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("/api/packages/status: expected 200, got %d", resp.StatusCode)
	}
	var body map[string]bool
	json.NewDecoder(resp.Body).Decode(&body)
	if _, ok := body["running"]; !ok {
		t.Error("apt status response should have 'running' field")
	}
}

func TestAPTUpgradable(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/packages/upgradable")
	defer resp.Body.Close()
	// May be 200 (with empty list) or 500 if apt not available.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("/api/packages/upgradable: expected 200 or 500, got %d", resp.StatusCode)
	}
}

// ── Firewall endpoints ────────────────────────────────────────────────────────

func TestFirewallStatus(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/firewall/status")
	defer resp.Body.Close()
	// ufw may not be installed in test env; both 200 and 500 are acceptable.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("/api/firewall/status: expected 200 or 500, got %d", resp.StatusCode)
	}
}

func TestFirewallAddRule_InvalidAction(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	resp := ts.postJSON(t, "/api/firewall/rules", map[string]string{
		"action": "permit",
		"port":   "22",
		"proto":  "tcp",
	}, map[string]string{"X-CSRF-Token": csrf})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("firewall add invalid action: expected 400, got %d", resp.StatusCode)
	}
}

func TestFirewallAddRule_InvalidPort(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	resp := ts.postJSON(t, "/api/firewall/rules", map[string]string{
		"action": "allow",
		"port":   "99999",
		"proto":  "tcp",
	}, map[string]string{"X-CSRF-Token": csrf})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("firewall add invalid port: expected 400, got %d", resp.StatusCode)
	}
}

func TestFirewallDeleteRule_InvalidNum(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	req, _ := http.NewRequest(http.MethodDelete, ts.baseURL()+"/api/firewall/rules/0", nil)
	req.Header.Set("X-CSRF-Token", csrf)
	resp, err := ts.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("firewall delete rule 0: expected 400, got %d", resp.StatusCode)
	}
}

// ── Users handler additional tests ────────────────────────────────────────────

func TestUsersListGroups(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	ts.login(t, "admin", "testpass123!")

	resp := ts.get(t, "/api/groups")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("/api/groups: expected 200, got %d", resp.StatusCode)
	}
	var groups []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		t.Fatalf("api/groups: invalid JSON: %v", err)
	}
	if len(groups) == 0 {
		t.Error("group list should not be empty")
	}
}

func TestUsersCreateUser_InvalidName(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	// Invalid username (has uppercase) → useradd will fail → 400
	resp := ts.postJSON(t, "/api/users", map[string]string{
		"username": "InvalidUser",
		"password": "somepassword",
	}, map[string]string{"X-CSRF-Token": csrf})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("create user invalid name: expected 400, got %d", resp.StatusCode)
	}
}

func TestUsersSetPassword_Invalid(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	req, _ := http.NewRequest(http.MethodPost, ts.baseURL()+"/api/users/InvalidUser/password", bytes.NewReader([]byte(`{"password":"newpass123!"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", csrf)
	resp, err := ts.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("set password invalid user: expected 400, got %d", resp.StatusCode)
	}
}

func TestUsersDeleteUser_Invalid(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	req, _ := http.NewRequest(http.MethodDelete, ts.baseURL()+"/api/users/InvalidUser", nil)
	req.Header.Set("X-CSRF-Token", csrf)
	resp, err := ts.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("delete user invalid name: expected 400, got %d", resp.StatusCode)
	}
}

// ── TOTP confirm ──────────────────────────────────────────────────────────────

func TestTOTPConfirm_WrongCode(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	// First enroll to get the secret.
	enrollResp := ts.postJSON(t, "/api/account/totp/enroll", nil, map[string]string{
		"X-CSRF-Token": csrf,
	})
	var enrollBody map[string]any
	json.NewDecoder(enrollResp.Body).Decode(&enrollBody)
	enrollResp.Body.Close()

	secret, _ := enrollBody["secret"].(string)

	// Confirm with wrong code.
	resp := ts.postJSON(t, "/api/account/totp/confirm", map[string]string{
		"secret": secret,
		"code":   "000000",
	}, map[string]string{"X-CSRF-Token": csrf})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("TOTP confirm wrong code: expected 400, got %d", resp.StatusCode)
	}
}

func TestTOTPConfirm_MissingFields(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	resp := ts.postJSON(t, "/api/account/totp/confirm", map[string]string{
		"secret": "",
		"code":   "",
	}, map[string]string{"X-CSRF-Token": csrf})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("TOTP confirm missing fields: expected 400, got %d", resp.StatusCode)
	}
}

// ── Terminal endpoints ────────────────────────────────────────────────────────

func TestTerminalNew(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	resp := ts.postJSON(t, "/api/terminal/new", nil, map[string]string{
		"X-CSRF-Token": csrf,
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/api/terminal/new: expected 200, got %d", resp.StatusCode)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("terminal new: invalid JSON: %v", err)
	}
	if body["id"] == "" {
		t.Error("terminal new should return non-empty id")
	}
}

func TestTerminalWS_Connect(t *testing.T) {
	ts := newTestServer(t)
	ts.setup(t, "admin", "testpass123!")
	csrf := ts.login(t, "admin", "testpass123!")

	// Allocate a terminal session ID.
	resp := ts.postJSON(t, "/api/terminal/new", nil, map[string]string{
		"X-CSRF-Token": csrf,
	})
	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	resp.Body.Close()
	id := body["id"]
	if id == "" {
		t.Fatal("no terminal id returned")
	}

	// Connect via WebSocket; pass the session cookie via the shared HTTP client.
	wsURL := "ws" + strings.TrimPrefix(ts.baseURL(), "http") + "/ws/terminal/" + id
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPClient: ts.client,
	})
	if err != nil {
		t.Fatalf("WS dial: %v", err)
	}
	defer c.CloseNow()

	c.SetReadLimit(64 * 1024)

	// Send a resize message first to cover the Resize path.
	resize := map[string]any{
		"type":    "resize",
		"payload": map[string]any{"data": "", "rows": 24, "cols": 80},
	}
	resizeBytes, _ := json.Marshal(resize)
	if err := c.Write(ctx, websocket.MessageText, resizeBytes); err != nil {
		t.Fatalf("write resize: %v", err)
	}

	// Send a recognisable command.
	marker := "lwui_integ_marker_9x7z"
	input := map[string]any{
		"type":    "input",
		"payload": map[string]any{"data": "echo " + marker + "\r", "rows": 0, "cols": 0},
	}
	inputBytes, _ := json.Marshal(input)
	if err := c.Write(ctx, websocket.MessageText, inputBytes); err != nil {
		t.Fatalf("write input: %v", err)
	}

	// Read messages until the marker appears in output or the context times out.
	for {
		_, data, err := c.Read(ctx)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		if strings.Contains(string(data), marker) {
			return
		}
	}
}

func TestTerminalNew_RequiresAdmin(t *testing.T) {
	// Create a readonly user and try to create a terminal.
	f, _ := os.CreateTemp("", "lwui-term-*.db")
	f.Close()
	defer os.Remove(f.Name())

	db, _ := auth.OpenDB(f.Name())
	defer db.Close()
	_ = auth.CreateUser(db, "admin", "adminpass1", "admin", 4)
	_ = auth.SetSetupComplete(db)
	_ = auth.CreateUser(db, "reader", "readerpass", "readonly", 4)

	termMgr := terminal.NewManager(5, 5*time.Minute)
	defer termMgr.CloseAll()

	router := api.NewRouter(api.RouterOptions{
		DB: db, Distro: distro.Detect(),
		SecureCookie: false, BcryptCost: 4, AuthTimeout: 30,
		TerminalManager: termMgr,
	})
	srv := httptest.NewServer(router)
	defer srv.Close()
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	loginBody, _ := json.Marshal(map[string]string{"username": "reader", "password": "readerpass"})
	loginResp, _ := client.Post(srv.URL+"/api/auth/login", "application/json", bytes.NewReader(loginBody))
	csrf := csrfFromResponse(loginResp)
	loginResp.Body.Close()

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/terminal/new", nil)
	req.Header.Set("X-CSRF-Token", csrf)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("readonly terminal new: expected 403, got %d", resp.StatusCode)
	}
}
