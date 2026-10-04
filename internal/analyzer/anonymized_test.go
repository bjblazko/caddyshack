package analyzer

import (
	"strings"
	"testing"
)

// Pre-anonymized logs drop the User-Agent header but keep the derived
// browser and OS names in request.browser / request.os.
func TestPreAnonymizedBrowserFields(t *testing.T) {
	line := func(browser, os string) string {
		return `{"ts":1773100800,"status":200,"request":{"remote_ip":"10.0.0.0","method":"GET","host":"example.com","uri":"/","browser":"` +
			browser + `","os":"` + os + `"}}` + "\n"
	}
	log := line("Firefox", "Linux") + line("Firefox", "Linux") + line("Bot", "Other") + line("curl", "Other") + line("Monitor", "Other")

	names := func(p FilterParams) map[string]int {
		got := map[string]int{}
		for _, b := range Analyze(strings.NewReader(log), p).Report.Browsers {
			got[b.Name] = b.Count
		}
		return got
	}
	all := names(FilterParams{})
	want := map[string]int{"Firefox": 2, "Bot": 1, "Script": 1, "Monitor": 1}
	for k, v := range want {
		if all[k] != v {
			t.Errorf("browsers = %v, want %v", all, want)
			break
		}
	}
	filtered := names(FilterParams{IgnoreBots: true, IgnoreMonitors: true})
	if filtered["Bot"] != 0 || filtered["Monitor"] != 0 || filtered["Firefox"] != 2 || filtered["Script"] != 1 {
		t.Errorf("with exclusions: %v", filtered)
	}
	os := Analyze(strings.NewReader(log), FilterParams{}).Report.OperatingSystems
	if len(os) == 0 || os[0].Name != "Other" && os[0].Name != "Linux" {
		t.Errorf("operating systems = %v", os)
	}
}

// A real User-Agent always wins over pre-derived fields.
func TestUserAgentWinsOverBrowserFields(t *testing.T) {
	log := `{"ts":1773100800,"status":200,"request":{"remote_ip":"10.0.0.1","host":"example.com","uri":"/","browser":"Firefox","os":"Linux",` +
		`"headers":{"User-Agent":["curl/8.0"]}}}` + "\n"
	b := Analyze(strings.NewReader(log), FilterParams{}).Report.Browsers
	if len(b) != 1 || b[0].Name != "Script" {
		t.Errorf("browsers = %v, want Script from the User-Agent", b)
	}
}
