package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/virajchitnis/linux-webui/internal/metrics"
	"github.com/virajchitnis/linux-webui/internal/ollama"
)

// AIHandler handles the /ws/ai WebSocket endpoint.
type AIHandler struct {
	Client     *ollama.Client
	Metrics    *metrics.Collector
	AcceptOpts *websocket.AcceptOptions
}

type aiIncoming struct {
	Type    string `json:"type"`
	Prompt  string `json:"prompt,omitempty"`
	Context string `json:"context,omitempty"` // optional extra context from UI
}

type aiOutgoing struct {
	Type    string `json:"type"`
	Token   string `json:"token,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (h *AIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, h.AcceptOpts)
	if err != nil {
		return
	}
	defer conn.CloseNow()

	ctx := conn.CloseRead(r.Context())

	for {
		var msg aiIncoming
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			return
		}
		if msg.Type != "generate" || msg.Prompt == "" {
			continue
		}

		system := h.buildSystemPrompt()
		if msg.Context != "" {
			system += "\n\nAdditional context from user:\n" + msg.Context
		}

		// Stream tokens back.
		tokens := make(chan string, 64)
		genCtx, cancel := context.WithCancel(ctx)
		go func() {
			defer cancel()
			defer close(tokens)
			if err := h.Client.Generate(genCtx, system, msg.Prompt, tokens); err != nil {
				log.Printf("ollama generate: %v", err)
			}
		}()

		for token := range tokens {
			if err := wsjson.Write(ctx, conn, aiOutgoing{Type: "token", Token: token}); err != nil {
				cancel()
				return
			}
		}
		cancel()

		// Send done signal.
		if err := wsjson.Write(ctx, conn, aiOutgoing{Type: "done"}); err != nil {
			return
		}
	}
}

func (h *AIHandler) buildSystemPrompt() string {
	sc := ollama.SystemContext{}

	if h.Metrics != nil {
		snap := h.Metrics.Latest()
		if snap != nil {
			sc.CPUPercent = snap.CPUPercent
			if snap.MemTotal > 0 {
				sc.MemPercent = float64(snap.MemUsed) / float64(snap.MemTotal) * 100
			}
		}
	}

	return ollama.BuildSystemPrompt(sc)
}

// AIModels returns the list of available Ollama models.
func (h *AIHandler) Models(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(h.Client.Endpoint() + "/api/tags")
	if err != nil {
		http.Error(w, "ollama unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	var v json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(v)
}
