package main_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestProductionPackagesDoNotImportDomainEntities(t *testing.T) {
	const forbiddenImport = "chronocone/backend/domain/entities"

	var violations []string
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch path {
			case ".git", "domain/entities", "serde/rdbms/gen":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imported := range file.Imports {
			importPath, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				return err
			}
			if importPath == forbiddenImport {
				violations = append(violations, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan production imports: %v", err)
	}

	if len(violations) > 0 {
		t.Fatalf("production packages must import chronocone/backend/domain aggregate roots, not %s: %s", forbiddenImport, strings.Join(violations, ", "))
	}
}

// This guard intentionally checks the conventional `domain.X` selector form.
// It complements the import-boundary test above; it is not a full Go type checker
// and will not catch aliased imports or indirect dispatch.
func TestProductionPackagesDoNotCallStandaloneOwnedChildHelpers(t *testing.T) {
	forbiddenSelectors := map[string]struct{}{
		"ListCategoryAttrs": {},
		"ListElementAttrs":  {},
	}

	var violations []string
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch path {
			case ".git", "serde/rdbms/gen":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.HasPrefix(path, "domain/") {
			return nil
		}

		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			ident, ok := selector.X.(*ast.Ident)
			if !ok || ident.Name != "domain" {
				return true
			}
			if _, forbidden := forbiddenSelectors[selector.Sel.Name]; forbidden {
				violations = append(violations, path+":"+selector.Sel.Name)
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("scan owned-child helper calls: %v", err)
	}

	if len(violations) > 0 {
		t.Fatalf("production packages must not call standalone owned-child helpers; use aggregate-root reads instead: %s", strings.Join(violations, ", "))
	}
}
