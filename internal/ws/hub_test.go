package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func TestParseOriginHost(t *testing.T) {
	cases := []struct {
		origin string
		want   string
		hasErr bool
	}{
		{"https://localhost:8443", "localhost", false},
		{"http://127.0.0.1:5173", "127.0.0.1", false},
		{"https://example.com", "example.com", false},
		{"https://192.168.1.1:8443", "192.168.1.1", false},
		{"http://myhost", "myhost", false},
		// No scheme — returned as-is.
		{"localhost:8443", "localhost:8443", false},
	}
	for _, tc := range cases {
		got, err := parseOriginHost(tc.origin)
		if tc.hasErr && err == nil {
			t.Errorf("parseOriginHost(%q) expected error", tc.origin)
		}
		if !tc.hasErr && err != nil {
			t.Errorf("parseOriginHost(%q) unexpected error: %v", tc.origin, err)
		}
		if got != tc.want {
			t.Errorf("parseOriginHost(%q) = %q, want %q", tc.origin, got, tc.want)
		}
	}
}

func TestErrOriginForbidden(t *testing.T) {
	if ErrOriginForbidden.Error() == "" {
		t.Error("ErrOriginForbidden.Error() should be non-empty")
	}
}

func TestAcceptOpts(t *testing.T) {
	opts := AcceptOpts()
	if opts == nil {
		t.Fatal("AcceptOpts() returned nil")
	}
	if opts.InsecureSkipVerify {
		t.Error("AcceptOpts() should not set InsecureSkipVerify in production")
	}
}

// wsEcho creates a test server that accepts one WS message, echoes it back, then closes.
func wsEcho(t *testing.T, handler http.Handler) (*httptest.Server, string) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	return srv, wsURL
}

func TestWriteJSON_ReadEnvelope(t *testing.T) {
	_, wsURL := wsEcho(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			return
		}
		defer c.CloseNow()
		ctx := r.Context()

		// Send one message to the client.
		if err := WriteJSON(ctx, c, "test", map[string]string{"hello": "world"}); err != nil {
			t.Errorf("server WriteJSON: %v", err)
		}
		// Wait for client to close.
		_, _, _ = c.Read(ctx)
	}))

	ctx := context.Background()
	c, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.CloseNow()

	var env Envelope
	if err := wsjson.Read(ctx, c, &env); err != nil {
		t.Fatalf("client read: %v", err)
	}
	if env.Type != "test" {
		t.Errorf("envelope type = %q, want test", env.Type)
	}
	var payload map[string]string
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload["hello"] != "world" {
		t.Errorf("payload hello = %q, want world", payload["hello"])
	}
	c.Close(websocket.StatusNormalClosure, "")
}

func TestReadEnvelope_ServerReceives(t *testing.T) {
	received := make(chan *Envelope, 1)
	_, wsURL := wsEcho(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			return
		}
		defer c.CloseNow()
		env, err := ReadEnvelope(r.Context(), c)
		if err != nil {
			close(received)
			return
		}
		received <- env
	}))

	ctx := context.Background()
	c, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.CloseNow()

	// Send an envelope from client.
	msg := Envelope{Type: "ping", Payload: json.RawMessage(`{"x":1}`)}
	if err := wsjson.Write(ctx, c, msg); err != nil {
		t.Fatalf("client write: %v", err)
	}

	env := <-received
	if env == nil {
		t.Fatal("server did not receive envelope")
	}
	if env.Type != "ping" {
		t.Errorf("received type = %q, want ping", env.Type)
	}
	c.Close(websocket.StatusNormalClosure, "")
}

func TestCloseWithErr(t *testing.T) {
	_, wsURL := wsEcho(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			return
		}
		CloseWithErr(c, "test error")
	}))

	ctx := context.Background()
	c, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.CloseNow()

	// Reading should fail with the close error.
	_, _, err = c.Read(ctx)
	if err == nil {
		t.Error("expected error after CloseWithErr, got nil")
	}
}
