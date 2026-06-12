package exec

import (
	"testing"
)

func TestValidPkgName(t *testing.T) {
	valid := []string{
		"curl", "vim", "python3", "lib32-gcc-libs",
		"apt-transport-https", "ca-certificates",
		"golang-1.21", "lib.thing",
	}
	for _, name := range valid {
		if !ValidPkgName(name) {
			t.Errorf("ValidPkgName(%q) = false, want true", name)
		}
	}

	invalid := []string{
		"", " curl", "curl ", "Curl", "../etc",
		"curl;rm", "pkg name", "toolongnamethatexceeds128characterslimitXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
	}
	for _, name := range invalid {
		if ValidPkgName(name) {
			t.Errorf("ValidPkgName(%q) = true, want false", name)
		}
	}
}

func TestValidUsername(t *testing.T) {
	valid := []string{
		"root", "alice", "_service", "user1", "web-app",
	}
	for _, name := range valid {
		if !ValidUsername(name) {
			t.Errorf("ValidUsername(%q) = false, want true", name)
		}
	}

	invalid := []string{
		"", "1user", "-user", "HasUpper", "has space",
		"toolongnamethatexceedsthirtytwocharacters",
	}
	for _, name := range invalid {
		if ValidUsername(name) {
			t.Errorf("ValidUsername(%q) = true, want false", name)
		}
	}
}

func TestResolve_KnownBinary(t *testing.T) {
	// /bin/sh should exist on any Linux system.
	path, ok := Resolve("sh")
	if !ok {
		t.Fatal("Resolve(sh) returned false, expected sh to be on PATH")
	}
	if path == "" {
		t.Error("Resolve(sh) returned empty path")
	}
}

func TestResolve_Unknown(t *testing.T) {
	_, ok := Resolve("__nonexistent_binary_xyz__")
	if ok {
		t.Error("Resolve(nonexistent) should return false")
	}
}

func TestResolve_Cached(t *testing.T) {
	// Resolve twice — second call uses cache.
	p1, ok1 := Resolve("sh")
	p2, ok2 := Resolve("sh")
	if ok1 != ok2 || p1 != p2 {
		t.Error("cached Resolve should return same result")
	}
}

func TestCommand_AfterResolve(t *testing.T) {
	Resolve("sh") // ensure cached
	cmd, err := Command("sh", "-c", "true")
	if err != nil {
		t.Fatalf("Command(sh) error: %v", err)
	}
	if cmd == nil {
		t.Fatal("Command returned nil")
	}
}

func TestCommand_NotResolved(t *testing.T) {
	_, err := Command("__not_resolved_binary_xyz__")
	if err != ErrForbidden {
		t.Errorf("Command(not-resolved) should return ErrForbidden, got %v", err)
	}
}
