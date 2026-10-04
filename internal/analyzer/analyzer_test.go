package analyzer

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "rewrite golden files")

var fixtures = map[string]string{
	"sample": "../../testdata/sample-example.jsonl",
	"mixed":  "testdata/mixed.jsonl",
}

var filterCases = map[string]FilterParams{
	"none":          {},
	"host":          {Host: "a.example"},
	"unknown-host":  {Host: "nope.example"},
	"dates":         {StartDate: "2026-03-11", EndDate: "2026-03-12"},
	"success":       {Status: "success"},
	"error":         {Status: "error"},
	"browser-os":    {Browser: "Firefox", OS: "Linux"},
	"page":          {Page: "/"},
	"method":        {Method: "HEAD"},
	"ignore-static": {IgnoreStatic: true},
	"ignore-images": {IgnoreImages: true},
	"ignore-both":   {IgnoreStatic: true, IgnoreImages: true},
	"search-uri":    {Search: "/blog*"},
	"search-ip":     {Search: "10.0.0.*"},
	"search-ref":    {Search: "*search.example*"},
	"search-exact":  {Search: "/MISSING"},
	"host-and-err":  {Host: "c.example", Status: "error"},
}

func openFixture(t *testing.T, path string) *os.File {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close() })
	return f
}

func assertGolden(t *testing.T, name string, got any) {
	t.Helper()
	data, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join("testdata", "golden", name+".json")
	if *update {
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("missing golden file (run with -update): %v", err)
	}
	if !bytes.Equal(want, data) {
		t.Errorf("%s differs from golden file %s", name, path)
	}
}

func TestAnalyzeGolden(t *testing.T) {
	for fx, path := range fixtures {
		for name, params := range filterCases {
			t.Run(fx+"/"+name, func(t *testing.T) {
				assertGolden(t, "analyze-"+fx+"-"+name, Analyze(openFixture(t, path), params))
			})
		}
	}
}

func TestListEventsGolden(t *testing.T) {
	pages := map[string][2]int{
		"first":    {0, 3},
		"default":  {0, 0},
		"tail":     {8, 5},
		"past-end": {100, 10},
		"too-big":  {0, 500},
		"negative": {-5, 3},
	}
	for name, page := range pages {
		t.Run("mixed/"+name, func(t *testing.T) {
			got := ListEvents(openFixture(t, fixtures["mixed"]), FilterParams{}, page[0], page[1])
			assertGolden(t, "events-mixed-"+name, got)
		})
	}
	for name, params := range filterCases {
		t.Run("sample/"+name, func(t *testing.T) {
			got := ListEvents(openFixture(t, fixtures["sample"]), params, 0, 20)
			assertGolden(t, "events-sample-"+name, got)
		})
	}
}

func TestMatchGlob(t *testing.T) {
	cases := []struct {
		pattern, text string
		want          bool
	}{
		{"/blog", "/blog", true},
		{"/blog", "/blog/x", false},
		{"/BLOG", "/blog", true},
		{"/blog*", "/blog/x", true},
		{"*.png", "/a/b.png", true},
		{"*.png", "/a/b.png?v=1", false},
		{"*foo*", "xxfooyy", true},
		{"a*b*c", "aXbYc", true},
		{"a*b*c", "aXcYb", false},
		{"ab*ba", "aba", false},
		{"*", "", true},
		{"**", "anything", true},
	}
	for _, c := range cases {
		if got := matchGlob(c.pattern, c.text); got != c.want {
			t.Errorf("matchGlob(%q, %q) = %v, want %v", c.pattern, c.text, got, c.want)
		}
	}
}

func TestResourceClassifiers(t *testing.T) {
	cases := []struct {
		uri                  string
		asset, static, image bool
	}{
		{"/", false, false, false},
		{"/css/site.css", true, true, false},
		{"/app.js?v=3", false, true, false},
		{"/img/logo.png", true, false, true},
		{"/photo.jpeg", false, false, true},
		{"/robots.txt", false, true, false},
		{"/sub/sitemap.xml", false, true, false},
		{"/api/x", true, false, false},
		{"/fonts/a.woff2", true, true, false},
	}
	for _, c := range cases {
		if got := isAsset(c.uri); got != c.asset {
			t.Errorf("isAsset(%q) = %v, want %v", c.uri, got, c.asset)
		}
		if got := isStaticResource(c.uri); got != c.static {
			t.Errorf("isStaticResource(%q) = %v, want %v", c.uri, got, c.static)
		}
		if got := isImageResource(c.uri); got != c.image {
			t.Errorf("isImageResource(%q) = %v, want %v", c.uri, got, c.image)
		}
	}
}
