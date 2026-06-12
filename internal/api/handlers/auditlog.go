package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/virajchitnis/linux-webui/internal/auth"
)

// AuditLogHandler serves the audit log to admin users.
type AuditLogHandler struct {
	DB *sql.DB
}

func (h *AuditLogHandler) List(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 200
	if limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 && n <= 1000 {
			limit = n
		}
	}
	entries, err := auth.GetAuditLog(h.DB, limit)
	if err != nil {
		log.Printf("audit log query: %v", err)
		http.Error(w, "failed to read audit log", http.StatusInternalServerError)
		return
	}
	if entries == nil {
		entries = []map[string]any{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}
