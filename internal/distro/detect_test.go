package distro

import (
	"os"
	"testing"
)

func TestDetectFamily(t *testing.T) {
	cases := []struct {
		id     string
		like   []string
		want   PackageFamily
	}{
		{"ubuntu", nil, FamilyDebian},
		{"debian", nil, FamilyDebian},
		{"linuxmint", nil, FamilyDebian},
		{"pop", nil, FamilyDebian},
		// Ubuntu reports ID=ubuntu but also has ID_LIKE=debian
		{"ubuntu", []string{"debian"}, FamilyDebian},
		{"fedora", nil, FamilyRPM},
		{"rhel", nil, FamilyRPM},
		{"centos", nil, FamilyRPM},
		{"almalinux", nil, FamilyRPM},
		{"rocky", nil, FamilyRPM},
		{"opensuse-leap", nil, FamilyRPM},
		{"arch", nil, FamilyPacman},
		{"manjaro", nil, FamilyPacman},
		{"endeavouros", nil, FamilyPacman},
		{"gentoo", nil, FamilyPortage},
		{"unknown-distro", nil, FamilyUnknown},
		{"", nil, FamilyUnknown},
		// Distro with unknown ID but known ID_LIKE
		{"linux", []string{"rhel", "fedora"}, FamilyRPM},
		{"linux", []string{"debian"}, FamilyDebian},
	}

	for _, tc := range cases {
		got := detectFamily(tc.id, tc.like)
		if got != tc.want {
			t.Errorf("detectFamily(%q, %v) = %v, want %v", tc.id, tc.like, got, tc.want)
		}
	}
}

func TestFamilyString(t *testing.T) {
	cases := []struct {
		family PackageFamily
		want   string
	}{
		{FamilyDebian, "debian"},
		{FamilyRPM, "rpm"},
		{FamilyPacman, "pacman"},
		{FamilyPortage, "portage"},
		{FamilyUnknown, "unknown"},
	}
	for _, tc := range cases {
		d := &Distro{Family: tc.family}
		if got := d.FamilyString(); got != tc.want {
			t.Errorf("FamilyString() = %q, want %q", got, tc.want)
		}
	}
}

func TestParseOSRelease(t *testing.T) {
	content := `NAME="Ubuntu"
ID=ubuntu
ID_LIKE=debian
VERSION_ID="24.04"
# comment line
EMPTY=
QUOTED="hello world"
`
	f, err := os.CreateTemp("", "os-release-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(f.Name()) })
	_, _ = f.WriteString(content)
	f.Close()

	m := parseOSRelease(f.Name())
	if m["ID"] != "ubuntu" {
		t.Errorf("ID = %q, want ubuntu", m["ID"])
	}
	if m["ID_LIKE"] != "debian" {
		t.Errorf("ID_LIKE = %q, want debian", m["ID_LIKE"])
	}
	if m["VERSION_ID"] != "24.04" {
		t.Errorf("VERSION_ID = %q, want 24.04", m["VERSION_ID"])
	}
	if m["QUOTED"] != "hello world" {
		t.Errorf("QUOTED = %q, want 'hello world'", m["QUOTED"])
	}
	if _, ok := m["# comment"]; ok {
		t.Error("comment lines should not be parsed")
	}
}

func TestParseOSRelease_Missing(t *testing.T) {
	m := parseOSRelease("/nonexistent/path/os-release")
	if m != nil {
		t.Error("missing file should return nil map")
	}
}
