package users

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// User mirrors a /etc/passwd entry.
type User struct {
	Username string `json:"username"`
	UID      int    `json:"uid"`
	GID      int    `json:"gid"`
	GECOS    string `json:"gecos"`
	Home     string `json:"home"`
	Shell    string `json:"shell"`
	System   bool   `json:"system"` // UID < 1000
}

// Group mirrors a /etc/group entry.
type Group struct {
	Name    string   `json:"name"`
	GID     int      `json:"gid"`
	Members []string `json:"members"`
}

// ListUsers returns all entries from /etc/passwd.
func ListUsers() ([]User, error) {
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []User
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		u, ok := parsePasswd(sc.Text())
		if ok {
			out = append(out, u)
		}
	}
	return out, sc.Err()
}

// ListGroups returns all entries from /etc/group.
func ListGroups() ([]Group, error) {
	f, err := os.Open("/etc/group")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []Group
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		g, ok := parseGroup(sc.Text())
		if ok {
			out = append(out, g)
		}
	}
	return out, sc.Err()
}

// CreateUser runs `sudo useradd -m -s /bin/bash username`.
func CreateUser(username string) error {
	if !validUsername(username) {
		return fmt.Errorf("invalid username: %q", username)
	}
	cmd := exec.Command("/usr/bin/sudo", "/usr/sbin/useradd", "-m", "-s", "/bin/bash", username)
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("useradd: %s", strings.TrimSpace(errBuf.String()))
	}
	return nil
}

// DeleteUser runs `sudo userdel -r username`.
func DeleteUser(username string) error {
	if !validUsername(username) {
		return fmt.Errorf("invalid username: %q", username)
	}
	cmd := exec.Command("/usr/bin/sudo", "/usr/sbin/userdel", "-r", username)
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("userdel: %s", strings.TrimSpace(errBuf.String()))
	}
	return nil
}

// SetPassword sets a user's password using `sudo chpasswd`.
// password must be the plain-text password (sent over HTTPS, never logged).
func SetPassword(username, password string) error {
	if !validUsername(username) {
		return fmt.Errorf("invalid username: %q", username)
	}
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	cmd := exec.Command("/usr/bin/sudo", "/usr/sbin/chpasswd")
	cmd.Stdin = strings.NewReader(username + ":" + password + "\n")
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("chpasswd: %s", strings.TrimSpace(errBuf.String()))
	}
	return nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func parsePasswd(line string) (User, bool) {
	if strings.HasPrefix(line, "#") || line == "" {
		return User{}, false
	}
	p := strings.SplitN(line, ":", 7)
	if len(p) < 7 {
		return User{}, false
	}
	uid, _ := strconv.Atoi(p[2])
	gid, _ := strconv.Atoi(p[3])
	return User{
		Username: p[0], UID: uid, GID: gid,
		GECOS: p[4], Home: p[5], Shell: p[6],
		System: uid < 1000,
	}, true
}

func parseGroup(line string) (Group, bool) {
	if strings.HasPrefix(line, "#") || line == "" {
		return Group{}, false
	}
	p := strings.SplitN(line, ":", 4)
	if len(p) < 4 {
		return Group{}, false
	}
	gid, _ := strconv.Atoi(p[2])
	members := []string{}
	if p[3] != "" {
		members = strings.Split(p[3], ",")
	}
	return Group{Name: p[0], GID: gid, Members: members}, true
}

// validUsername enforces the Debian/Ubuntu username policy.
func validUsername(name string) bool {
	if len(name) == 0 || len(name) > 32 {
		return false
	}
	for i, c := range name {
		if i == 0 && !((c >= 'a' && c <= 'z') || c == '_') {
			return false
		}
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return false
		}
	}
	return true
}
