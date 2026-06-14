package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/virajchitnis/linux-webui/internal/distro"
	"github.com/virajchitnis/linux-webui/internal/metrics"
	"github.com/virajchitnis/linux-webui/internal/ollama"
)

// ── Capabilities ──────────────────────────────────────────────────────────────

func TestInitCapabilities(t *testing.T) {
	d := distro.Detect()
	caps := InitCapabilities(d)
	if caps == nil {
		t.Fatal("InitCapabilities returned nil")
	}
}

func TestCapabilities_SetOllama(t *testing.T) {
	caps := InitCapabilities(distro.Detect())
	caps.SetOllama(true)
	if !caps.Get("ollama") {
		t.Error("SetOllama(true) should make Get(ollama) return true")
	}
	caps.SetOllama(false)
	if caps.Get("ollama") {
		t.Error("SetOllama(false) should make Get(ollama) return false")
	}
}

func TestCapabilities_Get_Unknown(t *testing.T) {
	caps := InitCapabilities(distro.Detect())
	if caps.Get("nonexistent_key") {
		t.Error("Get on nonexistent key should return false")
	}
}

func TestCapabilities_Handler(t *testing.T) {
	caps := InitCapabilities(distro.Detect())
	h := caps.Handler()
	if h == nil {
		t.Fatal("Handler() returned nil")
	}
}

// ── aptPkgOK ─────────────────────────────────────────────────────────────────

func TestAptPkgOK(t *testing.T) {
	valid := []string{
		"curl", "vim", "python3.10", "libssl-dev",
		"apt-transport-https", "golang-1.21",
	}
	for _, name := range valid {
		if !aptPkgOK(name) {
			t.Errorf("aptPkgOK(%q) = false, want true", name)
		}
	}

	invalid := []string{
		"", "Uppercase", "has space", "has_under", "has@symbol",
		"verylongname" + string(make([]byte, 130)),
	}
	for _, name := range invalid {
		if aptPkgOK(name) {
			t.Errorf("aptPkgOK(%q) = true, want false", name)
		}
	}
}

func TestSentinelErr(t *testing.T) {
	if errBadPkg.Error() != "invalid package name" {
		t.Errorf("errBadPkg = %q, want 'invalid package name'", errBadPkg.Error())
	}
	if errUnknownAction.Error() != "unknown action" {
		t.Errorf("errUnknownAction = %q, want 'unknown action'", errUnknownAction.Error())
	}
}

// ── dockerSocketReadable / hwmonPresent ───────────────────────────────────────

func TestDockerSocketReadable(t *testing.T) {
	// Just verify it doesn't panic — result depends on the environment.
	_ = dockerSocketReadable()
}

func TestHwmonPresent(t *testing.T) {
	// Just verify it doesn't panic.
	_ = hwmonPresent()
}

// ── AIHandler ─────────────────────────────────────────────────────────────────

func TestAIHandler_BuildSystemPrompt_NoMetrics(t *testing.T) {
	h := &AIHandler{Metrics: nil}
	prompt := h.buildSystemPrompt()
	if prompt == "" {
		t.Error("buildSystemPrompt should return non-empty string")
	}
}

func TestAIHandler_BuildSystemPrompt_WithMetrics(t *testing.T) {
	collector := metrics.NewCollector()
	collector.Start()
	// Wait briefly for collector to produce a snapshot.
	time.Sleep(1100 * time.Millisecond)

	h := &AIHandler{Metrics: collector}
	prompt := h.buildSystemPrompt()
	if prompt == "" {
		t.Error("buildSystemPrompt with metrics should return non-empty string")
	}
}

func TestAIHandler_Models_Unavailable(t *testing.T) {
	client := ollama.NewClient("http://127.0.0.1:19998", "test")
	h := &AIHandler{Client: client}

	srv := httptest.NewServer(http.HandlerFunc(h.Models))
	defer srv.Close()

	resp, err := http.Get(srv.URL) //nolint
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("models with unreachable ollama: expected 503, got %d", resp.StatusCode)
	}
}

func TestAIHandler_Models_MockServer(t *testing.T) {
	mockOllama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"models":[{"name":"llama3.2"}]}`))
		}
	}))
	defer mockOllama.Close()

	client := ollama.NewClient(mockOllama.URL, "llama3.2")
	h := &AIHandler{Client: client}

	srv := httptest.NewServer(http.HandlerFunc(h.Models))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("models with mock server: expected 200, got %d", resp.StatusCode)
	}
}


// ── TerminalHandler.New ───────────────────────────────────────────────────────

func TestTerminalHandler_New(t *testing.T) {
	h := &TerminalHandler{}
	srv := httptest.NewServer(http.HandlerFunc(h.New))
	defer srv.Close()

	resp, err := http.Post(srv.URL, "application/json", nil)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("terminal new: expected 200, got %d", resp.StatusCode)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["id"] == "" {
		t.Error("terminal new should return non-empty id")
	}
}

// ── JournalWSHandler ─────────────────────────────────────────────────────────

func TestJournalWSHandler_ReadyMessage(t *testing.T) {
	h := &JournalWSHandler{
		AcceptOpts: &websocket.AcceptOptions{InsecureSkipVerify: true},
	}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.CloseNow()

	// Read messages until we get "ready" — means recent entries were sent.
	c.SetReadLimit(256 * 1024)
	for {
		_, data, err := c.Read(ctx)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		if strings.Contains(string(data), `"ready"`) {
			break
		}
	}
}

func TestJournalWSHandler_WithPriority(t *testing.T) {
	h := &JournalWSHandler{
		AcceptOpts: &websocket.AcceptOptions{InsecureSkipVerify: true},
	}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	// Connect with priority=3 (errors only)
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "?priority=3"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.CloseNow()

	c.SetReadLimit(256 * 1024)
	for {
		_, data, err := c.Read(ctx)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		if strings.Contains(string(data), `"ready"`) {
			break
		}
	}
}

// ── MetricsWSHandler ──────────────────────────────────────────────────────────

func TestMetricsWSHandler_ContextCancel(t *testing.T) {
	collector := metrics.NewCollector()
	collector.Start()

	h := &MetricsWSHandler{
		Collector:  collector,
		AcceptOpts: &websocket.AcceptOptions{InsecureSkipVerify: true},
	}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	// Wait for at least one message (collector broadcasts every second).
	c.SetReadLimit(65536)
	_, _, err = c.Read(ctx)
	c.CloseNow()
	if err != nil {
		t.Fatalf("read metrics ws message: %v", err)
	}
}
