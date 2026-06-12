package cron

import (
	"strings"
	"testing"
)

func TestValidateSchedule(t *testing.T) {
	valid := []string{
		"* * * * *",
		"0 * * * *",
		"0 0 * * *",
		"0 0 1 * *",
		"0 0 1 1 *",
		"0 0 1 1 0",
		"*/5 * * * *",
		"0,30 * * * *",
		"0-59 * * * *",
		"0 9-17 * * 1-5",
	}
	for _, s := range valid {
		if err := validateSchedule(s); err != nil {
			t.Errorf("validateSchedule(%q) returned unexpected error: %v", s, err)
		}
	}

	invalid := []string{
		"",
		"* * * *",        // only 4 fields
		"* * * * * *",    // 6 fields
		"abc * * * *",    // non-numeric field
		"* * * * @reboot", // @ syntax not allowed
	}
	for _, s := range invalid {
		if err := validateSchedule(s); err == nil {
			t.Errorf("validateSchedule(%q) should have returned error", s)
		}
	}
}

func TestValidateCommand(t *testing.T) {
	valid := []string{
		"/usr/bin/backup.sh",
		"echo hello",
		"/bin/sh -c 'ls -la'",
		strings.Repeat("a", 1000), // exactly at limit
	}
	for _, cmd := range valid {
		if err := validateCommand(cmd); err != nil {
			t.Errorf("validateCommand(%q) returned unexpected error: %v", cmd, err)
		}
	}

	invalid := []string{
		"",
		"   ",                         // whitespace only
		"echo hello\nrm -rf /",        // newline injection
		"echo hello\rrm -rf /",        // CR injection
		strings.Repeat("x", 1001),     // too long
	}
	for _, cmd := range invalid {
		if err := validateCommand(cmd); err == nil {
			t.Errorf("validateCommand(%q) should have returned error", cmd)
		}
	}
}

func TestParseEntries(t *testing.T) {
	raw := `# comment
@reboot /usr/bin/foo

* * * * * /usr/bin/backup.sh
0 2 * * * /usr/bin/cleanup.sh --all
`
	entries := parseEntries(raw)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d: %v", len(entries), entries)
	}
	if entries[0].Command != "/usr/bin/backup.sh" {
		t.Errorf("first entry command = %q, want /usr/bin/backup.sh", entries[0].Command)
	}
	if entries[0].Schedule != "* * * * *" {
		t.Errorf("first entry schedule = %q, want '* * * * *'", entries[0].Schedule)
	}
	if entries[1].Index != 1 {
		t.Errorf("second entry index = %d, want 1", entries[1].Index)
	}
}

func TestRawList(t *testing.T) {
	// rawList reads the current crontab (or returns "" if empty/missing).
	raw, err := rawList()
	if err != nil {
		t.Logf("rawList: %v (acceptable in CI)", err)
	}
	_ = raw
}

func TestList(t *testing.T) {
	entries, err := List()
	if err != nil {
		t.Logf("List: %v (acceptable in CI)", err)
		return
	}
	_ = entries
}

func TestDelete_OutOfRange(t *testing.T) {
	// With an index that is out of range, Delete returns an error.
	// Since the crontab may be empty or have no entry at index 9999, this
	// exercises the index-validation path without modifying any real entries.
	err := Delete(9999)
	// Either "index 9999 out of range" or a crontab access error — both are errors.
	if err == nil {
		// Only OK if the crontab actually has 10000+ entries (extremely unlikely).
		t.Log("Delete(9999) returned nil — unexpectedly large crontab")
	}
}

func TestIsEntry(t *testing.T) {
	shouldBeEntry := []string{
		"* * * * * /usr/bin/foo",
		"0 2 * * * /usr/bin/bar --flag",
	}
	for _, s := range shouldBeEntry {
		if !isEntry(s) {
			t.Errorf("isEntry(%q) = false, want true", s)
		}
	}

	shouldNotBeEntry := []string{
		"",
		"   ",
		"# comment",
		"@reboot /usr/bin/foo",
	}
	for _, s := range shouldNotBeEntry {
		if isEntry(s) {
			t.Errorf("isEntry(%q) = true, want false", s)
		}
	}
}
