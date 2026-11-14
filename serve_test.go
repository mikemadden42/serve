package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newTestHandler serves a temp directory containing regular files, a dotfile,
// and a dot directory, and captures the request log in the returned buffer.
func newTestHandler(t *testing.T) (http.Handler, *bytes.Buffer) {
	t.Helper()

	dir := t.TempDir()
	files := map[string]string{
		"a.txt":       "hello",
		"sub/b.txt":   "nested",
		".env":        "SECRET",
		".git/config": "cfg",
		"sub/.hidden": "hidden",
	}
	for name, content := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags)
	})

	return wrapHandler(http.FileServer(dotFileHidingFS{http.Dir(dir)})), &buf
}

func get(h http.Handler, target string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec
}

func TestServe(t *testing.T) {
	tests := []struct {
		target string
		status int
		body   string
	}{
		{"/a.txt", http.StatusOK, "hello"},
		{"/sub/b.txt", http.StatusOK, "nested"},
		{"/", http.StatusOK, ""},
		{"/missing.txt", http.StatusNotFound, ""},
		{"/.env", http.StatusNotFound, ""},
		{"/.git/config", http.StatusNotFound, ""},
		{"/sub/.hidden", http.StatusNotFound, ""},
		{"/sub/../.env", http.StatusNotFound, ""},
		{"/%2eenv", http.StatusNotFound, ""},
	}
	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			h, _ := newTestHandler(t)
			rec := get(h, tt.target)
			if rec.Code != tt.status {
				t.Errorf("status = %d, want %d", rec.Code, tt.status)
			}
			if tt.body != "" && rec.Body.String() != tt.body {
				t.Errorf("body = %q, want %q", rec.Body.String(), tt.body)
			}
			if strings.Contains(rec.Body.String(), "SECRET") {
				t.Error("response leaked dotfile contents")
			}
		})
	}
}

func TestDirectoryListingHidesDotfiles(t *testing.T) {
	h, _ := newTestHandler(t)

	for _, target := range []string{"/", "/sub/"} {
		body := get(h, target).Body.String()
		for _, name := range []string{".env", ".git", ".hidden"} {
			if strings.Contains(body, name) {
				t.Errorf("listing of %s contains %s:\n%s", target, name, body)
			}
		}
	}

	body := get(h, "/").Body.String()
	for _, name := range []string{"a.txt", "sub/"} {
		if !strings.Contains(body, name) {
			t.Errorf("listing of / missing %s:\n%s", name, body)
		}
	}
}

func TestLogStatus(t *testing.T) {
	tests := []struct {
		target string
		status string
	}{
		{"/", " 200 "},
		{"/a.txt", " 200 "},
		{"/missing.txt", " 404 "},
	}
	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			h, buf := newTestHandler(t)
			get(h, tt.target)
			if !strings.Contains(buf.String(), tt.status) {
				t.Errorf("log = %q, want status %q", buf.String(), tt.status)
			}
		})
	}
}

func TestLogQuotesPathAndUserAgent(t *testing.T) {
	h, buf := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/x%0Afake%20entry", nil)
	req.Header.Set("User-Agent", "agent\x1b[31m")
	h.ServeHTTP(httptest.NewRecorder(), req)

	out := buf.String()
	if n := strings.Count(out, "\n"); n != 1 {
		t.Errorf("log has %d lines, want 1: %q", n, out)
	}
	for _, want := range []string{`"/x\nfake entry"`, `"agent\x1b[31m"`} {
		if !strings.Contains(out, want) {
			t.Errorf("log = %q, want it to contain %s", out, want)
		}
	}
}

// readerFromRecorder is a ResponseRecorder that records whether ReadFrom was used.
type readerFromRecorder struct {
	*httptest.ResponseRecorder
	called bool
}

func (r *readerFromRecorder) ReadFrom(src io.Reader) (int64, error) {
	r.called = true
	return io.Copy(r.ResponseRecorder, src)
}

func TestReadFromPassesThrough(t *testing.T) {
	under := &readerFromRecorder{ResponseRecorder: httptest.NewRecorder()}
	srw := &statusRespWr{ResponseWriter: under, status: http.StatusOK}

	// Mirror http.ServeContent, which copies through io.CopyN (a LimitReader,
	// so the source's WriteTo can't short-circuit the writer's ReadFrom).
	if _, err := io.CopyN(srw, strings.NewReader("payload"), 7); err != nil {
		t.Fatal(err)
	}
	if !under.called {
		t.Error("underlying ReadFrom was not used")
	}
	if got := under.Body.String(); got != "payload" {
		t.Errorf("body = %q, want %q", got, "payload")
	}
	if srw.status != http.StatusOK {
		t.Errorf("status = %d, want %d", srw.status, http.StatusOK)
	}
}

func TestUnwrap(t *testing.T) {
	rec := httptest.NewRecorder()
	srw := &statusRespWr{ResponseWriter: rec}
	if srw.Unwrap() != rec {
		t.Error("Unwrap did not return the underlying writer")
	}
}
