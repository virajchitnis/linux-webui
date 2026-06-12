package ollama

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClient_Defaults(t *testing.T) {
	c := NewClient("", "")
	if c.endpoint != defaultEndpoint {
		t.Errorf("default endpoint = %q, want %q", c.endpoint, defaultEndpoint)
	}
	if c.model != "llama3.2" {
		t.Errorf("default model = %q, want llama3.2", c.model)
	}
}

func TestNewClient_Custom(t *testing.T) {
	c := NewClient("http://localhost:11434/", "mistral")
	if c.endpoint != "http://localhost:11434" {
		t.Errorf("endpoint = %q (should strip trailing slash)", c.endpoint)
	}
	if c.model != "mistral" {
		t.Errorf("model = %q, want mistral", c.model)
	}
}

func TestEndpoint(t *testing.T) {
	c := NewClient("http://myhost:11434", "llama3")
	if c.Endpoint() != "http://myhost:11434" {
		t.Errorf("Endpoint() = %q, want http://myhost:11434", c.Endpoint())
	}
}

func TestAvailable_Unreachable(t *testing.T) {
	if Available("http://127.0.0.1:19999") {
		t.Error("Available on unreachable server should return false")
	}
}

func TestAvailable_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	if !Available(srv.URL) {
		t.Error("Available on reachable server should return true")
	}
}

func TestAvailable_WrongStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	if Available(srv.URL) {
		t.Error("Available with non-200 response should return false")
	}
}

func TestBuildSystemPrompt_Basic(t *testing.T) {
	sc := SystemContext{
		Hostname:   "myserver",
		Distro:     "Ubuntu 24.04",
		CPUPercent: 42.5,
		MemPercent: 67.8,
	}
	prompt := BuildSystemPrompt(sc)

	for _, want := range []string{"myserver", "Ubuntu 24.04", "42.5%", "67.8%"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt should contain %q", want)
		}
	}
}

func TestBuildSystemPrompt_WithServices(t *testing.T) {
	sc := SystemContext{
		Hostname:    "server",
		Distro:      "Debian",
		TopServices: []string{"nginx.service", "postgresql.service"},
	}
	prompt := BuildSystemPrompt(sc)
	if !strings.Contains(prompt, "nginx.service") {
		t.Error("prompt should contain service names")
	}
}

func TestBuildSystemPrompt_WithLogs(t *testing.T) {
	sc := SystemContext{
		Hostname:   "server",
		Distro:     "Ubuntu",
		RecentLogs: []string{"error: disk full", "OOM killer invoked"},
	}
	prompt := BuildSystemPrompt(sc)
	if !strings.Contains(prompt, "<<<LOG_START>>>") {
		t.Error("prompt should contain LOG_START marker")
	}
	if !strings.Contains(prompt, "error: disk full") {
		t.Error("prompt should contain log lines")
	}
}

func TestBuildSystemPrompt_LongLogLine(t *testing.T) {
	longLine := strings.Repeat("x", 600)
	sc := SystemContext{
		Hostname:   "server",
		Distro:     "Ubuntu",
		RecentLogs: []string{longLine},
	}
	prompt := BuildSystemPrompt(sc)
	if !strings.Contains(prompt, "…") {
		t.Error("long log lines should be truncated with ellipsis")
	}
}

func TestBuildSystemPrompt_NoLogs(t *testing.T) {
	sc := SystemContext{Hostname: "server", Distro: "Ubuntu"}
	prompt := BuildSystemPrompt(sc)
	// The rules section mentions <<<LOG_START>>> regardless; but "Recent system log errors"
	// should only appear when there are actual logs.
	if strings.Contains(prompt, "Recent system log errors") {
		t.Error("prompt without logs should not contain 'Recent system log errors'")
	}
}

func TestGenerate_MockServer(t *testing.T) {
	// Serve a streaming NDJSON Ollama response.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusOK)
		f := w.(http.Flusher)
		chunks := []string{
			`{"model":"test","response":"Hello","done":false}`,
			`{"model":"test","response":" world","done":false}`,
			`{"model":"test","response":"","done":true}`,
		}
		for _, line := range chunks {
			_, _ = w.Write([]byte(line + "\n"))
			f.Flush()
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test")
	tokens := make(chan string, 10)
	ctx := context.Background()
	err := c.Generate(ctx, "system", "user prompt", tokens)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	close(tokens)

	var got []string
	for tok := range tokens {
		got = append(got, tok)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 tokens, got %d: %v", len(got), got)
	}
	if got[0] != "Hello" || got[1] != " world" {
		t.Errorf("tokens = %v, want [Hello  world]", got)
	}
}

func TestGenerate_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test")
	tokens := make(chan string, 1)
	err := c.Generate(context.Background(), "", "prompt", tokens)
	if err == nil {
		t.Error("Generate with 500 response should return error")
	}
}

func TestGenerate_ContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusOK)
		// Block forever — context cancellation should abort.
		select {
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	c := NewClient(srv.URL, "test")
	tokens := make(chan string, 1)
	done := make(chan error, 1)
	go func() {
		done <- c.Generate(ctx, "", "prompt", tokens)
	}()
	cancel()
	err := <-done
	if err == nil {
		t.Error("Generate should return error when context is cancelled")
	}
}
