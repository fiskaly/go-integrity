package pkg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTempFile creates a file with the given content inside a per-test
// temporary directory and returns its path.
func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "fixture.txt")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write temp fixture: %v", err)
	}
	return p
}

func TestReadFirstLine(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "multi-line returns only first line",
			content: "first line\nsecond line\nthird line\n",
			want:    "first line",
		},
		{
			name:    "single line without trailing newline",
			content: "only line",
			want:    "only line",
		},
		{
			name:    "single line with trailing newline",
			content: "only line\n",
			want:    "only line",
		},
		{
			name:    "leading blank line is returned as empty first line",
			content: "\nsecond line\n",
			want:    "",
		},
		{
			name:    "hash-like first line",
			content: "7a874d8e721e57e16306d48cbd1b509cff2a0b4b21361bea2580c4fe70b67ef9\n---\nsource\n",
			want:    "7a874d8e721e57e16306d48cbd1b509cff2a0b4b21361bea2580c4fe70b67ef9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := writeTempFile(t, tt.content)
			got, err := ReadFirstLine(p)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ReadFirstLine() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReadFirstLineEmptyFile(t *testing.T) {
	p := writeTempFile(t, "")
	_, err := ReadFirstLine(p)
	if err == nil {
		t.Fatal("expected an error for an empty file, got nil")
	}
	if !strings.Contains(err.Error(), "file is empty") {
		t.Errorf("expected %q error, got: %v", "file is empty", err)
	}
}

func TestReadFirstLineMissingFile(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist.txt")
	_, err := ReadFirstLine(missing)
	if err == nil {
		t.Fatal("expected an error for a missing file, got nil")
	}
	if !strings.Contains(err.Error(), "failed to open file") {
		t.Errorf("expected %q error, got: %v", "failed to open file", err)
	}
}

// TestReadFirstLineScanError forces the scanner error branch: a single line
// larger than bufio.MaxScanTokenSize (64 KiB) with no newline makes
// bufio.Scanner fail with ErrTooLong.
func TestReadFirstLineScanError(t *testing.T) {
	oversized := strings.Repeat("x", 70*1024)
	p := writeTempFile(t, oversized)
	_, err := ReadFirstLine(p)
	if err == nil {
		t.Fatal("expected a scan error for an oversized line, got nil")
	}
	if !strings.Contains(err.Error(), "error reading file") {
		t.Errorf("expected %q error, got: %v", "error reading file", err)
	}
}

func TestStripEmptyLines(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "no empty lines unchanged",
			in:   "a\nb\nc\n",
			want: "a\nb\nc\n",
		},
		{
			name: "interior blank lines removed",
			in:   "a\n\nb\n\n\nc\n",
			want: "a\nb\nc\n",
		},
		{
			name: "leading and trailing blank lines removed",
			in:   "\n\na\nb\n\n",
			want: "a\nb\n",
		},
		{
			name: "whitespace-only lines removed",
			in:   "a\n   \n\t\nb\n",
			want: "a\nb\n",
		},
		{
			name: "content with leading whitespace preserved",
			in:   "\tindented\n\nplain\n",
			want: "\tindented\nplain\n",
		},
		{
			name: "empty input yields empty output",
			in:   "",
			want: "",
		},
		{
			name: "only blank lines yields empty output",
			in:   "\n  \n\t\n",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(stripEmptyLines([]byte(tt.in)))
			if got != tt.want {
				t.Errorf("stripEmptyLines(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
