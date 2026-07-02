package pkg

import (
	"strings"
	"testing"
)

// goldenSource is a small, representative Go file used across several tests.
// Its normalized implementation hash is pinned in TestHashingGoldenValue as a
// regression guard: if the hashing algorithm ever changes, that test fails.
const goldenSource = `package main

import "fmt"

// greet prints a greeting.
func greet(name string) {
	fmt.Println("hello", name)
}
`

// goldenHash is the SHA256 of the normalized implementation of goldenSource.
// It was produced by the current implementation and must only change on a
// deliberate, reviewed change to the hashing algorithm.
const goldenHash = "61654b64210f5c7ace6437aeac02e44d0ad07a65af07e3569ca11498c5b771be"

// mustHash is a helper that runs Hashing and fails the test on error.
func mustHash(t *testing.T, src string) (string, string) {
	t.Helper()
	content, hash, err := Hashing([]byte(src))
	if err != nil {
		t.Fatalf("Hashing returned unexpected error: %v", err)
	}
	return string(content), hash
}

// TestHashingGoldenValue pins a known input to a known hash so that any
// accidental change to the normalization/hashing pipeline is caught.
func TestHashingGoldenValue(t *testing.T) {
	_, hash := mustHash(t, goldenSource)
	if hash != goldenHash {
		t.Errorf("golden hash mismatch:\n got:  %s\n want: %s", hash, goldenHash)
	}
}

// TestHashingDeterministic ensures that hashing the same input twice yields
// the exact same result.
func TestHashingDeterministic(t *testing.T) {
	_, h1 := mustHash(t, goldenSource)
	_, h2 := mustHash(t, goldenSource)
	if h1 != h2 {
		t.Errorf("hash is not deterministic: %s != %s", h1, h2)
	}
}

