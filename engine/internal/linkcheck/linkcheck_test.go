package linkcheck

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCheckBidirectional_Symmetric(t *testing.T) {
	model := t.TempDir()

	write(t, filepath.Join(model, "requirements", "core", "001-foo.md"), `# Foo

## Implemented By

- [feat](../../features/core/foo.md)
`)
	write(t, filepath.Join(model, "features", "core", "foo.md"), `# Foo Feature

## Requirements

- [REQ-001](../../requirements/core/001-foo.md)
`)

	var buf bytes.Buffer
	r := &Result{}
	CheckBidirectional(&buf, model, "", r)
	if r.Failed() {
		t.Fatalf("expected no errors, got %v\n%s", r.Errors, buf.String())
	}
}

func TestCheckBidirectional_MissingReverse(t *testing.T) {
	model := t.TempDir()

	write(t, filepath.Join(model, "requirements", "core", "001-foo.md"), `# Foo

## Implemented By

- [feat](../../features/core/foo.md)
`)
	// Feature file exists but has no `## Requirements` section.
	write(t, filepath.Join(model, "features", "core", "foo.md"), `# Foo Feature

## Notes

(nothing here)
`)

	var buf bytes.Buffer
	r := &Result{}
	CheckBidirectional(&buf, model, "", r)
	if !r.Failed() {
		t.Fatalf("expected error for missing reverse link\noutput: %s", buf.String())
	}
	joined := strings.Join(r.Errors, "\n")
	if !strings.Contains(joined, "foo.md") {
		t.Errorf("error message should mention the offending file: %s", joined)
	}
}

func TestCheckBidirectional_SkipsIndexAndDirs(t *testing.T) {
	model := t.TempDir()

	// Reference an _index.md (should be ignored as a target).
	write(t, filepath.Join(model, "requirements", "core", "001-foo.md"), `# Foo

## Implemented By

- [area](../../features/core/_index.md)
- [dir](../../features/core)
`)
	write(t, filepath.Join(model, "features", "core", "_index.md"), "# Index\n")

	var buf bytes.Buffer
	r := &Result{}
	CheckBidirectional(&buf, model, "", r)
	if r.Failed() {
		t.Fatalf("_index.md and directory targets should be skipped, got: %v", r.Errors)
	}
}

func TestCheckBidirectional_ProductAxis_Symmetric(t *testing.T) {
	root := t.TempDir()
	model := filepath.Join(root, "model")
	product := filepath.Join(root, "product")

	write(t, filepath.Join(product, "core", "001-login.md"), `# Login PR

## Realised By

- [REQ](../../model/requirements/core/001-login.md)
`)
	write(t, filepath.Join(model, "requirements", "core", "001-login.md"), `# Login req

## Realises

- [PR](../../../product/core/001-login.md)
`)

	var buf bytes.Buffer
	r := &Result{}
	CheckBidirectional(&buf, model, product, r)
	if r.Failed() {
		t.Fatalf("expected no errors, got %v\n%s", r.Errors, buf.String())
	}
}

func TestCheckBidirectional_ProductAxis_MissingRealises(t *testing.T) {
	root := t.TempDir()
	model := filepath.Join(root, "model")
	product := filepath.Join(root, "product")

	write(t, filepath.Join(product, "core", "001-login.md"), `# Login PR

## Realised By

- [REQ](../../model/requirements/core/001-login.md)
`)
	// MR exists but has no `## Realises` section.
	write(t, filepath.Join(model, "requirements", "core", "001-login.md"), `# Login req

## Description

just a req
`)

	var buf bytes.Buffer
	r := &Result{}
	CheckBidirectional(&buf, model, product, r)
	if !r.Failed() {
		t.Fatalf("expected error for missing Realises reverse")
	}
	if !strings.Contains(strings.Join(r.Errors, "\n"), "Realises") {
		t.Errorf("error should mention `Realises`: %v", r.Errors)
	}
}

func TestCheckBidirectional_ProductAxis_SkippedWhenAbsent(t *testing.T) {
	model := t.TempDir()
	// Plain bidirectional model still passes when productDir is empty.
	write(t, filepath.Join(model, "requirements", "core", "001-foo.md"), `# Foo

## Implemented By

- [feat](../../features/core/foo.md)
`)
	write(t, filepath.Join(model, "features", "core", "foo.md"), `# Foo Feature

## Requirements

- [REQ-001](../../requirements/core/001-foo.md)
`)

	var buf bytes.Buffer
	r := &Result{}
	CheckBidirectional(&buf, model, "", r)
	if r.Failed() {
		t.Fatalf("expected no errors with empty productDir, got %v", r.Errors)
	}
}
