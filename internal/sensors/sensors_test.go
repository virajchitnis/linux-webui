package sensors

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAvailable(t *testing.T) {
	// Just verify it doesn't panic; may be true or false on CI.
	_ = Available()
}

func TestReadAll_NoPanic(t *testing.T) {
	// If hwmon is absent, ReadAll should return an error, not panic.
	chips, err := ReadAll()
	// Both outcomes are valid: a Docker/CI system may have no hwmon.
	if err != nil {
		t.Logf("ReadAll: %v (no hwmon sensors present)", err)
	} else {
		t.Logf("ReadAll: found %d chip(s)", len(chips))
	}
}

func TestReadFile(t *testing.T) {
	f, err := os.CreateTemp("", "sensor-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(f.Name()) })
	_, _ = f.WriteString("  42000  \n")
	f.Close()

	got := readFile(f.Name())
	if got != "42000" {
		t.Errorf("readFile = %q, want 42000", got)
	}
}

func TestReadFile_Missing(t *testing.T) {
	got := readFile("/nonexistent/path/sensor")
	if got != "" {
		t.Errorf("readFile(missing) = %q, want empty", got)
	}
}

func TestReadMilliDeg(t *testing.T) {
	f, _ := os.CreateTemp("", "millideg-*.txt")
	t.Cleanup(func() { _ = os.Remove(f.Name()) })
	_, _ = f.WriteString("52000")
	f.Close()

	got := readMilliDeg(f.Name())
	if got != 52.0 {
		t.Errorf("readMilliDeg = %f, want 52.0", got)
	}
}

func TestReadMilliDeg_Missing(t *testing.T) {
	got := readMilliDeg("/nonexistent/path")
	if got != 0 {
		t.Errorf("readMilliDeg(missing) = %f, want 0", got)
	}
}

func TestReadInt(t *testing.T) {
	f, _ := os.CreateTemp("", "readint-*.txt")
	t.Cleanup(func() { _ = os.Remove(f.Name()) })
	_, _ = f.WriteString("1200")
	f.Close()

	got := readInt(f.Name())
	if got != 1200 {
		t.Errorf("readInt = %d, want 1200", got)
	}
}

func TestFileExists(t *testing.T) {
	f, _ := os.CreateTemp("", "exists-*.txt")
	name := f.Name()
	f.Close()
	t.Cleanup(func() { _ = os.Remove(name) })

	if !fileExists(name) {
		t.Error("fileExists(existing) should be true")
	}
	if fileExists("/nonexistent/path/xyz") {
		t.Error("fileExists(nonexistent) should be false")
	}
}

func TestReadChip_SyntheticHwmon(t *testing.T) {
	// Build a synthetic hwmon directory to test readChip.
	dir := t.TempDir()

	_ = os.WriteFile(filepath.Join(dir, "name"), []byte("testchip\n"), 0644)
	_ = os.WriteFile(filepath.Join(dir, "temp1_input"), []byte("45000\n"), 0644)
	_ = os.WriteFile(filepath.Join(dir, "temp1_label"), []byte("Core 0\n"), 0644)
	_ = os.WriteFile(filepath.Join(dir, "temp1_max"), []byte("95000\n"), 0644)
	_ = os.WriteFile(filepath.Join(dir, "temp1_crit"), []byte("105000\n"), 0644)
	_ = os.WriteFile(filepath.Join(dir, "fan1_input"), []byte("1200\n"), 0644)
	_ = os.WriteFile(filepath.Join(dir, "fan1_label"), []byte("CPU Fan\n"), 0644)

	chip := readChip(dir)

	if chip.Name != "testchip" {
		t.Errorf("chip.Name = %q, want testchip", chip.Name)
	}
	if len(chip.Temps) != 1 {
		t.Fatalf("expected 1 temp, got %d", len(chip.Temps))
	}
	if chip.Temps[0].Cur != 45.0 {
		t.Errorf("temp cur = %f, want 45.0", chip.Temps[0].Cur)
	}
	if chip.Temps[0].Max != 95.0 {
		t.Errorf("temp max = %f, want 95.0", chip.Temps[0].Max)
	}
	if chip.Temps[0].Crit != 105.0 {
		t.Errorf("temp crit = %f, want 105.0", chip.Temps[0].Crit)
	}
	if len(chip.Fans) != 1 {
		t.Fatalf("expected 1 fan, got %d", len(chip.Fans))
	}
	if chip.Fans[0].RPM != 1200 {
		t.Errorf("fan rpm = %d, want 1200", chip.Fans[0].RPM)
	}
}
