package main

import (
	"encoding/json"
	"io/fs"
	"strings"
	"testing"
)

// credits is static/licenses/credits.json, which the "Licenses and thanks"
// page lists. The ISC, BSD and OFL licenses of what is built in require their
// texts to travel with every copy, so they are embedded beside it.
type credits struct {
	Inside []credit `json:"inside"`
	Data   []credit `json:"data"`
}

type credit struct {
	Name    string   `json:"name"`
	License string   `json:"license"`
	Home    string   `json:"home"`
	Text    string   `json:"text"`
	Files   []string `json:"files"`
}

func readCredits(t *testing.T) credits {
	t.Helper()
	data, err := fs.ReadFile(staticFiles, "static/licenses/credits.json")
	if err != nil {
		t.Fatalf("credits.json is not embedded: %v", err)
	}
	var c credits
	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatalf("credits.json: %v", err)
	}
	return c
}

// Every third-party file shipped in the binary must belong to a credit entry.
func TestThirdPartyFilesAreCredited(t *testing.T) {
	credited := map[string]bool{}
	for _, c := range readCredits(t).Inside {
		for _, f := range c.Files {
			credited[f] = true
		}
	}
	for _, dir := range []string{"vendor", "data", "fonts"} {
		entries, err := fs.ReadDir(staticFiles, "static/"+dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			if name := dir + "/" + e.Name(); !credited[name] {
				t.Errorf("%s is shipped but not listed in static/licenses/credits.json", name)
			}
		}
	}
}

// Every built-in part names its license and website, and its license text is
// embedded; listed files must exist.
func TestBuiltInCreditsAreComplete(t *testing.T) {
	for _, c := range readCredits(t).Inside {
		if c.License == "" || c.Home == "" || c.Text == "" {
			t.Errorf("%s: license, home and text are required", c.Name)
			continue
		}
		if _, err := fs.Stat(staticFiles, "static"+c.Text); err != nil {
			t.Errorf("%s: license text %s is not embedded", c.Name, c.Text)
		}
		for _, f := range c.Files {
			if _, err := fs.Stat(staticFiles, "static/"+f); err != nil {
				t.Errorf("%s: listed file %s does not exist", c.Name, f)
			}
		}
	}
	for _, c := range readCredits(t).Data {
		if c.License == "" || c.Home == "" {
			t.Errorf("%s: license and home are required", c.Name)
		}
		if c.Text != "" && !strings.HasPrefix(c.Text, "https://") {
			if _, err := fs.Stat(staticFiles, "static"+c.Text); err != nil {
				t.Errorf("%s: license text %s is not embedded", c.Name, c.Text)
			}
		}
	}
}
