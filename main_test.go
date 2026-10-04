package main

import (
	"io/fs"
	"strings"
	"testing"
)

// Every third-party file shipped in the binary must be credited, with its
// license text, in static/licenses.txt.
func TestThirdPartyFilesAreCredited(t *testing.T) {
	notices, err := fs.ReadFile(staticFiles, "static/licenses.txt")
	if err != nil {
		t.Fatalf("licenses.txt is not embedded: %v", err)
	}
	for _, dir := range []string{"static/vendor", "static/data"} {
		entries, err := fs.ReadDir(staticFiles, dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			if !strings.Contains(string(notices), dir+"/"+e.Name()) {
				t.Errorf("%s/%s is shipped but not listed in static/licenses.txt", dir, e.Name())
			}
		}
	}
}
