package files_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/virajchitnis/linux-webui/internal/files"
)

// makeRoot creates a temporary directory tree for tests.
func makeRoot(t *testing.T) string {
	t.Helper()
	root, err := os.MkdirTemp("", "files-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(root) })

	// Create some entries inside
	_ = os.MkdirAll(filepath.Join(root, "subdir"), 0755)
	_ = os.WriteFile(filepath.Join(root, "file.txt"), []byte("hello"), 0644)
	_ = os.WriteFile(filepath.Join(root, "subdir", "inner.txt"), []byte("inner"), 0644)
	return root
}

func TestJailValid(t *testing.T) {
	root := makeRoot(t)
	roots := files.Roots{root}

	// Root itself
	p, err := roots.Jail(root)
	if err != nil {
		t.Errorf("Jail(root) error: %v", err)
	}
	if p != root {
		t.Errorf("Jail(root) = %q, want %q", p, root)
	}

	// File inside root
	target := filepath.Join(root, "file.txt")
	p, err = roots.Jail(target)
	if err != nil {
		t.Errorf("Jail(file.txt) error: %v", err)
	}
	if p != target {
		t.Errorf("Jail(file.txt) = %q, want %q", p, target)
	}

	// Subdir
	sub := filepath.Join(root, "subdir")
	p, err = roots.Jail(sub)
	if err != nil {
		t.Errorf("Jail(subdir) error: %v", err)
	}
	if p != sub {
		t.Errorf("Jail(subdir) = %q, want %q", p, sub)
	}
}

func TestJailEscape_DotDot(t *testing.T) {
	root := makeRoot(t)
	roots := files.Roots{root}

	// Path traversal attempts
	escapes := []string{
		root + "/../etc/passwd",
		root + "/subdir/../../etc/shadow",
		"/etc/passwd",
		"/",
		root[:len(root)-1], // truncated root (parent or sibling)
	}

	for _, path := range escapes {
		_, err := roots.Jail(path)
		if err != files.ErrEscape {
			t.Errorf("Jail(%q) should return ErrEscape, got %v", path, err)
		}
	}
}

func TestJailEscape_SiblingDir(t *testing.T) {
	// /var/log should not grant access to /var/log2
	parent, err := os.MkdirTemp("", "sib-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(parent)

	jailRoot := filepath.Join(parent, "log")
	sibling := filepath.Join(parent, "log2")
	_ = os.Mkdir(jailRoot, 0755)
	_ = os.Mkdir(sibling, 0755)

	roots := files.Roots{jailRoot}
	_, err = roots.Jail(sibling)
	if err != files.ErrEscape {
		t.Errorf("Jail(sibling) should return ErrEscape, got %v (root=%q, sibling=%q)", err, jailRoot, sibling)
	}
}

func TestJailSymlinkEscape(t *testing.T) {
	root := makeRoot(t)
	// Create a symlink that points outside the jail
	outside, _ := os.MkdirTemp("", "outside-*")
	defer os.RemoveAll(outside)

	symlinkPath := filepath.Join(root, "evil-link")
	_ = os.Symlink(outside, symlinkPath)

	roots := files.Roots{root}
	_, err := roots.Jail(symlinkPath)
	if err != files.ErrEscape {
		t.Errorf("Jail(symlink-escape) should return ErrEscape, got %v", err)
	}
}

func TestJailSymlinkWithinRoot(t *testing.T) {
	root := makeRoot(t)
	// Symlink within the jail pointing to another file in the jail
	linkPath := filepath.Join(root, "link-to-file")
	target := filepath.Join(root, "file.txt")
	_ = os.Symlink(target, linkPath)

	roots := files.Roots{root}
	resolved, err := roots.Jail(linkPath)
	if err != nil {
		t.Errorf("Jail(symlink-within-root) error: %v", err)
	}
	if resolved != target {
		t.Errorf("Jail resolved to %q, want %q", resolved, target)
	}
}

func TestJailNonExistentPath(t *testing.T) {
	root := makeRoot(t)
	roots := files.Roots{root}

	// Non-existent path inside the jail should be allowed (returns cleaned path)
	nonExistent := filepath.Join(root, "does-not-exist.txt")
	p, err := roots.Jail(nonExistent)
	if err != nil {
		t.Errorf("Jail(non-existent inside root) should not error, got %v", err)
	}
	if p != nonExistent {
		t.Errorf("Jail(non-existent) = %q, want %q", p, nonExistent)
	}
}

func TestJailEmptyRoots(t *testing.T) {
	roots := files.Roots{}
	_, err := roots.Jail("/etc/passwd")
	if err != files.ErrEscape {
		t.Errorf("empty roots should always return ErrEscape, got %v", err)
	}
}

func TestListDir(t *testing.T) {
	root := makeRoot(t)
	roots := files.Roots{root}

	entries, err := files.ListDir(root, roots)
	if err != nil {
		t.Fatal(err)
	}

	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Name] = true
	}
	if !names["subdir"] {
		t.Error("expected 'subdir' in listing")
	}
	if !names["file.txt"] {
		t.Error("expected 'file.txt' in listing")
	}
}

func TestReadFile(t *testing.T) {
	root := makeRoot(t)
	roots := files.Roots{root}

	target := filepath.Join(root, "file.txt")
	data, err := files.ReadFile(target, roots)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Errorf("ReadFile content = %q, want 'hello'", data)
	}
}

func TestReadFile_Escape(t *testing.T) {
	root := makeRoot(t)
	roots := files.Roots{root}

	_, err := files.ReadFile("/etc/passwd", roots)
	if err != files.ErrEscape {
		t.Errorf("ReadFile(/etc/passwd) should return ErrEscape, got %v", err)
	}
}
