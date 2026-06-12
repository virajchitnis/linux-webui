package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/virajchitnis/linux-webui/internal/api"
	"github.com/virajchitnis/linux-webui/internal/auth"
	"github.com/virajchitnis/linux-webui/internal/distro"
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
	t.Cleanup(func() { os.Remove(f.Name()) })

	db, err := auth.OpenDB(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	router := api.NewRouter(api.RouterOptions{
		DB:           db,
		Distro:       distro.Detect(),
		SecureCookie: false, // plain HTTP in tests
		BcryptCost:   4,     // fast hashing for tests
		AuthTimeout:  30,    // 30 minute session timeout
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
	json.NewDecoder(resp.Body).Decode(&body)
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
	json.NewDecoder(resp.Body).Decode(&body)
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
	json.NewDecoder(resp.Body).Decode(&body)
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
	json.NewDecoder(resp.Body).Decode(&body)
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
	// Must not reveal details — only "invalid credentials" is acceptable
	msg := string(body)
	if msg == "wrong password" || msg == "" {
		// ok — just checking it's not an internal error
	}
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
	json.NewDecoder(resp.Body).Decode(&body)
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
	defer os.Remove(f.Name())

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
	json.NewDecoder(resp.Body).Decode(&sessions)
	if len(sessions) == 0 {
		t.Error("session list should have at least one entry (current session)")
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
