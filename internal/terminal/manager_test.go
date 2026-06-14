package terminal

import (
	"errors"
	"os"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	m := NewManager(5, time.Minute)
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.Count() != 0 {
		t.Errorf("new manager count = %d, want 0", m.Count())
	}
}

func TestManager_Get_Missing(t *testing.T) {
	m := NewManager(5, time.Minute)
	_, ok := m.Get("nonexistent")
	if ok {
		t.Error("Get(nonexistent) should return false")
	}
}

func TestManager_Count_Empty(t *testing.T) {
	m := NewManager(5, time.Minute)
	if m.Count() != 0 {
		t.Errorf("empty manager count = %d, want 0", m.Count())
	}
}

func TestManager_CloseAll_Empty(t *testing.T) {
	m := NewManager(5, time.Minute)
	m.CloseAll() // should not panic on empty manager
}

func TestManager_Close_Missing(t *testing.T) {
	m := NewManager(5, time.Minute)
	m.Close("nonexistent") // should be a no-op, not panic
}

func TestIsTimeout_True(t *testing.T) {
	// os.ErrDeadlineExceeded implements Timeout() == true
	if !isTimeout(os.ErrDeadlineExceeded) {
		t.Error("os.ErrDeadlineExceeded should be a timeout error")
	}
}

func TestIsTimeout_False(t *testing.T) {
	if isTimeout(errors.New("not a timeout")) {
		t.Error("plain error should not be a timeout")
	}
}

func TestManager_Create_MaxSessions(t *testing.T) {
	// Max 0 sessions → immediate error.
	m := NewManager(0, time.Minute)
	_, err := m.Create("test-id")
	if err == nil {
		t.Error("Create on maxSessions=0 manager should return error")
	}
}
