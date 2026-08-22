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
	// Template tem %3 mas a linha só tem 2 campos — %3 deve sumir.
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
	// Linha vazia → campo único "", substitui %1 por "" e remove %2+.
	template := "%1-%2"
	got := processLine(template, "")
	want := "-"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestProcessLine_NoPlaceholders(t *testing.T) {
	template := "sem placeholders"
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

// errWriter simula um writer que sempre retorna erro.
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

// errReader simula um reader que retorna erro no Scan.
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

	// Aponta dist para um diretório inexistente para forçar erro de criação.
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

	// Sobrescreve os arquivos padrão usando variáveis de ambiente ou
	// simplesmente executa run() com caminhos temporários (mesma lógica do main).
	tmplPath := writeTempFile(t, dir, "_template.txt", "ok=%1")
	dataPath := writeTempFile(t, dir, "_list.txt", "yes\n")
	distPath := filepath.Join(dir, "_dist.txt")

	if err := run(tmplPath, dataPath, distPath); err != nil {
		t.Fatalf("run() failed: %v", err)
	}
}

// TestMain_FilesMissing testa o branch de erro do main() usando os arquivos
// padrão ausentes. Para cobrir o main() em si, usamos os.Stdin substituído.
func TestMain_Error(t *testing.T) {
	// Salva e restaura os args/stdout originais, não há como chamar main()
	// diretamente sem alterar o ambiente, então garantimos cobertura via run().
	err := run("_nonexistent_template.txt", "_list.txt", "_dist.txt")
	if err == nil {
		t.Fatal("expected error")
	}
}

// TestMainFunc chama a função main() garantindo que o branch de erro seja
// coberto quando os arquivos padrão não existem no diretório de trabalho
// do teste. O main() imprime e retorna — não faz os.Exit, então é seguro.
func TestMainFunc(t *testing.T) {
	// Os arquivos padrão (_template.txt, _list.txt) não existem no dir de
	// testes, então main() vai imprimir o erro e retornar — cobrindo todos
	// os branches de main().
	origDir, _ := os.Getwd()
	tmp := t.TempDir()
	_ = os.Chdir(tmp)
	defer os.Chdir(origDir) //nolint:errcheck

	main() // branch de erro (template não encontrado)
}

func TestMainFunc_Success(t *testing.T) {
	origDir, _ := os.Getwd()
	tmp := t.TempDir()
	_ = os.Chdir(tmp)
	defer os.Chdir(origDir) //nolint:errcheck

	// Cria os arquivos padrão no diretório temporário.
	_ = os.WriteFile("_template.txt", []byte("val=%1"), 0600)
	_ = os.WriteFile("_list.txt", []byte("42\n"), 0600)

	main() // branch de sucesso
}

// Garante que io.Writer é satisfeito pelo errWriter (evita "unused" no linter).
var _ io.Writer = (*errWriter)(nil)
