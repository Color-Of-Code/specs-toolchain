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
		filepath.Join(modelDir, "use-cases"),
		filepath.Join(modelDir, "components"),
		filepath.Join(modelDir, "baselines"),
		productDir,
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeImportFile(t, filepath.Join(productDir, "alpha.md"), strings.Join([]string{
		"---",
		"realised_by:",
		"    - ../model/requirements/alpha-requirement.md",
		"---",
		"",
		"# Alpha",
	}, "\n"))
	writeImportFile(t, filepath.Join(modelDir, "requirements", "alpha-requirement.md"), strings.Join([]string{
		"---",
		"realises:",
		"    - ../../product/alpha.md",
		"implemented_by:",
		"    - ../use-cases/alpha-feature.md",
		"    - ../components/alpha-component.md",
		"---",
		"",
		"# Alpha Requirement",
	}, "\n"))
	writeImportFile(t, filepath.Join(modelDir, "use-cases", "alpha-feature.md"), strings.Join([]string{
		"---",
		"requirements:",
		"    - ../requirements/alpha-requirement.md",
		"---",
		"",
		"# Alpha Feature",
	}, "\n"))
	writeImportFile(t, filepath.Join(modelDir, "components", "alpha-component.md"), strings.Join([]string{
		"---",
		"requirements:",
		"    - ../requirements/alpha-requirement.md",
		"---",
		"",
		"# Alpha Component",
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
	if got := len(g.DeriveReqt); got != 1 {
		t.Fatalf("len(DeriveReqt) = %d, want 1", got)
	}
	if got := len(g.Satisfactions); got != 1 {
		t.Fatalf("len(Satisfactions) = %d, want 1", got)
	}
	if got := len(g.Refinements); got != 1 {
		t.Fatalf("len(Refinements) = %d, want 1", got)
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
	if len(reloaded.Manifest.Parts) != 4 {
		t.Fatalf("len(Manifest.Parts) = %d, want 4", len(reloaded.Manifest.Parts))
	}
}

func writeImportFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
