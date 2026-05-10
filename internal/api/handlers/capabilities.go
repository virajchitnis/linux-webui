package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/virajchitnis/linux-webui/internal/distro"
	safeexec "github.com/virajchitnis/linux-webui/internal/exec"
)

type Capabilities struct {
	mu     sync.RWMutex
	data   map[string]any
	distro *distro.Distro
}

func InitCapabilities(d *distro.Distro) *Capabilities {
	c := &Capabilities{distro: d, data: make(map[string]any)}
	c.probe()
	go func() {
		for range time.Tick(5 * time.Minute) {
			c.probe()
		}
	}()
	return c
}

func (c *Capabilities) probe() {
	_, firewallUFW := safeexec.Resolve("ufw")
	_, firewallFD := safeexec.Resolve("firewall-cmd")
	_, hasApt := safeexec.Resolve("apt-get")
	_, hasDNF := safeexec.Resolve("dnf")
	_, hasYUM := safeexec.Resolve("yum")
	_, hasPacman := safeexec.Resolve("pacman")
	_, hasBash := safeexec.Resolve("bash")
	_, hasCron := safeexec.Resolve("crontab")
	_, hasJournal := safeexec.Resolve("journalctl")
	_, hasUseradd := safeexec.Resolve("useradd")
	_, hasWG := safeexec.Resolve("wg")
	_, hasTS := safeexec.Resolve("tailscale")

	caps := map[string]any{
		"distro": map[string]any{
			"id":      c.distro.ID,
			"like":    c.distro.Like,
			"version": c.distro.Version,
			"family":  c.distro.FamilyString(),
		},
		"firewall_ufw":       firewallUFW,
		"firewall_firewalld": firewallFD,
		"apt":                hasApt,
		"dnf":                hasDNF,
		"yum":                hasYUM,
		"pacman":             hasPacman,
		"terminal":           hasBash,
		"cron":               hasCron,
		"journalctl":         hasJournal,
		"useradd":            hasUseradd,
		"wireguard":          hasWG,
		"tailscale":          hasTS,
		"docker":             dockerSocketReadable(),
		"sensors":            hwmonPresent(),
		"ollama":             false,
	}

	c.mu.Lock()
	c.data = caps
	c.mu.Unlock()
}

func (c *Capabilities) SetOllama(ok bool) {
	c.mu.Lock()
	c.data["ollama"] = ok
	c.mu.Unlock()
}

func (c *Capabilities) Get(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, _ := c.data[key]
	b, _ := v.(bool)
	return b
}

func dockerSocketReadable() bool {
	f, err := os.Open("/var/run/docker.sock")
	if err != nil {
		return false
	}
	f.Close()
	return true
}

func hwmonPresent() bool {
	entries, err := os.ReadDir("/sys/class/hwmon")
	return err == nil && len(entries) > 0
}

func (c *Capabilities) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.mu.RLock()
		data := c.data
		c.mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}
