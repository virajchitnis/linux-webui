package exec

import (
	"errors"
	"os/exec"
	"regexp"
	"sync"
)

var (
	rePkgName  = regexp.MustCompile(`^[a-z0-9][a-z0-9.+\-]{0,127}$`)
	reUsername = regexp.MustCompile(`^[a-z_][a-z0-9_\-]{0,31}$`)
)

// paths is the resolved absolute path cache.
var (
	mu    sync.RWMutex
	paths = make(map[string]string)
)

// Resolve looks up a binary and caches its absolute path.
func Resolve(name string) (string, bool) {
	mu.RLock()
	p, ok := paths[name]
	mu.RUnlock()
	if ok {
		return p, true
	}
	p, err := exec.LookPath(name)
	if err != nil {
		return "", false
	}
	mu.Lock()
	paths[name] = p
	mu.Unlock()
	return p, true
}

// ValidPkgName returns true if name is safe to pass to apt/dnf/pacman.
func ValidPkgName(name string) bool {
	return rePkgName.MatchString(name)
}

// ValidUsername returns true if name is safe to pass to useradd/userdel.
func ValidUsername(name string) bool {
	return reUsername.MatchString(name)
}

var ErrForbidden = errors.New("binary not in resolved path cache")

// Command builds an *exec.Cmd using the resolved absolute path.
// Returns ErrForbidden if the binary was not previously resolved.
func Command(name string, args ...string) (*exec.Cmd, error) {
	path, ok := Resolve(name)
	if !ok {
		return nil, ErrForbidden
	}
	return exec.Command(path, args...), nil
}
