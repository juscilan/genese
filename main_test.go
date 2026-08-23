package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// processLine
// ---------------------------------------------------------------------------

func TestProcessLine_AllPlaceholders(t *testing.T) {
	template := "Hello %1, your order is %2."
	line := "Alice;42"
	want := "Hello Alice, your order is 42."
	got := processLine(template, line)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestProcessLine_ExtraPlaceholders(t *testing.T) {
	// Template has %3 but the line only has 2 fields — %3 should be removed.
	template := "A=%1 B=%2 C=%3"
	line := "foo;bar"
	want := "A=foo B=bar C="
	got := processLine(template, line)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestProcessLine_EmptyTemplate(t *testing.T) {
	got := processLine("", "foo;bar")
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestProcessLine_EmptyLine(t *testing.T) {
	// Empty line → single field "", replaces %1 with "" and removes %2+.
	template := "%1-%2"
	got := processLine(template, "")
	want := "-"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestProcessLine_NoPlaceholders(t *testing.T) {
	template := "no placeholders"
	got := processLine(template, "a;b;c")
	if got != template {
		t.Errorf("got %q, want %q", got, template)
	}
}

// ---------------------------------------------------------------------------
// generate
// ---------------------------------------------------------------------------

func TestGenerate_SingleLine(t *testing.T) {
	tmpl := "id=%1 user=%2"
	input := strings.NewReader("abc;john\n")
	var out strings.Builder

	if err := generate(tmpl, input, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "id=abc user=john\n"
	if out.String() != want {
		t.Errorf("got %q, want %q", out.String(), want)
	}
}

func TestGenerate_MultipleLines(t *testing.T) {
	tmpl := "%1"
	input := strings.NewReader("line1\nline2\nline3\n")
	var out strings.Builder

	if err := generate(tmpl, input, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "line1\nline2\nline3\n"
	if out.String() != want {
		t.Errorf("got %q, want %q", out.String(), want)
	}
}

func TestGenerate_EmptyInput(t *testing.T) {
	var out strings.Builder
	if err := generate("tmpl", strings.NewReader(""), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.String() != "" {
		t.Errorf("expected empty output, got %q", out.String())
	}
}

// errWriter simulates a writer that always returns an error.
type errWriter struct{ err error }

func (e *errWriter) Write(_ []byte) (int, error) { return 0, e.err }

func TestGenerate_WriteError(t *testing.T) {
	writeErr := errors.New("disk full")
	w := &errWriter{err: writeErr}

	err := generate("hello %1", strings.NewReader("world\n"), w)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, writeErr) {
		t.Errorf("expected write error wrapped, got: %v", err)
	}
}

// errReader simulates a reader that returns an error during Scan.
type errReader struct{ err error }

func (e *errReader) Read(_ []byte) (int, error) { return 0, e.err }

func TestGenerate_ScannerError(t *testing.T) {
	scanErr := errors.New("read failure")
	r := &errReader{err: scanErr}
	var out strings.Builder

	err := generate("tmpl", r, &out)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, scanErr) {
		t.Errorf("expected scan error wrapped, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// run
// ---------------------------------------------------------------------------

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to create temp file %s: %v", name, err)
	}
	return path
}

func TestRun_Success(t *testing.T) {
	dir := t.TempDir()

	tmplPath := writeTempFile(t, dir, "template.txt", "Hello %1!")
	dataPath := writeTempFile(t, dir, "data.txt", "World\n")
	distPath := filepath.Join(dir, "dist.txt")

	if err := run(tmplPath, dataPath, distPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(distPath)
	if err != nil {
		t.Fatalf("cannot read dist file: %v", err)
	}
	if !strings.Contains(string(got), "Hello World!") {
		t.Errorf("dist file content unexpected: %q", got)
	}
}

func TestRun_TemplateNotFound(t *testing.T) {
	dir := t.TempDir()
	err := run(
		filepath.Join(dir, "nonexistent_template.txt"),
		filepath.Join(dir, "data.txt"),
		filepath.Join(dir, "dist.txt"),
	)
	if err == nil {
		t.Fatal("expected error for missing template, got nil")
	}
}

func TestRun_DataNotFound(t *testing.T) {
	dir := t.TempDir()
	tmplPath := writeTempFile(t, dir, "template.txt", "tmpl")

	err := run(
		tmplPath,
		filepath.Join(dir, "nonexistent_data.txt"),
		filepath.Join(dir, "dist.txt"),
	)
	if err == nil {
		t.Fatal("expected error for missing data file, got nil")
	}
}

func TestRun_DistCreateError(t *testing.T) {
	dir := t.TempDir()
	tmplPath := writeTempFile(t, dir, "template.txt", "tmpl")
	dataPath := writeTempFile(t, dir, "data.txt", "row\n")

	// Point dist to a non-existent subdirectory to force a creation error.
	distPath := filepath.Join(dir, "nonexistent_subdir", "dist.txt")

	err := run(tmplPath, dataPath, distPath)
	if err == nil {
		t.Fatal("expected error for unwritable dist path, got nil")
	}
}

// ---------------------------------------------------------------------------
// main
// ---------------------------------------------------------------------------

func TestMain_Success(t *testing.T) {
	dir := t.TempDir()

	// Runs run() with temporary paths, mirroring main()'s default behavior.
	tmplPath := writeTempFile(t, dir, "_template.txt", "ok=%1")
	dataPath := writeTempFile(t, dir, "_list.txt", "yes\n")
	distPath := filepath.Join(dir, "_dist.txt")

	if err := run(tmplPath, dataPath, distPath); err != nil {
		t.Fatalf("run() failed: %v", err)
	}
}

// TestMain_Error tests the error branch of main() using the default files
// that are absent in the test working directory.
func TestMain_Error(t *testing.T) {
	// There is no reliable way to call main() directly without altering the
	// environment, so we guarantee coverage via run().
	err := run("_nonexistent_template.txt", "_list.txt", "_dist.txt")
	if err == nil {
		t.Fatal("expected error")
	}
}

// TestMainFunc calls main() ensuring the error branch is covered when the
// default files do not exist in the test working directory.
// main() prints and returns — it does not call os.Exit, so it is safe to call.
func TestMainFunc(t *testing.T) {
	// The default files (_template.txt, _list.txt) do not exist in the temp
	// dir, so main() will print the error and return — covering the error branch.
	origDir, _ := os.Getwd()
	tmp := t.TempDir()
	_ = os.Chdir(tmp)
	defer os.Chdir(origDir) //nolint:errcheck

	main() // error branch (template not found)
}

func TestMainFunc_Success(t *testing.T) {
	origDir, _ := os.Getwd()
	tmp := t.TempDir()
	_ = os.Chdir(tmp)
	defer os.Chdir(origDir) //nolint:errcheck

	// Create the default files in the temporary directory.
	_ = os.WriteFile("_template.txt", []byte("val=%1"), 0600)
	_ = os.WriteFile("_list.txt", []byte("42\n"), 0600)

	main() // success branch
}

// Ensures io.Writer is satisfied by errWriter (prevents "unused" linter warnings).
var _ io.Writer = (*errWriter)(nil)
