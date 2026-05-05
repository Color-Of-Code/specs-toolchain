package graph

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadValidGraph(t *testing.T) {
	dir := t.TempDir()
	writeGraphFile(t, dir, "graph.yaml", strings.Join([]string{
		"schema_version: 1",
		"node_id_format: repo_relative_markdown_path_without_extension",
		"parts:",
		"  - name: derive_reqt",
		"    file: deriveReqt.yaml",
		"    kind: deriveReqt",
		"    required: true",
		"  - name: refinements",
		"    file: refinements.yaml",
		"    kind: refine",
		"    required: true",
		"  - name: satisfactions",
		"    file: satisfactions.yaml",
		"    kind: satisfy",
		"    required: true",
		"generation:",
		"  markdown_relationship_fields: true",
		"  stable_sort: lexical_id",
	}, "\n"))
	writeGraphFile(t, dir, "deriveReqt.yaml", strings.Join([]string{
		"kind: deriveReqt",
		"entries:",
		"  - source: product/alpha",
		"    targets:",
		"      - model/requirements/alpha-requirement",
	}, "\n"))
	writeGraphFile(t, dir, "refinements.yaml", strings.Join([]string{
		"kind: refine",
		"entries:",
		"  - source: model/requirements/alpha-requirement",
		"    targets:",
		"      - model/use-cases/alpha-feature",
	}, "\n"))
	writeGraphFile(t, dir, "satisfactions.yaml", strings.Join([]string{
		"kind: satisfy",
		"entries:",
		"  - source: model/requirements/alpha-requirement",
		"    targets:",
		"      - model/components/alpha-component",
	}, "\n"))

	g, err := Load(filepath.Join(dir, "graph.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	wantIDs := []string{
		"model/components/alpha-component",
		"model/requirements/alpha-requirement",
		"model/use-cases/alpha-feature",
		"product/alpha",
	}
	gotIDs := g.NodeIDs()
	if strings.Join(gotIDs, ",") != strings.Join(wantIDs, ",") {
		t.Fatalf("NodeIDs() = %v, want %v", gotIDs, wantIDs)
	}
	if got := MarkdownPath("model/use-cases/alpha-feature"); got != "model/use-cases/alpha-feature.md" {
		t.Fatalf("MarkdownPath() = %q", got)
	}
}

func TestLoadRejectsOutOfOrderParts(t *testing.T) {
	dir := t.TempDir()
	writeGraphFile(t, dir, "graph.yaml", strings.Join([]string{
		"schema_version: 1",
		"node_id_format: repo_relative_markdown_path_without_extension",
		"parts:",
		"  - kind: satisfy",
		"    required: true",
		"  - kind: deriveReqt",
		"    required: true",
		"  - kind: refine",
		"    required: true",
		"generation:",
		"  markdown_relationship_fields: true",
		"  stable_sort: lexical_id",
	}, "\n"))

	if _, err := Load(filepath.Join(dir, "graph.yaml")); err == nil || !strings.Contains(err.Error(), "out of order") {
		t.Fatalf("Load() error = %v, want out of order error", err)
	}
}

func TestLoadRejectsNonCanonicalNodeIDs(t *testing.T) {
	dir := t.TempDir()
	writeGraphFile(t, dir, "graph.yaml", strings.Join([]string{
		"schema_version: 1",
		"node_id_format: repo_relative_markdown_path_without_extension",
		"parts:",
		"  - kind: deriveReqt",
		"    required: true",
		"  - kind: refine",
		"    required: true",
		"  - kind: satisfy",
		"    required: true",
		"generation:",
		"  markdown_relationship_fields: true",
		"  stable_sort: lexical_id",
	}, "\n"))
	writeGraphFile(t, dir, "deriveReqt.yaml", strings.Join([]string{
		"kind: deriveReqt",
		"entries:",
		"  - source: ./product/alpha.md",
		"    targets:",
		"      - model/requirements/alpha-requirement",
	}, "\n"))
	writeGraphFile(t, dir, "satisfactions.yaml", "kind: satisfy\nentries: []\n")
	writeGraphFile(t, dir, "refinements.yaml", "kind: refine\nentries: []\n")

	if _, err := Load(filepath.Join(dir, "graph.yaml")); err == nil || !strings.Contains(err.Error(), "must be normalized") {
		t.Fatalf("Load() error = %v, want normalization error", err)
	}
}

func TestNormalizeNodeID(t *testing.T) {
	got, err := NormalizeNodeID(`model\\use-cases\\alpha.md`)
	if err != nil {
		t.Fatalf("NormalizeNodeID() error = %v", err)
	}
	if got != "model/use-cases/alpha" {
		t.Fatalf("NormalizeNodeID() = %q, want %q", got, "model/use-cases/alpha")
	}
}

func writeGraphFile(t *testing.T, dir string, name string, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
