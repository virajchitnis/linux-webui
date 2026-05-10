// Package netif reads network interface statistics from /proc and sysfs.
package netif

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Interface holds information and traffic stats for one network interface.
type Interface struct {
	Name      string   `json:"name"`
	State     string   `json:"state"`  // "up" | "down" | "unknown"
	Type      string   `json:"type"`   // "ethernet" | "wifi" | "loopback" | "virtual" | "other"
	Addresses []string `json:"addresses"`
	MTU       int      `json:"mtu"`
	RxBytes   uint64   `json:"rx_bytes"`
	TxBytes   uint64   `json:"tx_bytes"`
	RxPackets uint64   `json:"rx_packets"`
	TxPackets uint64   `json:"tx_packets"`
	MAC       string   `json:"mac"`
}

// List returns all network interfaces with their current statistics.
func List() ([]Interface, error) {
	traffic, err := readProcNetDev()
	if err != nil {
		return nil, err
	}

	goIfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	out := make([]Interface, 0, len(goIfaces))
	for _, gi := range goIfaces {
		iface := Interface{
			Name:  gi.Name,
			MTU:   gi.MTU,
			MAC:   gi.HardwareAddr.String(),
			State: "down",
			Type:  guessType(gi.Name),
		}
		if gi.Flags&net.FlagUp != 0 {
			iface.State = "up"
		}
		if gi.Flags&net.FlagLoopback != 0 {
			iface.Type = "loopback"
		}

		addrs, _ := gi.Addrs()
		for _, a := range addrs {
			iface.Addresses = append(iface.Addresses, a.String())
		}

		if t, ok := traffic[gi.Name]; ok {
			iface.RxBytes = t.rxBytes
			iface.TxBytes = t.txBytes
			iface.RxPackets = t.rxPackets
			iface.TxPackets = t.txPackets
		}

		// Read operstate from sysfs (more reliable than Flags).
		if state := readSysFS(gi.Name, "operstate"); state != "" {
			iface.State = strings.TrimSpace(state)
		}

		out = append(out, iface)
	}
	return out, nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

type trafficRow struct {
	rxBytes, rxPackets uint64
	txBytes, txPackets uint64
}

func readProcNetDev() (map[string]trafficRow, error) {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m := make(map[string]trafficRow)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		name := strings.TrimSpace(parts[0])
		fields := strings.Fields(parts[1])
		if len(fields) < 9 {
			continue
		}
		rx, _ := strconv.ParseUint(fields[0], 10, 64)
		rxp, _ := strconv.ParseUint(fields[1], 10, 64)
		tx, _ := strconv.ParseUint(fields[8], 10, 64)
		txp, _ := strconv.ParseUint(fields[9], 10, 64)
		m[name] = trafficRow{rxBytes: rx, rxPackets: rxp, txBytes: tx, txPackets: txp}
	}
	return m, sc.Err()
}

func readSysFS(iface, file string) string {
	data, err := os.ReadFile(filepath.Join("/sys/class/net", iface, file))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func guessType(name string) string {
	switch {
	case strings.HasPrefix(name, "lo"):
		return "loopback"
	case strings.HasPrefix(name, "eth") || strings.HasPrefix(name, "en") || strings.HasPrefix(name, "eno"):
		return "ethernet"
	case strings.HasPrefix(name, "wl") || strings.HasPrefix(name, "wlan") || strings.HasPrefix(name, "wifi"):
		return "wifi"
	case strings.HasPrefix(name, "docker") || strings.HasPrefix(name, "br-") || strings.HasPrefix(name, "virbr"):
		return "bridge"
	case strings.HasPrefix(name, "veth") || strings.HasPrefix(name, "tun") || strings.HasPrefix(name, "tap"):
		return "virtual"
	case strings.HasPrefix(name, "wg"):
		return "wireguard"
	default:
		return "other"
	}
}

// FmtBytes formats a byte count as a human-readable string.
func FmtBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
