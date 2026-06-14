package journal

import (
	"context"
	"testing"
)

func TestBuildArgs_Follow(t *testing.T) {
	args := buildArgs(true, "", -1, 0)
	has := func(s string) bool {
		for _, a := range args {
			if a == s {
				return true
			}
		}
		return false
	}
	if !has("-f") {
		t.Error("follow=true should add -f")
	}
	if !has("--output=json") {
		t.Error("should always include --output=json")
	}
	if !has("--no-pager") {
		t.Error("should always include --no-pager")
	}
}

func TestBuildArgs_NoFollow_N(t *testing.T) {
	args := buildArgs(false, "", -1, 50)
	has := func(s string) bool {
		for _, a := range args {
			if a == s {
				return true
			}
		}
		return false
	}
	if !has("-n50") {
		t.Error("n=50 should add -n50")
	}
	if has("-f") {
		t.Error("follow=false should not add -f")
	}
}

func TestBuildArgs_Unit(t *testing.T) {
	args := buildArgs(false, "nginx.service", -1, 0)
	found := false
	for i, a := range args {
		if a == "-u" && i+1 < len(args) && args[i+1] == "nginx.service" {
			found = true
		}
	}
	if !found {
		t.Error("unit=nginx.service should add -u nginx.service")
	}
}

func TestBuildArgs_Priority(t *testing.T) {
	args := buildArgs(false, "", 3, 0)
	has := func(s string) bool {
		for _, a := range args {
			if a == s {
				return true
			}
		}
		return false
	}
	if !has("-p3") {
		t.Error("priority=3 should add -p3")
	}
}

func TestBuildArgs_NegativePriority(t *testing.T) {
	args := buildArgs(false, "", -1, 0)
	for _, a := range args {
		if len(a) > 2 && a[:2] == "-p" {
			t.Errorf("negative priority should not add -p flag, got %q", a)
		}
	}
}

func TestParseEntry_StringMessage(t *testing.T) {
	line := `{"__REALTIME_TIMESTAMP":"1718200000000000","_SYSTEMD_UNIT":"nginx.service","PRIORITY":"3","MESSAGE":"hello world","_PID":"1234"}`
	e, err := parseEntry(line)
	if err != nil {
		t.Fatalf("parseEntry error: %v", err)
	}
	if e.Unit != "nginx.service" {
		t.Errorf("unit = %q, want nginx.service", e.Unit)
	}
	if e.Message != "hello world" {
		t.Errorf("message = %q, want hello world", e.Message)
	}
	if e.Priority != 3 {
		t.Errorf("priority = %d, want 3", e.Priority)
	}
	if e.PID != "1234" {
		t.Errorf("pid = %q, want 1234", e.PID)
	}
	if e.Timestamp == "" {
		t.Error("timestamp should be set")
	}
}

func TestParseEntry_ByteArrayMessage(t *testing.T) {
	// MESSAGE as byte array: [72, 105] = "Hi"
	line := `{"__REALTIME_TIMESTAMP":"1718200000000000","_SYSTEMD_UNIT":"test.service","PRIORITY":"6","MESSAGE":[72,105],"_PID":"99"}`
	e, err := parseEntry(line)
	if err != nil {
		t.Fatalf("parseEntry error: %v", err)
	}
	if e.Message != "Hi" {
		t.Errorf("byte-array message = %q, want Hi", e.Message)
	}
}

func TestParseEntry_FallbackToSyslogIdentifier(t *testing.T) {
	line := `{"SYSLOG_IDENTIFIER":"myapp","PRIORITY":"7","MESSAGE":"test"}`
	e, err := parseEntry(line)
	if err != nil {
		t.Fatalf("parseEntry error: %v", err)
	}
	if e.Unit != "myapp" {
		t.Errorf("unit = %q, want myapp", e.Unit)
	}
}

func TestParseEntry_DefaultPriority(t *testing.T) {
	// No priority field — should default to 7
	line := `{"MESSAGE":"no prio"}`
	e, err := parseEntry(line)
	if err != nil {
		t.Fatalf("parseEntry error: %v", err)
	}
	if e.Priority != 7 {
		t.Errorf("priority = %d, want 7 (default)", e.Priority)
	}
}

func TestParseEntry_InvalidJSON(t *testing.T) {
	_, err := parseEntry("this is not json")
	if err == nil {
		t.Error("parseEntry(invalid JSON) should return error")
	}
}

func TestParseEntry_EmptyTimestamp(t *testing.T) {
	line := `{"MESSAGE":"no ts"}`
	e, err := parseEntry(line)
	if err != nil {
		t.Fatalf("parseEntry error: %v", err)
	}
	if e.Timestamp != "" {
		t.Errorf("empty timestamp field should produce empty Timestamp, got %q", e.Timestamp)
	}
}

func TestStreamRecent_NoUnit(t *testing.T) {
	ctx := t.Context()
	ch := make(chan Entry, 100)
	// Request recent kernel entries — journalctl is available in this env.
	err := StreamRecent(ctx, "", 7, 5, ch)
	// We don't assert specific entries, just that it doesn't panic.
	if err != nil {
		t.Logf("StreamRecent: %v (may be normal in CI)", err)
	}
}

func TestStreamRecent_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	ch := make(chan Entry, 1)
	// With a cancelled context, the function should return promptly.
	_ = StreamRecent(ctx, "", 7, 100, ch)
}
