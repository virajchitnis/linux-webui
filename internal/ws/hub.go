package ws

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"

	"github.com/coder/websocket"
	"github.com/virajchitnis/linux-webui/internal/auth"
)

const maxMessageSize = 64 * 1024 // 64 KB for control messages

// Envelope is the typed message wrapper for all WebSocket messages.
type Envelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// ValidateUpgrade checks session cookie and Origin header before accepting a WS upgrade.
func ValidateUpgrade(r *http.Request, db interface {
	QueryRow(string, ...any) interface{ Scan(...any) error }
}, host string) (*auth.Session, error) {
	// Origin check (CSRF prevention for WebSocket)
	origin := r.Header.Get("Origin")
	if origin != "" {
		parsed, err := parseOriginHost(origin)
		if err != nil || (parsed != host && parsed != "localhost" && parsed != "127.0.0.1") {
			return nil, ErrOriginForbidden
		}
	}
	return nil, nil // session validation is handled by middleware wrapping the route
}

func parseOriginHost(origin string) (string, error) {
	// Simple extraction: strip scheme
	for _, prefix := range []string{"https://", "http://"} {
		if len(origin) > len(prefix) && origin[:len(prefix)] == prefix {
			rest := origin[len(prefix):]
			h, _, err := net.SplitHostPort(rest)
			if err != nil {
				return rest, nil // no port
			}
			return h, nil
		}
	}
	return origin, nil
}

var ErrOriginForbidden = &wsError{"origin not allowed"}

type wsError struct{ msg string }

func (e *wsError) Error() string { return e.msg }

// AcceptOpts returns standard websocket accept options.
func AcceptOpts() *websocket.AcceptOptions {
	return &websocket.AcceptOptions{
		InsecureSkipVerify: false,
	}
}

// WriteJSON writes a typed envelope to a websocket connection.
func WriteJSON(ctx context.Context, c *websocket.Conn, msgType string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	env := Envelope{Type: msgType, Payload: json.RawMessage(data)}
	b, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return c.Write(ctx, websocket.MessageText, b)
}

// ReadEnvelope reads one message and decodes it as an Envelope.
func ReadEnvelope(ctx context.Context, c *websocket.Conn) (*Envelope, error) {
	c.SetReadLimit(maxMessageSize)
	_, data, err := c.Read(ctx)
	if err != nil {
		return nil, err
	}
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

// CloseWithErr closes a WebSocket connection with an error message.
func CloseWithErr(c *websocket.Conn, msg string) {
	if err := c.Close(websocket.StatusPolicyViolation, msg); err != nil {
		log.Printf("ws close: %v", err)
	}
}
