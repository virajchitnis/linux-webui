package handlers

import (
	"net/http"
	"strconv"

	"github.com/coder/websocket"
	"github.com/virajchitnis/linux-webui/internal/journal"
	appws "github.com/virajchitnis/linux-webui/internal/ws"
)

// JournalWSHandler streams journalctl output to WebSocket clients.
type JournalWSHandler struct{}

func (h *JournalWSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	defer c.CloseNow()
	ctx := r.Context()

	unit := r.URL.Query().Get("unit")
	priority := -1
	if p := r.URL.Query().Get("priority"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n >= 0 && n <= 7 {
			priority = n
		}
	}

	// Send recent entries first (last 200 lines).
	recent := make(chan journal.Entry, 256)
	go func() {
		defer close(recent)
		_ = journal.StreamRecent(ctx, unit, priority, 200, recent)
	}()
	for e := range recent {
		if err := appws.WriteJSON(ctx, c, "log", e); err != nil {
			return
		}
	}

	// Signal to frontend that initial batch is done.
	if err := appws.WriteJSON(ctx, c, "ready", nil); err != nil {
		return
	}

	// Follow live journal output.
	live := make(chan journal.Entry, 256)
	go func() {
		defer close(live)
		_ = journal.Stream(ctx, unit, priority, live)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-live:
			if !ok {
				return
			}
			if err := appws.WriteJSON(ctx, c, "log", e); err != nil {
				return
			}
		}
	}
}
