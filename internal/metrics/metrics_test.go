package metrics

import (
	"testing"
	"time"
)

func TestReadCPUStat(t *testing.T) {
	s := readCPUStat()
	total := s.user + s.system + s.idle
	if total == 0 {
		t.Error("CPU stat total should be non-zero on a running system")
	}
}

func TestReadMemInfo(t *testing.T) {
	m := readMemInfo()
	if m["MemTotal"] == 0 {
		t.Error("MemTotal should be non-zero")
	}
	if _, ok := m["MemAvailable"]; !ok {
		t.Error("MemAvailable key should be present")
	}
}

func TestReadLoadAvg(t *testing.T) {
	la := readLoadAvg()
	if la[0] < 0 || la[1] < 0 || la[2] < 0 {
		t.Error("load averages should be non-negative")
	}
}

func TestReadUptime(t *testing.T) {
	up := readUptime()
	if up <= 0 {
		t.Errorf("uptime should be positive, got %f", up)
	}
}

func TestReadNetStat(t *testing.T) {
	s := readNetStat()
	// Just verify it doesn't panic; totals can be 0 on a fresh container.
	_ = s.rx
	_ = s.tx
}

func TestNewCollector(t *testing.T) {
	c := NewCollector()
	if c == nil {
		t.Fatal("NewCollector() returned nil")
	}
}

func TestCollector_LatestNilBeforeStart(t *testing.T) {
	c := NewCollector()
	if snap := c.Latest(); snap != nil {
		t.Error("Latest() should be nil before the first collect")
	}
}

func TestCollector_Collect(t *testing.T) {
	c := NewCollector()
	snap := c.collect()
	if snap == nil {
		t.Fatal("collect() returned nil")
	}
	if snap.MemTotal == 0 {
		t.Error("snap.MemTotal should be non-zero")
	}
	if snap.Uptime <= 0 {
		t.Error("snap.Uptime should be positive")
	}
	if snap.Timestamp.IsZero() {
		t.Error("snap.Timestamp should be set")
	}
}

func TestCollector_SubscribeUnsubscribe(t *testing.T) {
	c := NewCollector()
	ch := c.Subscribe()
	if ch == nil {
		t.Fatal("Subscribe() returned nil channel")
	}

	snap := c.collect()
	c.broadcast(snap)

	select {
	case received := <-ch:
		if received == nil {
			t.Error("received nil snapshot")
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("did not receive broadcast snapshot within timeout")
	}

	c.Unsubscribe(ch)
	// Channel must be closed after Unsubscribe.
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("channel should be closed after Unsubscribe")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("channel was not closed after Unsubscribe")
	}
}

func TestCollector_MultipleSubscribers(t *testing.T) {
	c := NewCollector()
	ch1 := c.Subscribe()
	ch2 := c.Subscribe()

	snap := c.collect()
	c.broadcast(snap)

	for i, ch := range []chan *Snapshot{ch1, ch2} {
		select {
		case got := <-ch:
			if got != snap {
				t.Errorf("subscriber %d: wrong snapshot", i)
			}
		case <-time.After(200 * time.Millisecond):
			t.Errorf("subscriber %d: did not receive broadcast", i)
		}
	}

	c.Unsubscribe(ch1)
	c.Unsubscribe(ch2)
}
