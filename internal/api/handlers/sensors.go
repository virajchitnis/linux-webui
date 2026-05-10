package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/virajchitnis/linux-webui/internal/sensors"
)

// SensorsHandler serves hardware sensor data.
type SensorsHandler struct{}

func (h *SensorsHandler) Read(w http.ResponseWriter, r *http.Request) {
	chips, err := sensors.ReadAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chips)
}
