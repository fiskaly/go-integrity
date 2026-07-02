# go-integrity

Deterministic integrity hashing for Go source files based on their core
implementation logic.

## Overview

`go-integrity` computes a stable SHA256 hash for one or more Go source files.
Instead of hashing raw bytes, it hashes the *semantic implementation* of the
code, so that changes to formatting, comments, package declarations, imports,
or import aliases do **not** change the resulting hash.

Two files that implement identical logic — but differ in whitespace, comments,
or the local names used for their imports — will produce the **same** hash.
This makes the tool useful for verifying that the meaningful behavior of a file
has not changed, independent of cosmetic edits.

## How It Works

For each input file, the `pkg.Hashing` function runs the following pipeline:

1. Parse the source into an AST, dropping all comments.
2. Normalize import identifiers: every reference to an imported package is
   rewritten to the fixed identifier `namespace` (so a call like `fmt.Println`
   becomes `namespace.Println`), meaning the local alias used for an import has
   no effect on the hash.
3. Strip `package` and `import` declarations entirely.
4. Remove all empty and whitespace-only lines.
5. Compute the SHA256 hash of the resulting compacted source.

## Installation

Install with `go install`:

```sh
go install github.com/fiskaly/go-integrity@latest
```

Or build from source:

```sh
git clone https://github.com/fiskaly/go-integrity.git
cd go-integrity
go build -o go-integrity .
```

## Usage

```
Usage: go-integrity [-h | -v | -s | -c] <file1.go> [file2.go ...]
```

| Flag | Description                                                        |
| ---- | ------------------------------------------------------------------ |
| `-h` | Print usage information.                                           |
| `-v` | Print the application version.                                     |
| `-s` | Include the normalized source of `<file.go>` in `<file.go-integrity>`. |
| `-c` | Verify that `<file.go>` still matches the hash stored in `<file.go-integrity>`. |

By default (no `-c`), for every input file the tool writes a sidecar file named
`<file.go>-integrity` containing the computed hash (and, with `-s`, the
normalized source used to derive it). It also prints each `<hash> <path>` pair
to stdout, sorted by file path.

### Verifying Integrity

With `-c`, the tool runs in **check mode** instead of writing sidecar files. For
each input file it recomputes the hash and compares it against the first line of
the existing `<file.go>-integrity` sidecar. This is useful in CI to confirm that
a file's meaningful implementation has not changed since its hash was recorded.

Each result is printed with a status marker:

| Marker | Meaning                                                        |
| ------ | -------------------------------------------------------------- |
| `🔄`   | The sidecar file was written/updated (default mode).           |
| `✅`   | Check mode: the file matches its recorded hash.                |
| `❌`   | Check mode: the file no longer matches its recorded hash.      |

In check mode the tool never modifies the sidecar files. If any file fails
verification, the process exits with a non-zero status code, so `-c` can be
wired directly into CI pipelines and pre-commit hooks.

### Examples

Hash a single file:

```sh
go-integrity pkg/hashing.go
```

```
cd1b2c3682ee2dbd113a4ba6f3f4353f4c3f4bb2d1984e1c4fc535c1b914c989 pkg/hashing.go
```

Hash multiple files and embed the normalized source in each sidecar:

```sh
go-integrity -s main.go pkg/hashing.go
```

Print the version:

```sh
go-integrity -v
```

Verify that files still match their recorded hashes:

```sh
go-integrity -c main.go pkg/hashing.go
```

```
✅ 7f3a...c989 main.go
❌ cd1b...4b2d pkg/hashing.go
```

## Docker

Build a minimal (`scratch`-based, non-root) container image using the provided
`Makefile`:

```sh
make build
```

Run it against files mounted into the container:

```sh
docker run --rm -v "$PWD":/src go-integrity:latest /src/main.go
```

## Requirements

- Go 1.25 or newer.
- No external dependencies.

## License

Copyright (C) 2026 fiskaly GmbH. Licensed under the Apache License 2.0.
See [LICENSE](LICENSE) for details.
