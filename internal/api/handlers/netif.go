package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/virajchitnis/linux-webui/internal/netif"
)

// NetifHandler serves network interface statistics.
type NetifHandler struct{}

func (h *NetifHandler) List(w http.ResponseWriter, r *http.Request) {
	ifaces, err := netif.List()
	if err != nil {
		log.Printf("netif list: %v", err)
		http.Error(w, "failed to read network interfaces", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ifaces)
}
