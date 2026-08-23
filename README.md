# genese ☕

**genese** is a lightweight CLI tool written in Go that generates text output by merging a template file with a CSV-like data file. Think of it as a minimal, fast template engine for repetitive text generation — great for generating SQL scripts, API payloads, config snippets, and more.

---

## How It Works

```
_template.txt  +  _list.txt  →  _dist.txt
```

1. **`_template.txt`** — defines the output structure using numbered placeholders (`%1`, `%2`, `%3`, …)
2. **`_list.txt`** — contains one entry per line, with fields separated by `;`
3. **`_dist.txt`** — the generated output, one rendered entry per line

For each line in `_list.txt`, genese replaces `%1` with the first field, `%2` with the second, and so on. Any unused placeholders are automatically removed.

---

## Example

**`_template.txt`**
```
template example %1 %2 %3
```

**`_list.txt`**
```
61980a3bf2673f61dd88a836;602b3067c3233546de26f79a;test0@gmail.com 
61980a3bf2673f61dd88a837;602b3067c3233546de26f79b;test1@gmail.com 
61980a3bf2673f61dd88a838;602b3067c3233546de26f79c;test2@gmail.com
```

**`_dist.txt`** (generated)
```
template example 61980a3bf2673f61dd88a836 602b3067c3233546de26f79a test0@gmail.com 
template example 61980a3bf2673f61dd88a837 602b3067c3233546de26f79b test1@gmail.com 
template example 61980a3bf2673f61dd88a838 602b3067c3233546de26f79c test2@gmail.com
```

---

## Getting Started

### Prerequisites

- [Go 1.22+](https://go.dev/dl/)

### Installation

```bash
git clone https://github.com/juscilan/genese.git
cd genese
go build -o genese .
```

### Usage

1. Create your `_template.txt` with `%1`, `%2`, … placeholders
2. Create your `_list.txt` with `;`-separated values, one entry per line
3. Run:

```bash
./genese
```

The output will be written to `_dist.txt` in the same directory.

---

## File Reference

| File | Description |
|---|---|
| `_template.txt` | Template with `%N` placeholders |
| `_list.txt` | Input data, fields separated by `;` |
| `_dist.txt` | Generated output (created/overwritten on each run) |

### Placeholder Rules

- Placeholders are numbered starting at `%1`
- Fields are split by `;`
- Extra placeholders (more than available fields) are silently removed
- Extra fields (more than placeholders in the template) are ignored

---

## Project Structure

```
genese/
├── main.go          # Core logic: processLine, generate, run, main
├── main_test.go     # Unit tests — 18 tests, 100% statement coverage
├── coverage.out     # Coverage profile (generated)
├── coverage.html    # HTML coverage report (generated)
├── _template.txt    # Example template
├── _list.txt        # Example input data
├── _dist.txt        # Generated output
└── go.mod
```

---

## Development

### Run tests

```bash
go test -v ./...
```

### Run tests with coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### Generate HTML coverage report

```bash
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

---

## Test Coverage

All 18 tests pass with **100% statement coverage** across every function.

| Test | Function | What it covers |
|---|---|---|
| `TestProcessLine_AllPlaceholders` | `processLine` | All placeholders replaced correctly |
| `TestProcessLine_ExtraPlaceholders` | `processLine` | Unused placeholders removed |
| `TestProcessLine_EmptyTemplate` | `processLine` | Empty template returns empty string |
| `TestProcessLine_EmptyLine` | `processLine` | Empty CSV line handled safely |
| `TestProcessLine_NoPlaceholders` | `processLine` | Template without placeholders is unchanged |
| `TestGenerate_SingleLine` | `generate` | Single line processed and written |
| `TestGenerate_MultipleLines` | `generate` | Multiple lines processed in order |
| `TestGenerate_EmptyInput` | `generate` | Empty reader produces no output |
| `TestGenerate_WriteError` | `generate` | Write error propagated correctly |
| `TestGenerate_ScannerError` | `generate` | Read error propagated correctly |
| `TestRun_Success` | `run` | Full happy path with temp files |
| `TestRun_TemplateNotFound` | `run` | Error when template file is missing |
| `TestRun_DataNotFound` | `run` | Error when data file is missing |
| `TestRun_DistCreateError` | `run` | Error when dist file cannot be created |
| `TestMain_Success` | `run` / `main` | run() with default-mirroring paths |
| `TestMain_Error` | `run` | run() error branch |
| `TestMainFunc` | `main` | main() error branch (template not found) |
| `TestMainFunc_Success` | `main` | main() success branch |

---

## License

MIT — feel free to use and adapt.
