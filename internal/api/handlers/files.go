package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/virajchitnis/linux-webui/internal/files"
)

// FilesHandler serves the read-only file browser.
type FilesHandler struct {
	AllowedRoots files.Roots
}

func (h *FilesHandler) List(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}
	entries, err := files.ListDir(path, h.AllowedRoots)
	if err != nil {
		if errors.Is(err, files.ErrEscape) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		log.Printf("files list %q: %v", path, err)
		http.Error(w, "failed to list directory", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func (h *FilesHandler) Read(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path required", http.StatusBadRequest)
		return
	}
	data, err := files.ReadFile(path, h.AllowedRoots)
	if err != nil {
		switch {
		case errors.Is(err, files.ErrEscape):
			http.Error(w, "forbidden", http.StatusForbidden)
		case errors.Is(err, files.ErrTooLarge):
			http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
		default:
			log.Printf("files read %q: %v", path, err)
			http.Error(w, "failed to read file", http.StatusInternalServerError)
		}
		return
	}
	// Detect content type; force text for unknown binary sniff.
	ct := http.DetectContentType(data)
	if strings.HasPrefix(ct, "application/octet-stream") {
		ct = "text/plain; charset=utf-8"
	}
	w.Header().Set("Content-Type", ct)
	w.Write(data)
}

func (h *FilesHandler) GetRoots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.AllowedRoots)
}
