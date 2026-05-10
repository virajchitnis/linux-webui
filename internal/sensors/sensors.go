// Package sensors reads hardware sensor data from the kernel hwmon sysfs.
package sensors

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const hwmonBase = "/sys/class/hwmon"

// Chip represents one hwmon chip with all its sensor readings.
type Chip struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Temps   []Temp    `json:"temps,omitempty"`
	Fans    []Fan     `json:"fans,omitempty"`
}

// Temp is a temperature reading in °C.
type Temp struct {
	Label string  `json:"label"`
	Cur   float64 `json:"current_c"`
	Max   float64 `json:"max_c,omitempty"`
	Crit  float64 `json:"crit_c,omitempty"`
}

// Fan is a fan speed reading.
type Fan struct {
	Label string `json:"label"`
	RPM   int    `json:"rpm"`
}

// Available returns true if at least one hwmon chip is present.
func Available() bool {
	entries, _ := os.ReadDir(hwmonBase)
	return len(entries) > 0
}

// ReadAll returns sensor data for all hwmon chips.
func ReadAll() ([]Chip, error) {
	entries, err := os.ReadDir(hwmonBase)
	if err != nil {
		return nil, fmt.Errorf("hwmon: %w", err)
	}
	var chips []Chip
	for _, e := range entries {
		chipPath := filepath.Join(hwmonBase, e.Name())
		chip := readChip(chipPath)
		if len(chip.Temps) > 0 || len(chip.Fans) > 0 {
			chips = append(chips, chip)
		}
	}
	return chips, nil
}

func readChip(path string) Chip {
	name := strings.TrimSpace(readFile(filepath.Join(path, "name")))
	if name == "" {
		name = filepath.Base(path)
	}
	c := Chip{Name: name, Path: path}

	// Temperatures: temp{N}_input (millidegrees)
	for n := 1; n <= 32; n++ {
		input := filepath.Join(path, fmt.Sprintf("temp%d_input", n))
		val := readMilliDeg(input)
		if val == 0 && !fileExists(input) {
			break
		}
		label := readFile(filepath.Join(path, fmt.Sprintf("temp%d_label", n)))
		if label == "" {
			label = fmt.Sprintf("temp%d", n)
		}
		t := Temp{Label: strings.TrimSpace(label), Cur: val}
		t.Max = readMilliDeg(filepath.Join(path, fmt.Sprintf("temp%d_max", n)))
		t.Crit = readMilliDeg(filepath.Join(path, fmt.Sprintf("temp%d_crit", n)))
		c.Temps = append(c.Temps, t)
	}

	// Fans: fan{N}_input (RPM)
	for n := 1; n <= 16; n++ {
		input := filepath.Join(path, fmt.Sprintf("fan%d_input", n))
		if !fileExists(input) {
			break
		}
		rpm := readInt(input)
		label := readFile(filepath.Join(path, fmt.Sprintf("fan%d_label", n)))
		if label == "" {
			label = fmt.Sprintf("fan%d", n)
		}
		c.Fans = append(c.Fans, Fan{Label: strings.TrimSpace(label), RPM: rpm})
	}

	sort.Slice(c.Temps, func(i, j int) bool { return c.Temps[i].Label < c.Temps[j].Label })
	return c
}

func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func readMilliDeg(path string) float64 {
	s := readFile(path)
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v / 1000.0
}

func readInt(path string) int {
	s := readFile(path)
	v, _ := strconv.Atoi(s)
	return v
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
