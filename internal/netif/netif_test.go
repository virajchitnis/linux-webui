package netif

import (
	"testing"
)

func TestList(t *testing.T) {
	ifaces, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if len(ifaces) == 0 {
		t.Error("expected at least one network interface")
	}
	var foundLoopback bool
	for _, iface := range ifaces {
		if iface.Name == "" {
			t.Error("interface name should not be empty")
		}
		if iface.Name == "lo" {
			foundLoopback = true
			if iface.Type != "loopback" {
				t.Errorf("lo type = %q, want loopback", iface.Type)
			}
		}
	}
	if !foundLoopback {
		t.Error("expected loopback interface 'lo'")
	}
}

func TestGuessType(t *testing.T) {
	cases := []struct{ name, want string }{
		{"lo", "loopback"},
		{"lo0", "loopback"},
		{"eth0", "ethernet"},
		{"enp3s0", "ethernet"},
		{"eno1", "ethernet"},
		{"wlan0", "wifi"},
		{"wlp2s0", "wifi"},
		{"wifi0", "wifi"},
		{"docker0", "bridge"},
		{"br-abc123", "bridge"},
		{"virbr0", "bridge"},
		{"veth0abc", "virtual"},
		{"tun0", "virtual"},
		{"tap1", "virtual"},
		{"wg0", "wireguard"},
		{"ppp0", "other"},
		{"bond0", "other"},
	}
	for _, tc := range cases {
		got := guessType(tc.name)
		if got != tc.want {
			t.Errorf("guessType(%q) = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestFmtBytes(t *testing.T) {
	cases := []struct {
		input uint64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
	}
	for _, tc := range cases {
		got := FmtBytes(tc.input)
		if got != tc.want {
			t.Errorf("FmtBytes(%d) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
