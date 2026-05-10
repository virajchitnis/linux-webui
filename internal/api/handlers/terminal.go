package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/virajchitnis/linux-webui/internal/api/middleware"
	"github.com/virajchitnis/linux-webui/internal/auth"
	"github.com/virajchitnis/linux-webui/internal/terminal"
	appws "github.com/virajchitnis/linux-webui/internal/ws"
)

// TerminalHandler manages PTY sessions over WebSocket.
type TerminalHandler struct {
	Manager *terminal.Manager
	DB      *sql.DB
}

// New allocates a new terminal session ID.
func (h *TerminalHandler) New(w http.ResponseWriter, r *http.Request) {
	id := uuid.New().String()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

type termMsg struct {
	Data string `json:"data"`
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
}

// Connect handles WS /ws/terminal/{id}.
func (h *TerminalHandler) Connect(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	defer c.CloseNow()
	ctx := r.Context()

	sess, ok := h.Manager.Get(id)
	if !ok {
		sess, err = h.Manager.Create(id)
		if err != nil {
			_ = appws.WriteJSON(ctx, c, "error", map[string]string{"message": err.Error()})
			return
		}
		if s := middleware.SessionFromContext(ctx); s != nil {
			auth.LogAction(h.DB, s.UserID, s.Username, "terminal_open", id, middleware.ClientIP(r))
		}
	}

	// Stream PTY output → WebSocket.
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		_ = sess.ReadLoop(ctx, func(data []byte) {
			_ = appws.WriteJSON(ctx, c, "output", map[string]string{"data": string(data)})
		})
		_ = c.Close(websocket.StatusNormalClosure, "terminal closed")
	}()

	// WebSocket input → PTY.
	for {
		env, err := appws.ReadEnvelope(ctx, c)
		if err != nil {
			<-readDone
			return
		}
		var msg termMsg
		if err := json.Unmarshal(env.Payload, &msg); err != nil {
			continue
		}
		switch env.Type {
		case "input":
			_ = sess.Write([]byte(msg.Data))
		case "resize":
			if msg.Rows > 0 && msg.Cols > 0 {
				_ = sess.Resize(msg.Rows, msg.Cols)
			}
		}
	}
}
