package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestSite(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	html := filepath.Join(dir, "index.html")

	err := os.WriteFile(html, []byte(`<html><body><h1>Hello</h1></body></html>`), 0644)
	if err != nil {
		t.Fatalf("failed to write test html: %v", err)
	}

	return dir
}

func TestServeIndexHTML(t *testing.T) {
	root := setupTestSite(t)

	w := httptest.NewRecorder()

	path := filepath.Join(root, "index.html")
	serveFileWithInjectedReloadScript(path, w)

	res := w.Result()
	body := w.Body.String()

	if res.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	if !strings.Contains(body, `<script src="/reload.js"></script>`) {
		t.Error("expected injected reload <script> in response")
	}
}

func TestStaticFileFallback(t *testing.T) {
	dir := setupTestSite(t)

	err := os.WriteFile(filepath.Join(dir, "style.css"), []byte("body { color: red; }"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/style.css", nil)
	w := httptest.NewRecorder()
	handler := http.FileServer(http.Dir(dir))
	handler.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != 200 {
		t.Errorf("expected 200 OK for static file, got %d", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); !strings.Contains(ct, "text/css") {
		t.Errorf("expected text/css content type, got %s", ct)
	}
}
