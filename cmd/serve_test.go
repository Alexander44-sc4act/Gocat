package cmd

import (
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFileServerAddFileUsesPrefixAndStableMapping(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "tool.sh")
	if err := os.WriteFile(file, []byte("echo ok\n"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fs := NewFileServer("127.0.0.1", 8080, "tools")
	first := fs.AddFile(file)
	second := fs.AddFile(file)

	if first != "/tools/tool.sh" {
		t.Fatalf("AddFile() = %q, want /tools/tool.sh", first)
	}
	if second != first {
		t.Fatalf("AddFile() for same path = %q, want %q", second, first)
	}
}

func TestFileServerAddFileMakesDuplicateBasenamesUnique(t *testing.T) {
	dir := t.TempDir()
	aDir := filepath.Join(dir, "a")
	bDir := filepath.Join(dir, "b")
	if err := os.MkdirAll(aDir, 0o700); err != nil {
		t.Fatalf("MkdirAll a: %v", err)
	}
	if err := os.MkdirAll(bDir, 0o700); err != nil {
		t.Fatalf("MkdirAll b: %v", err)
	}
	a := filepath.Join(aDir, "same.txt")
	b := filepath.Join(bDir, "same.txt")
	if err := os.WriteFile(a, []byte("a"), 0o600); err != nil {
		t.Fatalf("WriteFile a: %v", err)
	}
	if err := os.WriteFile(b, []byte("b"), 0o600); err != nil {
		t.Fatalf("WriteFile b: %v", err)
	}

	fs := NewFileServer("127.0.0.1", 8080, "")
	first := fs.AddFile(a)
	second := fs.AddFile(b)

	if first != "/same.txt" {
		t.Fatalf("first AddFile() = %q, want /same.txt", first)
	}
	if second != "/same_1.txt" {
		t.Fatalf("second AddFile() = %q, want /same_1.txt", second)
	}
}

func TestAuthHandlerRequiresBasicAuth(t *testing.T) {
	fs := NewFileServer("127.0.0.1", 8080, "")
	fs.Auth = "user:pass"
	handler := authHandler(fs, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	unauthorized := httptest.NewRecorder()
	handler.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want %d", unauthorized.Code, http.StatusUnauthorized)
	}

	authorized := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("user", "pass")
	handler.ServeHTTP(authorized, req)
	if authorized.Code != http.StatusNoContent {
		t.Fatalf("authorized status = %d, want %d", authorized.Code, http.StatusNoContent)
	}
}

func TestHandleUploadPUTSanitizesFilename(t *testing.T) {
	dir := t.TempDir()
	req := httptest.NewRequest(http.MethodPut, "/upload?filename=../../owned.txt", strings.NewReader("hello"))
	rec := httptest.NewRecorder()

	handleUpload(rec, req, dir)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	got, err := os.ReadFile(filepath.Join(dir, "owned.txt"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("uploaded content = %q, want hello", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "..", "owned.txt")); !os.IsNotExist(err) {
		t.Fatalf("upload should not write outside upload dir; stat err = %v", err)
	}
}

func TestSanitizeUploadFilename(t *testing.T) {
	// On Windows, backslash is a separator so Base() yields the safe
	// basename "evil.exe"; on Unix the whole string is rejected.
	backslashWant := ""
	if runtime.GOOS == "windows" {
		backslashWant = "evil.exe"
	}
	cases := []struct {
		in   string
		want string
	}{
		{"tool.sh", "tool.sh"},
		{"../../owned.txt", "owned.txt"},
		{`..\..\evil.exe`, backslashWant},
		{"/etc/cron.d/x", "x"},
		{"", ""},
		{".", ""},
		{"..", ""},
		{"  spaced.txt  ", "spaced.txt"},
	}
	for _, tc := range cases {
		if got := sanitizeUploadFilename(tc.in); got != tc.want {
			t.Errorf("sanitizeUploadFilename(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestHandleUploadMultipartRejectsTraversal(t *testing.T) {
	dir := t.TempDir()

	var body strings.Builder
	w := multipart.NewWriter(&body)
	fw, err := w.CreateFormFile("file", "../../escape.txt")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	_, _ = fw.Write([]byte("pwned"))
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader(body.String()))
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()

	handleUpload(rec, req, dir)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "escape.txt")); err != nil {
		t.Fatalf("sanitized upload missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "..", "escape.txt")); !os.IsNotExist(err) {
		t.Fatalf("upload escaped upload dir; stat err = %v", err)
	}
}
