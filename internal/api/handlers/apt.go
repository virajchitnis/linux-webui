package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/coder/websocket"
	"github.com/virajchitnis/linux-webui/internal/api/middleware"
	"github.com/virajchitnis/linux-webui/internal/apt"
	"github.com/virajchitnis/linux-webui/internal/auth"
	appws "github.com/virajchitnis/linux-webui/internal/ws"
)

// AptHandler handles package-manager REST and WebSocket endpoints.
type AptHandler struct {
	DB *sql.DB
}

func (h *AptHandler) Upgradable(w http.ResponseWriter, r *http.Request) {
	pkgs, err := apt.ListUpgradable()
	if err != nil {
		http.Error(w, "apt list: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if pkgs == nil {
		pkgs = []apt.Package{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pkgs)
}

func (h *AptHandler) Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"running": apt.IsRunning()})
}

type aptRunPayload struct {
	Action  string `json:"action"`  // "update", "upgrade", "install"
	Package string `json:"package"` // only for "install"
}

// Stream handles WS /ws/apt — client sends a run command, server streams apt-get output.
func (h *AptHandler) Stream(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	defer c.CloseNow()
	ctx := r.Context()

	env, err := appws.ReadEnvelope(ctx, c)
	if err != nil {
		return
	}
	if env.Type != "run" {
		appws.CloseWithErr(c, "expected run message")
		return
	}
	var payload aptRunPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		appws.CloseWithErr(c, "bad payload")
		return
	}

	if !apt.TryLock() {
		_ = appws.WriteJSON(ctx, c, "error", map[string]string{"message": "an apt job is already running"})
		return
	}
	defer apt.Unlock()

	session := middleware.SessionFromContext(ctx)
	if session != nil {
		auth.LogAction(h.DB, session.UserID, session.Username,
			"apt_"+payload.Action, payload.Package, middleware.ClientIP(r))
	}

	lineCh := make(chan string, 256)
	var runErr error
	go func() {
		defer close(lineCh)
		switch payload.Action {
		case "update":
			runErr = apt.RunUpdate(ctx, lineCh)
		case "upgrade":
			runErr = apt.RunUpgrade(ctx, lineCh)
		case "install":
			if !aptPkgOK(payload.Package) {
				runErr = errBadPkg
				return
			}
			runErr = apt.RunInstall(ctx, payload.Package, lineCh)
		default:
			runErr = errUnknownAction
		}
	}()

	for line := range lineCh {
		if err := appws.WriteJSON(ctx, c, "output", map[string]string{"line": line}); err != nil {
			return
		}
	}

	exitCode := 0
	errMsg := ""
	if runErr != nil {
		exitCode = 1
		errMsg = runErr.Error()
	}
	_ = appws.WriteJSON(ctx, c, "done", map[string]interface{}{
		"exit_code": exitCode,
		"error":     errMsg,
	})
}

func aptPkgOK(name string) bool {
	if len(name) == 0 || len(name) > 128 {
		return false
	}
	for _, c := range name {
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '.' && c != '+' && c != '-' {
			return false
		}
	}
	return true
}

type sentinelErr string

func (e sentinelErr) Error() string { return string(e) }

const (
	errBadPkg        sentinelErr = "invalid package name"
	errUnknownAction sentinelErr = "unknown action"
)
