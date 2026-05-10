package firewall

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Status contains the overall UFW state and active rules.
type Status struct {
	Enabled bool   `json:"enabled"`
	Rules   []Rule `json:"rules"`
}

// Rule represents a single UFW rule entry.
type Rule struct {
	Num      int    `json:"num"`
	To       string `json:"to"`
	Action   string `json:"action"`
	From     string `json:"from"`
	Comment  string `json:"comment,omitempty"`
}

// reRule matches lines like: [ 1] 22/tcp  ALLOW IN  Anywhere
var reRule = regexp.MustCompile(`^\[\s*(\d+)\]\s+(\S+(?:\s+\(v6\))?)\s+(ALLOW IN|DENY IN|REJECT IN|LIMIT IN|ALLOW OUT|DENY OUT|ALLOW|DENY|REJECT|LIMIT)\s+(.+)$`)

// GetStatus returns the current UFW status and numbered rules.
func GetStatus() (*Status, error) {
	out, err := runSudo("ufw", "status", "numbered")
	if err != nil {
		return nil, fmt.Errorf("ufw status: %w", err)
	}

	s := &Status{}
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "Status: active") {
			s.Enabled = true
		}
		m := reRule.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		num, _ := strconv.Atoi(m[1])
		action := m[3]
		// Normalise action to just ALLOW/DENY/REJECT/LIMIT
		if idx := strings.Index(action, " "); idx >= 0 {
			action = action[:idx]
		}
		s.Rules = append(s.Rules, Rule{
			Num:    num,
			To:     strings.TrimSpace(m[2]),
			Action: action,
			From:   strings.TrimSpace(m[4]),
		})
	}
	return s, nil
}

// AddRule runs `sudo ufw {action} {port}/{proto}`.
// action must be "allow" or "deny"; proto must be "tcp" or "udp".
func AddRule(action, port, proto string) error {
	if err := validateAction(action); err != nil {
		return err
	}
	if err := validatePort(port); err != nil {
		return err
	}
	if proto != "tcp" && proto != "udp" && proto != "any" {
		return fmt.Errorf("proto must be tcp, udp, or any")
	}
	target := port
	if proto != "any" {
		target = port + "/" + proto
	}
	_, err := runSudo("ufw", action, target)
	return err
}

// DeleteRule runs `sudo ufw delete {num}` where num is the 1-based rule number.
func DeleteRule(num int) error {
	if num < 1 {
		return fmt.Errorf("rule number must be ≥ 1")
	}
	// ufw delete is interactive; pass "yes" via stdin
	cmd := exec.Command("/usr/bin/sudo", "/usr/sbin/ufw", "--force", "delete", strconv.Itoa(num))
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ufw delete: %s", strings.TrimSpace(errBuf.String()))
	}
	return nil
}

// Enable runs `sudo ufw --force enable`.
func Enable() error {
	_, err := runSudo("ufw", "--force", "enable")
	return err
}

// Disable runs `sudo ufw disable`.
func Disable() error {
	_, err := runSudo("ufw", "disable")
	return err
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func runSudo(args ...string) (string, error) {
	full := append([]string{"/usr/sbin/ufw"}, args[1:]...)
	cmd := exec.Command("/usr/bin/sudo", full...)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errBuf.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%s", msg)
	}
	return out.String(), nil
}

var rePort = regexp.MustCompile(`^(\d{1,5})(:\d{1,5})?$`)

func validatePort(port string) error {
	m := rePort.FindStringSubmatch(port)
	if m == nil {
		return fmt.Errorf("invalid port: %q", port)
	}
	p, _ := strconv.Atoi(m[1])
	if p < 1 || p > 65535 {
		return fmt.Errorf("port out of range: %d", p)
	}
	return nil
}

func validateAction(action string) error {
	switch action {
	case "allow", "deny", "reject", "limit":
		return nil
	}
	return fmt.Errorf("action must be allow, deny, reject, or limit")
}
