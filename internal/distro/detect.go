package distro

import (
	"bufio"
	"os"
	"strings"
)

type PackageFamily int

const (
	FamilyDebian  PackageFamily = iota // apt-get
	FamilyRPM                          // dnf / yum
	FamilyPacman                       // pacman
	FamilyPortage                      // emerge
	FamilyUnknown
)

type Distro struct {
	ID      string
	Like    []string
	Version string
	Family  PackageFamily
}

func (d *Distro) FamilyString() string {
	switch d.Family {
	case FamilyDebian:
		return "debian"
	case FamilyRPM:
		return "rpm"
	case FamilyPacman:
		return "pacman"
	case FamilyPortage:
		return "portage"
	default:
		return "unknown"
	}
}

func Detect() *Distro {
	fields := parseOSRelease("/etc/os-release")
	id := strings.ToLower(fields["ID"])
	like := strings.Fields(strings.ToLower(fields["ID_LIKE"]))
	version := fields["VERSION_ID"]

	d := &Distro{ID: id, Like: like, Version: version}
	d.Family = detectFamily(id, like)
	return d
}

func detectFamily(id string, like []string) PackageFamily {
	all := append([]string{id}, like...)
	for _, v := range all {
		switch v {
		case "debian", "ubuntu", "linuxmint", "raspbian", "pop":
			return FamilyDebian
		case "fedora", "rhel", "centos", "almalinux", "rocky", "opensuse", "opensuse-leap", "opensuse-tumbleweed", "sles":
			return FamilyRPM
		case "arch", "manjaro", "endeavouros":
			return FamilyPacman
		case "gentoo":
			return FamilyPortage
		}
	}
	return FamilyUnknown
}

func parseOSRelease(path string) map[string]string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	m := make(map[string]string)
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"`)
		m[key] = val
	}
	return m
}
