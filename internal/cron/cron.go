// Package cron manages the linux-webui user's personal crontab.
package cron

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// Entry represents a parsed crontab entry.
type Entry struct {
	Schedule string `json:"schedule"` // "* * * * *"
	Command  string `json:"command"`
	Raw      string `json:"raw"`
	Index    int    `json:"index"`
}

// List runs `crontab -l` and returns all non-comment, non-empty entries.
func List() ([]Entry, error) {
	cmd := exec.Command("crontab", "-l")
	out, err := cmd.Output()
	if err != nil {
		// Exit 1 with "no crontab for user" is treated as empty
		if strings.Contains(string(out), "no crontab") {
			return []Entry{}, nil
		}
		if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) > 0 {
			msg := strings.TrimSpace(string(exitErr.Stderr))
			if strings.Contains(msg, "no crontab") {
				return []Entry{}, nil
			}
		}
		return nil, fmt.Errorf("crontab -l: %w", err)
	}
	return parseEntries(string(out)), nil
}

// Add appends a new cron entry to the crontab.
func Add(schedule, command string) error {
	if err := validateSchedule(schedule); err != nil {
		return err
	}
	if err := validateCommand(command); err != nil {
		return err
	}
	current, err := rawList()
	if err != nil {
		return err
	}
	newLine := schedule + " " + command + "\n"
	return writeCrontab(current + newLine)
}

// Delete removes the cron entry at the given index (0-based over non-comment lines).
func Delete(index int) error {
	raw, err := rawList()
	if err != nil {
		return err
	}
	entries := parseEntries(raw)
	if index < 0 || index >= len(entries) {
		return fmt.Errorf("index %d out of range", index)
	}
	// Rebuild without the target line
	sc := bufio.NewScanner(strings.NewReader(raw))
	var lines []string
	entryIdx := 0
	for sc.Scan() {
		line := sc.Text()
		if isEntry(line) {
			if entryIdx != index {
				lines = append(lines, line)
			}
			entryIdx++
		} else {
			lines = append(lines, line)
		}
	}
	return writeCrontab(strings.Join(lines, "\n") + "\n")
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func rawList() (string, error) {
	cmd := exec.Command("crontab", "-l")
	out, err := cmd.Output()
	if err != nil {
		if strings.Contains(string(out), "no crontab") {
			return "", nil
		}
		// Ignore "no crontab for user" errors
		return "", nil
	}
	return string(out), nil
}

func writeCrontab(content string) error {
	cmd := exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(content)
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("crontab write: %s", strings.TrimSpace(errBuf.String()))
	}
	return nil
}

func parseEntries(raw string) []Entry {
	var entries []Entry
	idx := 0
	sc := bufio.NewScanner(strings.NewReader(raw))
	for sc.Scan() {
		line := sc.Text()
		if !isEntry(line) {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		sched := strings.Join(fields[:5], " ")
		cmd := strings.Join(fields[5:], " ")
		entries = append(entries, Entry{Schedule: sched, Command: cmd, Raw: line, Index: idx})
		idx++
	}
	return entries
}

func isEntry(line string) bool {
	t := strings.TrimSpace(line)
	return t != "" && !strings.HasPrefix(t, "#") && !strings.HasPrefix(t, "@")
}

// reField validates a single schedule field.
var reField = regexp.MustCompile(`^(\*|[0-9*,/\-]+)$`)

func validateSchedule(s string) error {
	fields := strings.Fields(s)
	if len(fields) != 5 {
		return fmt.Errorf("schedule must have exactly 5 fields (got %d)", len(fields))
	}
	for _, f := range fields {
		if !reField.MatchString(f) {
			return fmt.Errorf("invalid schedule field: %q", f)
		}
	}
	return nil
}

func validateCommand(cmd string) error {
	if strings.ContainsAny(cmd, "\n\r") {
		return fmt.Errorf("command must not contain newlines")
	}
	if len(cmd) > 1000 {
		return fmt.Errorf("command too long")
	}
	if strings.TrimSpace(cmd) == "" {
		return fmt.Errorf("command must not be empty")
	}
	return nil
}
