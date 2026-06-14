package firewall

import (
	"testing"
)

func TestValidatePort(t *testing.T) {
	valid := []string{"22", "80", "443", "8080", "65535", "1", "3000:4000"}
	for _, p := range valid {
		if err := validatePort(p); err != nil {
			t.Errorf("validatePort(%q) returned unexpected error: %v", p, err)
		}
	}

	invalid := []string{
		"",
		"0",
		"65536",
		"99999",
		"abc",
		"22/tcp",
		"-1",
		"22 80",
	}
	for _, p := range invalid {
		if err := validatePort(p); err == nil {
			t.Errorf("validatePort(%q) should have returned error", p)
		}
	}
}

func TestValidateAction(t *testing.T) {
	valid := []string{"allow", "deny", "reject", "limit"}
	for _, a := range valid {
		if err := validateAction(a); err != nil {
			t.Errorf("validateAction(%q) returned unexpected error: %v", a, err)
		}
	}

	invalid := []string{"", "permit", "block", "ALLOW", "Allow", "drop", "ufw allow"}
	for _, a := range invalid {
		if err := validateAction(a); err == nil {
			t.Errorf("validateAction(%q) should have returned error", a)
		}
	}
}

func TestDeleteRule_InvalidNum(t *testing.T) {
	// DeleteRule validates that num >= 1 before running any process.
	if err := DeleteRule(0); err == nil {
		t.Error("DeleteRule(0) should return error")
	}
	if err := DeleteRule(-5); err == nil {
		t.Error("DeleteRule(-5) should return error")
	}
}

func TestAddRule_InvalidProto(t *testing.T) {
	// AddRule validates proto before running any process.
	if err := AddRule("allow", "22", "sctp"); err == nil {
		t.Error("AddRule with invalid proto should return error")
	}
}

func TestRuleRegex(t *testing.T) {
	// Lines that should match the rule pattern
	matchLines := []string{
		"[ 1] 22/tcp                     ALLOW IN    Anywhere",
		"[ 2] 80/tcp                     DENY IN     Anywhere",
		"[10] 443                         ALLOW IN    Anywhere",
		"[ 3] 22/tcp (v6)                ALLOW IN    Anywhere (v6)",
	}
	for _, line := range matchLines {
		if m := reRule.FindStringSubmatch(line); m == nil {
			t.Errorf("rule regex should match: %q", line)
		}
	}

	// Lines that should not match
	noMatchLines := []string{
		"Status: active",
		"",
		"     To                         Action      From",
		"     --                         ------      ----",
	}
	for _, line := range noMatchLines {
		if m := reRule.FindStringSubmatch(line); m != nil {
			t.Errorf("rule regex should not match: %q", line)
		}
	}
}
