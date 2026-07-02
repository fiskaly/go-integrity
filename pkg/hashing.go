package pkg

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"path"
	"regexp"
	"strings"
)

// Hashing processes Go source code and returns a deterministic SHA256 hash
// of its core implementation logic without "package" and "import" definitions
// and import paths are normalized.
func Hashing(src []byte) ([]byte, string, error) {
	var undefined string

	fset := token.NewFileSet()

	// parse the file. We omit parser.ParseComments so comments are completely dropped.
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		return nil, undefined, fmt.Errorf("failed to parse Go code: %w", err)
	}

	// map local package aliases/names to their canonical import paths
	localToPath := make(map[string]string)
	versionRegex := regexp.MustCompile(`^v\d+$`)

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			continue
		}
		for _, spec := range genDecl.Specs {
			importSpec := spec.(*ast.ImportSpec)
			pathVal := strings.Trim(importSpec.Path.Value, `"`)

			var localName string
			if importSpec.Name != nil {
				localName = importSpec.Name.Name
			} else {
				base := path.Base(pathVal)
				if versionRegex.MatchString(base) {
					base = path.Base(path.Dir(pathVal))
				}
				localName = base
			}
			localToPath[localName] = pathVal
		}
	}

	// normalize import aliases within the implementation code: every reference
	// to an imported package is rewritten to the fixed identifier "namespace".
	ast.Inspect(f, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		if _, exists := localToPath[ident.Name]; exists {
			ident.Name = "namespace"
		}
		return true
	})

	// filter out import declarations entirely from the final output list
	var filteredDecls []ast.Decl
	for _, decl := range f.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			continue
		}
		filteredDecls = append(filteredDecls, decl)
	}

	// print the remaining declaration blocks into a buffer
	var buf bytes.Buffer
	for _, decl := range filteredDecls {
		err := printer.Fprint(&buf, fset, decl)
		if err != nil {
			return nil, undefined, fmt.Errorf("failed to print AST node: %w", err)
		}
		buf.WriteByte('\n')
	}

	// sanitize the output buffer to remove all internal empty lines
	cleanedBytes := stripEmptyLines(buf.Bytes())

	// generate the SHA256 hash of the fully compacted implementation string
	hash := sha256.Sum256(cleanedBytes)

	return cleanedBytes, fmt.Sprintf("%x", hash), nil
}
