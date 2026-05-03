package graph

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportMarkdownAndWriteRoundTrip(t *testing.T) {
	root := t.TempDir()
	modelDir := filepath.Join(root, "model")
	productDir := filepath.Join(root, "product")
	baselineFile := filepath.Join(modelDir, "baselines", "repo-baseline.md")
	for _, dir := range []string{
		filepath.Join(modelDir, "requirements"),
		filepath.Join(modelDir, "features"),
		filepath.Join(modelDir, "components"),
		filepath.Join(modelDir, "baselines"),
		productDir,
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeImportFile(t, filepath.Join(productDir, "alpha.md"), strings.Join([]string{
		"# Alpha",
		"",
		"| Field | Value |",
		"| ----- | ----- |",
		"| Realised By | [Alpha Requirement](../model/requirements/alpha-requirement.md) |",
	}, "\n"))
	writeImportFile(t, filepath.Join(modelDir, "requirements", "alpha-requirement.md"), strings.Join([]string{
		"# Alpha Requirement",
		"",
		"| Field | Value |",
		"| ----- | ----- |",
		"| Realises | [Alpha](../../product/alpha.md) |",
		"| Implemented By | [Alpha Feature](../features/alpha-feature.md), [Alpha Component](../components/alpha-component.md) |",
	}, "\n"))
	writeImportFile(t, filepath.Join(modelDir, "features", "alpha-feature.md"), strings.Join([]string{
		"# Alpha Feature",
		"",
		"## Requirements",
		"- [Alpha Requirement](../requirements/alpha-requirement.md)",
	}, "\n"))
	writeImportFile(t, filepath.Join(modelDir, "components", "alpha-component.md"), strings.Join([]string{
		"# Alpha Component",
		"",
		"| Field | Value |",
		"| ----- | ----- |",
		"| Requirements | [Alpha Requirement](../requirements/alpha-requirement.md) |",
	}, "\n"))
	writeImportFile(t, baselineFile, strings.Join([]string{
		"# Baselines",
		"",
		"## Components",
		"| Component | Repo | Path | SHA | Date |",
		"| --------- | ---- | ---- | --- | ---- |",
		"| [Alpha Component](../components/alpha-component.md) | `host-repo` | `/` | `0123456789abcdef0123456789abcdef01234567` | 2026-05-03 |",
	}, "\n"))

	g, err := ImportMarkdown(modelDir, productDir, baselineFile)
	if err != nil {
		t.Fatalf("ImportMarkdown() error = %v", err)
	}
	if got := len(g.Realizations); got != 1 {
		t.Fatalf("len(Realizations) = %d, want 1", got)
	}
	if got := len(g.FeatureImplementations); got != 1 {
		t.Fatalf("len(FeatureImplementations) = %d, want 1", got)
	}
	if got := len(g.ComponentImplementations); got != 1 {
		t.Fatalf("len(ComponentImplementations) = %d, want 1", got)
	}
	if got := len(g.Baselines); got != 1 {
		t.Fatalf("len(Baselines) = %d, want 1", got)
	}

	manifestPath := filepath.Join(modelDir, "traceability", "graph.yaml")
	if err := Write(manifestPath, g); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	reloaded, err := Load(manifestPath)
	if err != nil {
		t.Fatalf("Load() after Write error = %v", err)
	}
	if got := len(reloaded.NodeIDs()); got != 4 {
		t.Fatalf("len(NodeIDs()) = %d, want 4", got)
	}
	if len(reloaded.Manifest.Parts) != 6 {
		t.Fatalf("len(Manifest.Parts) = %d, want 6", len(reloaded.Manifest.Parts))
	}
}

func writeImportFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}