package journal

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Entry is a parsed journalctl log entry.
type Entry struct {
	Timestamp string `json:"timestamp"`
	Unit      string `json:"unit"`
	Priority  int    `json:"priority"`
	Message   string `json:"message"`
	PID       string `json:"pid"`
}

// StreamRecent sends up to n recent log entries (not following) then returns.
func StreamRecent(ctx context.Context, unit string, priority int, n int, ch chan<- Entry) error {
	args := buildArgs(false, unit, priority, n)
	cmd := exec.CommandContext(ctx, "journalctl", args...)
	out, err := cmd.Output()
	if err != nil && len(out) == 0 {
		return fmt.Errorf("journalctl: %w", err)
	}
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	sc.Buffer(make([]byte, 256*1024), 256*1024)
	for sc.Scan() {
		e, err := parseEntry(sc.Text())
		if err != nil {
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- e:
		}
	}
	return nil
}

// Stream follows the journal and sends entries to ch until ctx is cancelled.
func Stream(ctx context.Context, unit string, priority int, ch chan<- Entry) error {
	args := buildArgs(true, unit, priority, 0)
	cmd := exec.CommandContext(ctx, "journalctl", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	sc := bufio.NewScanner(stdout)
	sc.Buffer(make([]byte, 256*1024), 256*1024)
	for sc.Scan() {
		e, err := parseEntry(sc.Text())
		if err != nil {
			continue
		}
		select {
		case <-ctx.Done():
			_ = cmd.Process.Kill()
			return ctx.Err()
		case ch <- e:
		}
	}
	return cmd.Wait()
}

func buildArgs(follow bool, unit string, priority, n int) []string {
	args := []string{"--output=json", "--no-pager"}
	if follow {
		args = append(args, "-f")
	} else if n > 0 {
		args = append(args, fmt.Sprintf("-n%d", n))
	}
	if unit != "" {
		args = append(args, "-u", unit)
	}
	if priority >= 0 {
		args = append(args, fmt.Sprintf("-p%d", priority))
	}
	return args
}

type rawEntry struct {
	RealtimeTimestamp string      `json:"__REALTIME_TIMESTAMP"`
	SystemdUnit       string      `json:"_SYSTEMD_UNIT"`
	SyslogIdentifier  string      `json:"SYSLOG_IDENTIFIER"`
	Priority          string      `json:"PRIORITY"`
	Message           interface{} `json:"MESSAGE"` // string or []interface{} of byte values
	PID               string      `json:"_PID"`
}

func parseEntry(line string) (Entry, error) {
	var raw rawEntry
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return Entry{}, err
	}
	unit := raw.SystemdUnit
	if unit == "" {
		unit = raw.SyslogIdentifier
	}
	var ts string
	if raw.RealtimeTimestamp != "" {
		var usec int64
		fmt.Sscanf(raw.RealtimeTimestamp, "%d", &usec)
		t := time.Unix(usec/1_000_000, (usec%1_000_000)*1000).UTC()
		ts = t.Format(time.RFC3339)
	}
	prio := 7
	fmt.Sscanf(raw.Priority, "%d", &prio)

	var msg string
	switch v := raw.Message.(type) {
	case string:
		msg = v
	case []interface{}:
		bs := make([]byte, len(v))
		for i, b := range v {
			if n, ok := b.(float64); ok {
				bs[i] = byte(n)
			}
		}
		msg = string(bs)
	}

	return Entry{Timestamp: ts, Unit: unit, Priority: prio, Message: msg, PID: raw.PID}, nil
}
