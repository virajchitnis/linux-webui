// Package files provides a read-only, path-jailed file browser.
package files

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var ErrEscape = errors.New("path escapes jail root")
var ErrTooLarge = errors.New("file too large to read (limit 1 MB)")

const maxReadSize = 1 << 20 // 1 MB

// Entry represents a single directory entry.
type Entry struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"` // relative to jail root
	IsDir   bool      `json:"is_dir"`
	Size    int64     `json:"size"`
	Mode    string    `json:"mode"`
	ModTime time.Time `json:"mod_time"`
	Symlink string    `json:"symlink,omitempty"`
}

// Roots holds the allowed top-level directories.
type Roots []string

// Jail resolves userPath relative to one of the roots and validates
// that the cleaned, symlink-resolved result stays within that root.
func (rs Roots) Jail(userPath string) (string, error) {
	cleaned := filepath.Clean("/" + userPath)
	for _, root := range rs {
		root = filepath.Clean(root)
		if cleaned == root || strings.HasPrefix(cleaned, root+string(filepath.Separator)) {
			// Resolve symlinks to prevent escapes.
			resolved, err := filepath.EvalSymlinks(cleaned)
			if err != nil {
				// Allow the path even if it doesn't exist yet (for listing).
				return cleaned, nil
			}
			if resolved == root || strings.HasPrefix(resolved, root+string(filepath.Separator)) {
				return resolved, nil
			}
			return "", ErrEscape
		}
	}
	return "", ErrEscape
}

// ListDir returns directory entries for the given (jail-validated) path.
func ListDir(path string, roots Roots) ([]Entry, error) {
	jailed, err := roots.Jail(path)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(jailed)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	infos, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}
	sort.Slice(infos, func(i, j int) bool {
		if infos[i].IsDir() != infos[j].IsDir() {
			return infos[i].IsDir()
		}
		return infos[i].Name() < infos[j].Name()
	})

	entries := make([]Entry, 0, len(infos))
	for _, info := range infos {
		full := filepath.Join(jailed, info.Name())
		e := Entry{
			Name:    info.Name(),
			Path:    filepath.Join(path, info.Name()),
			IsDir:   info.IsDir(),
			Size:    info.Size(),
			Mode:    info.Mode().String(),
			ModTime: info.ModTime(),
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			if target, err := os.Readlink(full); err == nil {
				e.Symlink = target
			}
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// ReadFile returns the content of a file (max 1 MB).
func ReadFile(path string, roots Roots) ([]byte, error) {
	jailed, err := roots.Jail(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(jailed)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, errors.New("path is a directory")
	}
	if info.Size() > maxReadSize {
		return nil, ErrTooLarge
	}
	f, err := os.Open(jailed)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(io.LimitReader(f, maxReadSize+1))
}
