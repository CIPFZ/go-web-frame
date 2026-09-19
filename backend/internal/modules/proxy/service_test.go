package proxy

import (
	"strings"
	"testing"
)

func validInstance() Instance {
	return Instance{Name: "sing-box", Engine: EngineSingBox, Scope: ScopeSystem,
		Unit: "sing-box.service", BinaryPath: "/usr/bin/sing-box", ConfigPath: "/etc/sing-box/config.json"}
}

func TestValidateInstanceAllowlist(t *testing.T) {
	if err := validateInstance(validInstance()); err != nil {
		t.Fatalf("valid instance rejected: %v", err)
	}
	cases := []Instance{
		func() Instance { v := validInstance(); v.Unit = "sing box.service"; return v }(),
		func() Instance { v := validInstance(); v.ConfigPath = "tmp/config.json"; return v }(),
		func() Instance { v := validInstance(); v.BinaryPath = "/tmp/sing-box"; return v }(),
		func() Instance { v := validInstance(); v.Engine = "unknown"; return v }(),
	}
	for _, item := range cases {
		if err := validateInstance(item); err == nil {
			t.Fatalf("expected invalid instance to be rejected: %+v", item)
		}
	}
}

func TestParseProperties(t *testing.T) {
	parsed := parseProperties(`ActiveState=active
SubState=running
MainPID=123
`)
	if parsed["ActiveState"] != "active" || parsed["SubState"] != "running" || parsed["MainPID"] != "123" {
		t.Fatalf("unexpected properties: %#v", parsed)
	}
}

func TestDigest(t *testing.T) {
	got := digest([]byte("base-frame"))
	if len(got) != 64 || strings.Trim(got, "0123456789abcdef") != "" {
		t.Fatalf("invalid sha256 digest: %q", got)
	}
	if got != digest([]byte("base-frame")) {
		t.Fatal("digest is not deterministic")
	}
	if got == digest([]byte("other")) {
		t.Fatal("different content has the same digest")
	}
}
