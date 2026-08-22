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
await db.orders.updateOne({ _id: ObjectId("%1") }, { $set: { "status": "paid" } });
await db.users.updateOne({ _id: ObjectId('%2') }, { $set: { "progress.payments.status": "complete" } });
```

**`_list.txt`**
```
61980a3bf2673f61dd88a836;602b3067c3233546de26f79a
7a1bc4e20f931b44cc57d981;9f2e6a870c14e235bb91f304
```

**`_dist.txt`** (generated)
```
await db.orders.updateOne({ _id: ObjectId("61980a3bf2673f61dd88a836") }, { $set: { "status": "paid" } });
await db.users.updateOne({ _id: ObjectId('602b3067c3233546de26f79a') }, { $set: { "progress.payments.status": "complete" } });

await db.orders.updateOne({ _id: ObjectId("7a1bc4e20f931b44cc57d981") }, { $set: { "status": "paid" } });
await db.users.updateOne({ _id: ObjectId('9f2e6a870c14e235bb91f304') }, { $set: { "progress.payments.status": "complete" } });
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
├── main_test.go     # Unit tests (100% coverage)
├── _template.txt    # Example template
├── _list.txt        # Example data
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

> Current test coverage: **100%** across all statements.

---

## License

MIT — feel free to use and adapt.
