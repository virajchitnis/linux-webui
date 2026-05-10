package handlers

import (
	"net/http"

	"github.com/coder/websocket"
	"github.com/virajchitnis/linux-webui/internal/metrics"
	appws "github.com/virajchitnis/linux-webui/internal/ws"
)

type MetricsWSHandler struct {
	Collector *metrics.Collector
}

func (h *MetricsWSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // Origin validated by middleware
	})
	if err != nil {
		return
	}
	defer c.CloseNow()

	ctx := r.Context()
	ch := h.Collector.Subscribe()
	defer h.Collector.Unsubscribe(ch)

	for {
		select {
		case <-ctx.Done():
			return
		case snap, ok := <-ch:
			if !ok {
				return
			}
			if err := appws.WriteJSON(ctx, c, "metrics", snap); err != nil {
				return
			}
		}
	}
}
