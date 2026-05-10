package apt

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

// Package represents a single APT package entry.
type Package struct {
	Name      string `json:"name"`
	Available string `json:"available"`
	Current   string `json:"current"`
	Arch      string `json:"arch"`
}

// reUpgradable matches lines from `apt list --upgradable`:
// bash/jammy-updates 5.1-6ubuntu1.1 amd64 [upgradable from: 5.1-6ubuntu1]
var reUpgradable = regexp.MustCompile(`^([^/]+)/\S+\s+(\S+)\s+(\S+)\s+\[upgradable from:\s+([^\]]+)\]`)

// ListUpgradable returns all packages that have available upgrades.
func ListUpgradable() ([]Package, error) {
	cmd := exec.Command("apt", "list", "--upgradable")
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive", "LANG=C")
	out, err := cmd.Output()
	if err != nil && len(out) == 0 {
		return nil, err
	}
	var pkgs []Package
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		m := reUpgradable.FindStringSubmatch(sc.Text())
		if m == nil {
			continue
		}
		pkgs = append(pkgs, Package{Name: m[1], Available: m[2], Arch: m[3], Current: m[4]})
	}
	return pkgs, nil
}

// jobMu serialises apt-get jobs — only one job at a time.
var jobMu sync.Mutex

// TryLock attempts to acquire the single-job mutex.
func TryLock() bool { return jobMu.TryLock() }

// Unlock releases the single-job mutex.
func Unlock() { jobMu.Unlock() }

// IsRunning reports whether an apt job is currently executing.
func IsRunning() bool {
	if jobMu.TryLock() {
		jobMu.Unlock()
		return false
	}
	return true
}

// RunUpdate runs `sudo apt-get update` and streams output lines to ch.
func RunUpdate(ctx context.Context, ch chan<- string) error {
	return runApt(ctx, ch, "update")
}

// RunUpgrade runs `sudo apt-get upgrade -y` and streams output lines to ch.
func RunUpgrade(ctx context.Context, ch chan<- string) error {
	return runApt(ctx, ch, "upgrade", "-y")
}

// RunInstall runs `sudo apt-get install -y -- pkg` and streams output lines to ch.
func RunInstall(ctx context.Context, pkg string, ch chan<- string) error {
	return runApt(ctx, ch, "install", "-y", "--", pkg)
}

func runApt(ctx context.Context, ch chan<- string, args ...string) error {
	// Use absolute path; sudoers grants linux-admin NOPASSWD for apt-get
	full := append([]string{"/usr/bin/apt-get"}, args...)
	cmd := exec.CommandContext(ctx, "/usr/bin/sudo", full...)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive", "LANG=C")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	pipe := func(r io.Reader) {
		defer wg.Done()
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			select {
			case <-ctx.Done():
				return
			case ch <- sc.Text():
			}
		}
	}
	wg.Add(2)
	go pipe(stdout)
	go pipe(stderr)
	wg.Wait()

	return cmd.Wait()
}
