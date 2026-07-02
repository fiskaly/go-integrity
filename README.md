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
   rewritten to a deterministic, hash-based alias (`pkg_xxxxxxxx`) derived from
   the import path, so renaming an alias has no effect on the hash.
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
Usage: go-integrity [-h | -v | -s] <file1.go> [file2.go ...]
```

| Flag | Description                                                        |
| ---- | ------------------------------------------------------------------ |
| `-h` | Print usage information.                                           |
| `-v` | Print the application version.                                     |
| `-s` | Include the normalized source of `<file.go>` in `<file.go-integrity>`. |

For every input file, the tool writes a sidecar file named
`<file.go>-integrity` containing the computed hash (and, with `-s`, the
normalized source used to derive it). It also prints each `<hash> <path>` pair
to stdout, sorted by file path.

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
