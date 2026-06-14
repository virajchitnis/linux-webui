package process

import (
	"os"
	"syscall"
	"testing"
)

func TestList(t *testing.T) {
	procs, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if len(procs) == 0 {
		t.Error("expected at least one process")
	}
	var foundSelf bool
	self := os.Getpid()
	for _, p := range procs {
		if p.PID <= 0 {
			t.Errorf("invalid PID %d", p.PID)
		}
		if p.PID == self {
			foundSelf = true
			if p.Name == "" {
				t.Error("self process should have a name")
			}
		}
	}
	if !foundSelf {
		t.Errorf("own PID %d not found in process list", self)
	}
}

func TestList_SortedByPID(t *testing.T) {
	procs, err := List()
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(procs); i++ {
		if procs[i].PID < procs[i-1].PID {
			t.Errorf("processes not sorted: procs[%d].PID=%d < procs[%d].PID=%d",
				i, procs[i].PID, i-1, procs[i-1].PID)
		}
	}
}

func TestKill_RefuseLowPIDs(t *testing.T) {
	for _, pid := range []int{0, 1, -1} {
		err := Kill(pid, syscall.SIGTERM)
		if err == nil {
			t.Errorf("Kill(%d) should return error", pid)
		}
	}
}

func TestRenice_InvalidPriority(t *testing.T) {
	pid := os.Getpid()
	for _, pri := range []int{-21, 20, 100, -100} {
		if err := Renice(pid, pri); err == nil {
			t.Errorf("Renice(%d) should be rejected", pri)
		}
	}
}

func TestRenice_ValidBoundary(t *testing.T) {
	pid := os.Getpid()
	// Setting own nice to 0 is always permitted.
	if err := Renice(pid, 0); err != nil {
		t.Errorf("Renice(self, 0) unexpected error: %v", err)
	}
}

func TestSystemParams(t *testing.T) {
	uptime, hertz, numCPU := systemParams()
	if uptime <= 0 {
		t.Errorf("uptime should be positive, got %f", uptime)
	}
	if hertz <= 0 {
		t.Errorf("hertz should be positive, got %f", hertz)
	}
	if numCPU <= 0 {
		t.Errorf("numCPU should be positive, got %d", numCPU)
	}
}

func TestBuildUIDMap(t *testing.T) {
	m := buildUIDMap()
	if m == nil {
		t.Fatal("buildUIDMap returned nil")
	}
	// UID 0 must be root on any Linux system.
	if m["0"] != "root" {
		t.Errorf("UID 0 should map to root, got %q", m["0"])
	}
}
