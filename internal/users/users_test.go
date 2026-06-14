package users

import (
	"testing"
)

func TestValidUsername(t *testing.T) {
	valid := []string{
		"alice",
		"bob",
		"_service",
		"user1",
		"web-app",
		"a",
		"a1b2c3",
	}
	for _, name := range valid {
		if !validUsername(name) {
			t.Errorf("validUsername(%q) = false, want true", name)
		}
	}

	invalid := []string{
		"",
		"1starts-with-digit",
		"-starts-with-dash",
		"Has-Uppercase",
		"has space",
		"has.dot",
		"has@symbol",
		"toolongnamethatexceedsthirtytwocharacterslimit_x",
	}
	for _, name := range invalid {
		if validUsername(name) {
			t.Errorf("validUsername(%q) = true, want false", name)
		}
	}
}

func TestValidUsernameLength(t *testing.T) {
	// Exactly 32 chars (max)
	name32 := "abcdefghijklmnopqrstuvwxyz123456"
	if len(name32) != 32 {
		t.Fatalf("test fixture wrong length: %d", len(name32))
	}
	if !validUsername(name32) {
		t.Errorf("32-char username should be valid")
	}

	// 33 chars (too long)
	name33 := name32 + "a"
	if validUsername(name33) {
		t.Errorf("33-char username should be invalid")
	}
}

func TestParsePasswd(t *testing.T) {
	cases := []struct {
		line  string
		ok    bool
		uid   int
		gid   int
		uname string
	}{
		{
			line:  "root:x:0:0:root:/root:/bin/bash",
			ok:    true, uid: 0, gid: 0, uname: "root",
		},
		{
			line:  "alice:x:1000:1000:Alice Jones:/home/alice:/bin/bash",
			ok:    true, uid: 1000, gid: 1000, uname: "alice",
		},
		{
			line:  "nobody:x:65534:65534:nobody:/nonexistent:/usr/sbin/nologin",
			ok:    true, uid: 65534, gid: 65534, uname: "nobody",
		},
		{line: "# comment", ok: false},
		{line: "", ok: false},
		{line: "malformed:line", ok: false},
	}

	for _, tc := range cases {
		u, ok := parsePasswd(tc.line)
		if ok != tc.ok {
			t.Errorf("parsePasswd(%q) ok=%v, want %v", tc.line, ok, tc.ok)
			continue
		}
		if !ok {
			continue
		}
		if u.Username != tc.uname {
			t.Errorf("parsePasswd(%q).Username = %q, want %q", tc.line, u.Username, tc.uname)
		}
		if u.UID != tc.uid {
			t.Errorf("parsePasswd(%q).UID = %d, want %d", tc.line, u.UID, tc.uid)
		}
		if u.GID != tc.gid {
			t.Errorf("parsePasswd(%q).GID = %d, want %d", tc.line, u.GID, tc.gid)
		}
	}
}

func TestParsePasswd_SystemFlag(t *testing.T) {
	sysUser, _ := parsePasswd("daemon:x:1:1:daemon:/usr/sbin:/usr/sbin/nologin")
	if !sysUser.System {
		t.Error("UID < 1000 should set System=true")
	}

	normalUser, _ := parsePasswd("alice:x:1000:1000::/home/alice:/bin/bash")
	if normalUser.System {
		t.Error("UID >= 1000 should set System=false")
	}
}

func TestListUsers(t *testing.T) {
	users, err := ListUsers()
	if err != nil {
		t.Fatal(err)
	}
	if len(users) == 0 {
		t.Error("expected at least one user from /etc/passwd")
	}
	var foundRoot bool
	for _, u := range users {
		if u.Username == "root" && u.UID == 0 {
			foundRoot = true
		}
	}
	if !foundRoot {
		t.Error("expected root user (UID=0) in /etc/passwd")
	}
}

func TestListGroups(t *testing.T) {
	groups, err := ListGroups()
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) == 0 {
		t.Error("expected at least one group from /etc/group")
	}
	var foundRoot bool
	for _, g := range groups {
		if g.Name == "root" && g.GID == 0 {
			foundRoot = true
		}
	}
	if !foundRoot {
		t.Error("expected root group (GID=0) in /etc/group")
	}
}

func TestCreateUser_InvalidUsername(t *testing.T) {
	err := CreateUser("Invalid-Name-Uppercase")
	if err == nil {
		t.Error("CreateUser with invalid username should return error")
	}
}

func TestDeleteUser_InvalidUsername(t *testing.T) {
	err := DeleteUser("Has Spaces")
	if err == nil {
		t.Error("DeleteUser with invalid username should return error")
	}
}

func TestSetPassword_InvalidUsername(t *testing.T) {
	err := SetPassword("HasUppercase", "password123")
	if err == nil {
		t.Error("SetPassword with invalid username should return error")
	}
}

func TestSetPassword_TooShort(t *testing.T) {
	err := SetPassword("alice", "short")
	if err == nil {
		t.Error("SetPassword with short password should return error")
	}
}

func TestParseGroup(t *testing.T) {
	cases := []struct {
		line    string
		ok      bool
		name    string
		gid     int
		members []string
	}{
		{
			line: "root:x:0:",
			ok:   true, name: "root", gid: 0, members: []string{},
		},
		{
			line: "sudo:x:27:alice,bob",
			ok:   true, name: "sudo", gid: 27, members: []string{"alice", "bob"},
		},
		{
			line: "plugdev:x:46:alice",
			ok:   true, name: "plugdev", gid: 46, members: []string{"alice"},
		},
		{line: "# comment", ok: false},
		{line: "", ok: false},
	}

	for _, tc := range cases {
		g, ok := parseGroup(tc.line)
		if ok != tc.ok {
			t.Errorf("parseGroup(%q) ok=%v, want %v", tc.line, ok, tc.ok)
			continue
		}
		if !ok {
			continue
		}
		if g.Name != tc.name {
			t.Errorf("parseGroup(%q).Name = %q, want %q", tc.line, g.Name, tc.name)
		}
		if g.GID != tc.gid {
			t.Errorf("parseGroup(%q).GID = %d, want %d", tc.line, g.GID, tc.gid)
		}
		if len(g.Members) != len(tc.members) {
			t.Errorf("parseGroup(%q) members = %v, want %v", tc.line, g.Members, tc.members)
		}
	}
}
