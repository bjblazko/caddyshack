package analyzer

import (
	"strings"
	"testing"
)

func TestTopReferrersExcludeOwnSites(t *testing.T) {
	entry := func(host, referer string) string {
		return `{"ts":1773100800,"status":200,"request":{"remote_ip":"10.0.0.1","method":"GET","host":"` + host +
			`","uri":"/","headers":{"Referer":["` + referer + `"]}}}` + "\n"
	}
	log := entry("example.com", "https://example.com/about") + // same site
		entry("example.com", "https://blog.example.org:443/post") + // another site in the log
		entry("blog.example.org", "https://www.search.test/?q=x") + // external
		entry("blog.example.org", "https://www.search.test/?q=x") +
		entry("example.com", "http://news.test/item") + // external
		entry("example.com", "not a url") + // unparseable counts as external
		entry("", "https://www.search.test/?q=x") // entry without host is no site

	names := func(p FilterParams) string {
		var got []string
		for _, r := range Analyze(strings.NewReader(log), p).Report.TopReferrers {
			got = append(got, r.Name)
		}
		return strings.Join(got, " | ")
	}

	want := "https://www.search.test/?q=x | http://news.test/item | not a url"
	if got := names(FilterParams{}); got != want {
		t.Errorf("referrers = %q, want %q", got, want)
	}
	// Own sites come from the whole log, not just the filtered entries:
	// blog.example.org stays internal while filtering on example.com.
	if got := names(FilterParams{Host: "example.com"}); got != "http://news.test/item | not a url" {
		t.Errorf("host-filtered referrers = %q", got)
	}
}

func TestReferrerHost(t *testing.T) {
	cases := map[string]string{
		"https://Example.com:443/a?b": "example.com",
		"http://example.com:8080/":    "example.com:8080",
		"android-app://com.x/":        "com.x",
		"not a url":                   "",
		"":                            "",
	}
	for in, want := range cases {
		if got := referrerHost(in); got != want {
			t.Errorf("referrerHost(%q) = %q, want %q", in, got, want)
		}
	}
}
