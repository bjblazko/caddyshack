package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/bjblazko/caddyshack/internal/analyzer"
)

const fixture = "../analyzer/testdata/mixed.jsonl"

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

func useTempDir(t *testing.T) {
	t.Helper()
	old := tempDir
	tempDir = t.TempDir()
	t.Cleanup(func() { tempDir = old })
}

func upload(t *testing.T) analyzer.AnalysisResult {
	t.Helper()
	return uploadWithQuery(t, "")
}

func uploadWithQuery(t *testing.T, query string) analyzer.AnalysisResult {
	t.Helper()
	data, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("logfile", "access.log")
	_, _ = fw.Write(data)
	_ = mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload"+query, &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	Upload(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("upload: status %d: %s", rec.Code, rec.Body)
	}
	var res analyzer.AnalysisResult
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	return res
}

func get(h http.HandlerFunc, url string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, url, nil))
	return rec
}

func TestUploadThenAnalyze(t *testing.T) {
	useTempDir(t)
	up := upload(t)
	if len(up.FileID) != 32 || up.Report.TotalRequests != 13 {
		t.Fatalf("unexpected upload result: id=%q total=%d", up.FileID, up.Report.TotalRequests)
	}

	rec := get(Analyze, "/api/analyze?file="+up.FileID+"&host=c.example&status=error")
	var res analyzer.AnalysisResult
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.FileID != "" || res.Report.TotalRequests != 1 || len(res.Hosts) != 3 {
		t.Errorf("filtered analyze: id=%q total=%d hosts=%v", res.FileID, res.Report.TotalRequests, res.Hosts)
	}
}

func TestEventsPagination(t *testing.T) {
	useTempDir(t)
	id := upload(t).FileID

	cases := []struct {
		query            string
		total, offset, n int
		limit            int
	}{
		{"&offset=2&limit=4", 13, 2, 4, 4},
		{"", 13, 0, 13, 100},
		{"&limit=abc&offset=xyz", 13, 0, 13, 100},
		{"&search=/blog*", 1, 0, 1, 100},
	}
	for _, c := range cases {
		rec := get(Events, "/api/events?file="+id+c.query)
		var res analyzer.EventsResult
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("%s: %v (%s)", c.query, err, rec.Body)
		}
		if res.Total != c.total || res.Offset != c.offset || len(res.Events) != c.n || res.Limit != c.limit {
			t.Errorf("%q: got total=%d offset=%d n=%d limit=%d", c.query, res.Total, res.Offset, len(res.Events), res.Limit)
		}
	}
}

func TestFileParamValidation(t *testing.T) {
	useTempDir(t)
	cases := map[string]int{
		"":                         http.StatusBadRequest,
		"?file=../etc":             http.StatusBadRequest,
		"?file=a/b":                http.StatusBadRequest,
		"?file=a.b":                http.StatusBadRequest,
		"?file=deadbeef":           http.StatusNotFound,
		"?name=../passwd":          http.StatusBadRequest,
		"?name=sub/access.log":     http.StatusBadRequest,
		"?name=does-not-exist.log": http.StatusNotFound,
	}
	for query, want := range cases {
		for name, h := range map[string]http.HandlerFunc{"analyze": Analyze, "events": Events} {
			if got := get(h, "/api/x"+query).Code; got != want {
				t.Errorf("%s %q: status %d, want %d", name, query, got, want)
			}
		}
	}
}

func TestUploadAppliesFilterParams(t *testing.T) {
	useTempDir(t)
	all := upload(t).Report.TotalRequests
	noBots := uploadWithQuery(t, "?ignore_bots=1&ignore_monitors=1").Report.TotalRequests
	if all != 13 || noBots != 11 {
		t.Errorf("total requests: all=%d (want 13), without bots=%d (want 11)", all, noBots)
	}
}
