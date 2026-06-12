package apt

import (
	"testing"
)

func TestTryLockUnlock(t *testing.T) {
	if !TryLock() {
		t.Fatal("TryLock() should succeed when lock is free")
	}
	if TryLock() {
		Unlock()
		t.Fatal("TryLock() should fail when already locked")
	}
	Unlock()
	if !TryLock() {
		t.Fatal("TryLock() should succeed after Unlock()")
	}
	Unlock()
}

func TestIsRunning(t *testing.T) {
	if IsRunning() {
		t.Fatal("IsRunning() should be false when no job is running")
	}
	if !TryLock() {
		t.Fatal("TryLock() failed unexpectedly")
	}
	if !IsRunning() {
		Unlock()
		t.Fatal("IsRunning() should be true while locked")
	}
	Unlock()
	if IsRunning() {
		t.Error("IsRunning() should be false after Unlock()")
	}
}

func TestUpgradableRegex(t *testing.T) {
	matchLines := []string{
		"bash/jammy-updates 5.1-6ubuntu1.1 amd64 [upgradable from: 5.1-6ubuntu1]",
		"vim/focal-updates 2:8.1.2269-1ubuntu5.9 amd64 [upgradable from: 2:8.1.2269-1ubuntu5.7]",
		"python3/jammy 3.10.6-1~22.04 amd64 [upgradable from: 3.10.4-1~22.04]",
	}
	for _, line := range matchLines {
		m := reUpgradable.FindStringSubmatch(line)
		if m == nil {
			t.Errorf("reUpgradable should match: %q", line)
			continue
		}
		if m[1] == "" {
			t.Errorf("package name should not be empty for: %q", line)
		}
		if m[2] == "" {
			t.Errorf("available version should not be empty for: %q", line)
		}
		if m[3] == "" {
			t.Errorf("arch should not be empty for: %q", line)
		}
		if m[4] == "" {
			t.Errorf("current version should not be empty for: %q", line)
		}
	}
}

func TestUpgradableRegex_NoMatch(t *testing.T) {
	noMatchLines := []string{
		"Listing... Done",
		"",
		"WARNING: apt does not have a stable CLI interface.",
		"bash 5.1-6ubuntu1.1 amd64",
	}
	for _, line := range noMatchLines {
		if m := reUpgradable.FindStringSubmatch(line); m != nil {
			t.Errorf("reUpgradable should not match: %q", line)
		}
	}
}

func TestUpgradableRegex_ParseFields(t *testing.T) {
	line := "curl/jammy-updates 7.81.0-1ubuntu1.14 amd64 [upgradable from: 7.81.0-1ubuntu1.13]"
	m := reUpgradable.FindStringSubmatch(line)
	if m == nil {
		t.Fatal("expected match")
	}
	if m[1] != "curl" {
		t.Errorf("name = %q, want curl", m[1])
	}
	if m[2] != "7.81.0-1ubuntu1.14" {
		t.Errorf("available = %q, want 7.81.0-1ubuntu1.14", m[2])
	}
	if m[3] != "amd64" {
		t.Errorf("arch = %q, want amd64", m[3])
	}
	if m[4] != "7.81.0-1ubuntu1.13" {
		t.Errorf("current = %q, want 7.81.0-1ubuntu1.13", m[4])
	}
}
