package analyzer

import (
	"strings"
	"testing"
)

func TestCanonicalHost(t *testing.T) {
	cases := map[string]string{
		"huepattl.de":        "huepattl.de",
		"huepattl.de:443":    "huepattl.de",
		"huepattl.de:80":     "huepattl.de",
		"HuePattl.DE:443":    "huepattl.de",
		"huepattl.de:8443":   "huepattl.de:8443",
		"[2001:db8::1]:443":  "[2001:db8::1]",
		"[2001:db8::1]":      "[2001:db8::1]",
		"[2001:db8::1]:8080": "[2001:db8::1]:8080",
		"":                   "",
	}
	for in, want := range cases {
		if got := canonicalHost(in); got != want {
			t.Errorf("canonicalHost(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestAnalyzeMergesDefaultPortHosts(t *testing.T) {
	line := func(host string) string {
		return `{"ts":1773100800,"status":200,"request":{"remote_ip":"10.0.0.1","method":"GET","host":"` + host + `","uri":"/"}}` + "\n"
	}
	log := line("huepattl.de") + line("huepattl.de:443") + line("HUEPATTL.de") + line("other.de:8443")

	res := Analyze(strings.NewReader(log), FilterParams{})
	if got := strings.Join(res.Hosts, ","); got != "huepattl.de,other.de:8443" {
		t.Errorf("hosts = %s", got)
	}

	filtered := Analyze(strings.NewReader(log), FilterParams{Host: "huepattl.de:443"})
	if filtered.Report.TotalRequests != 3 {
		t.Errorf("host filter matched %d requests, want 3", filtered.Report.TotalRequests)
	}

	events := ListEvents(strings.NewReader(log), FilterParams{Host: "HuePattl.de:443"}, 0, 10)
	for _, e := range events.Events {
		if e.Host != "huepattl.de" {
			t.Errorf("event host = %q, want canonical form", e.Host)
		}
	}
	if events.Total != 3 {
		t.Errorf("events total = %d, want 3", events.Total)
	}
}