// TestHashingEquivalence verifies that inputs which share the same semantic
// implementation - but differ in comments, formatting, package name, or import
// aliases - all produce the identical hash.
func TestHashingEquivalence(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{
			name: "baseline",
			src:  goldenSource,
		},
		{
			name: "comments dropped",
			src: `package main

import "fmt"

func greet(name string) {
	// this comment must not affect the hash
	fmt.Println("hello", name) // trailing comment
}
`,
		},
		{
			name: "reformatted with blank lines",
			src: `package main

import "fmt"



func greet(name string) {

	fmt.Println("hello", name)

}
`,
		},
		{
			name: "different package name",
			src: `package somethingelse

import "fmt"

func greet(name string) {
	fmt.Println("hello", name)
}
`,
		},
		{
			name: "different import alias",
			src: `package main

import f "fmt"

func greet(name string) {
	f.Println("hello", name)
}
`,
		},
		{
			name: "grouped import block",
			src: `package main

import (
	"fmt"
)

func greet(name string) {
	fmt.Println("hello", name)
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, hash := mustHash(t, tt.src)
			if hash != goldenHash {
				t.Errorf("expected equivalence with golden hash\n got:  %s\n want: %s", hash, goldenHash)
			}
		})
	}
}

// TestHashingSemanticChangeDiffers ensures a real change to the implementation
// produces a different hash.
func TestHashingSemanticChangeDiffers(t *testing.T) {
	changed := `package main

import "fmt"

func greet(name string) {
	fmt.Println("goodbye", name)
}
`
	_, hash := mustHash(t, changed)
	if hash == goldenHash {
		t.Errorf("semantic change did not alter the hash: %s", hash)
	}
}

// TestHashingImportAliasNormalized verifies that references to imported
// packages are rewritten to the fixed "namespace" identifier and that the
// original alias no longer appears in the normalized output.
func TestHashingImportAliasNormalized(t *testing.T) {
	src := `package main

import custom "fmt"

func greet(name string) {
	custom.Println("hello", name)
}
`
	content, _ := mustHash(t, src)

	if strings.Contains(content, "custom.") {
		t.Errorf("original import alias %q was not normalized away:\n%s", "custom", content)
	}
	if !strings.Contains(content, "namespace.") {
		t.Errorf("expected normalized alias %q in output:\n%s", "namespace.", content)
	}
}

// TestHashingVersionedImportPath exercises the branch that strips a trailing
// version segment (for example "/v2") from an import path when deriving the
// local package name.
func TestHashingVersionedImportPath(t *testing.T) {
	versioned := `package main

import "github.com/foo/bar/v2"

func use() {
	bar.Do()
}
`
	// The versioned import resolves to the local name "bar", which is then
	// rewritten to the fixed "namespace" identifier. The resulting hash must
	// be stable and non-empty.
	contentV, hashV := mustHash(t, versioned)

	if strings.Contains(contentV, "bar.") {
		t.Errorf("versioned import alias was not normalized:\n%s", contentV)
	}
	if !strings.Contains(contentV, "namespace.") {
		t.Errorf("expected normalized alias %q in versioned output:\n%s", "namespace.", contentV)
	}
	if hashV == "" {
		t.Error("expected a non-empty hash for versioned import source")
	}
}

// TestHashingStripsEmptyLines asserts the normalized content returned by
// Hashing contains no empty or whitespace-only lines.
func TestHashingStripsEmptyLines(t *testing.T) {
	src := `package main

import "fmt"


func a() {

	fmt.Println("a")

}


func b() {
	fmt.Println("b")
}
`
	content, _ := mustHash(t, src)

	for i, line := range strings.Split(content, "\n") {
		// The final split element is the empty string after the trailing
		// newline; that is expected and not an "internal" empty line.
		if i == len(strings.Split(content, "\n"))-1 && line == "" {
			continue
		}
		if strings.TrimSpace(line) == "" {
			t.Errorf("normalized output contains an empty/whitespace-only line at index %d:\n%q", i, content)
		}
	}
}

// TestHashingChainedSelector exercises the AST inspector branch where a
// selector expression's operand is itself a selector (for example
// "o.inner.Field") rather than a plain identifier, so it must not be treated
// as an import reference.
func TestHashingChainedSelector(t *testing.T) {
	src := `package main

type inner struct{ Field int }
type outer struct{ inner inner }

func get(o outer) int {
	return o.inner.Field
}
`
	content, hash := mustHash(t, src)
	if hash == "" {
		t.Error("expected a non-empty hash for chained selector source")
	}
	if strings.Contains(content, "namespace.") {
		t.Errorf("chained selector must not be rewritten as an import alias:\n%s", content)
	}
	if !strings.Contains(content, "o.inner.Field") {
		t.Errorf("expected chained selector preserved in output:\n%s", content)
	}
}

// TestHashingNoImports covers a valid file that has no import declarations.
func TestHashingNoImports(t *testing.T) {
	src := `package main

func add(a, b int) int {
	return a + b
}
`
	content, hash := mustHash(t, src)
	if hash == "" {
		t.Error("expected a non-empty hash for source without imports")
	}
	if !strings.Contains(content, "func add(a, b int) int") {
		t.Errorf("expected function body preserved in output:\n%s", content)
	}
}

// TestHashingMinimalFile covers the edge case of a file with only a package
// clause and no declarations.
func TestHashingMinimalFile(t *testing.T) {
	content, hash := mustHash(t, "package p\n")
	if hash == "" {
		t.Error("expected a non-empty hash for a minimal package-only file")
	}
	if strings.TrimSpace(content) != "" {
		t.Errorf("expected empty normalized content for package-only file, got:\n%q", content)
	}
}

// TestHashingInvalidSource verifies the error path when the input cannot be
// parsed as Go source.
func TestHashingInvalidSource(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{name: "not go at all", src: "this is not valid go source {{{"},
		{name: "missing package clause", src: "func main() {}"},
		{name: "empty input", src: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, hash, err := Hashing([]byte(tt.src))
			if err == nil {
				t.Fatalf("expected an error for invalid source, got hash %q", hash)
			}
			if hash != "" {
				t.Errorf("expected empty hash on error, got %q", hash)
			}
			if content != nil {
				t.Errorf("expected nil content on error, got %q", string(content))
			}
		})
	}
}
