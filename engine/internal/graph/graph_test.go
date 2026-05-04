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
		"  - name: realizations",
		"    file: realizations.yaml",
		"    kind: realization",
		"    required: true",
		"  - name: feature_implementations",
		"    file: feature_implementations.yaml",
		"    kind: feature_implementation",
		"    required: true",
		"  - name: component_implementations",
		"    file: component_implementations.yaml",
		"    kind: component_implementation",
		"    required: true",
		"  - name: baselines",
		"    file: baselines.yaml",
		"    kind: baseline",
		"    required: false",
		"generation:",
		"  markdown_relationship_fields: true",
		"  markdown_baseline_fields: true",
		"  stable_sort: lexical_id",
	}, "\n"))
	writeGraphFile(t, dir, "realizations.yaml", strings.Join([]string{
		"kind: realization",
		"entries:",
		"  - source: product/alpha",
		"    targets:",
		"      - model/requirements/alpha-requirement",
	}, "\n"))
	writeGraphFile(t, dir, "feature_implementations.yaml", strings.Join([]string{
		"kind: feature_implementation",
		"entries:",
		"  - source: model/requirements/alpha-requirement",
		"    targets:",
		"      - model/features/alpha-feature",
	}, "\n"))
	writeGraphFile(t, dir, "component_implementations.yaml", "kind: component_implementation\nentries: []\n")
	writeGraphFile(t, dir, "baselines.yaml", strings.Join([]string{
		"kind: baseline",
		"entries:",
		"  - component: model/components/alpha-component",
		"    repo: host-repo",
		"    path: services/alpha",
		"    commit: 0123456789abcdef0123456789abcdef01234567",
	}, "\n"))

	g, err := Load(filepath.Join(dir, "graph.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	wantIDs := []string{
		"model/components/alpha-component",
		"model/features/alpha-feature",
		"model/requirements/alpha-requirement",
		"product/alpha",
	}
	gotIDs := g.NodeIDs()
	if strings.Join(gotIDs, ",") != strings.Join(wantIDs, ",") {
		t.Fatalf("NodeIDs() = %v, want %v", gotIDs, wantIDs)
	}
	if got := MarkdownPath("model/features/alpha-feature"); got != "model/features/alpha-feature.md" {
		t.Fatalf("MarkdownPath() = %q", got)
	}
}

func TestLoadRejectsOutOfOrderParts(t *testing.T) {
	dir := t.TempDir()
	writeGraphFile(t, dir, "graph.yaml", strings.Join([]string{
		"schema_version: 1",
		"node_id_format: repo_relative_markdown_path_without_extension",
		"parts:",
		"  - name: feature_implementations",
		"    file: feature_implementations.yaml",
		"    kind: feature_implementation",
		"    required: true",
		"  - name: realizations",
		"    file: realizations.yaml",
		"    kind: realization",
		"    required: true",
		"  - name: component_implementations",
		"    file: component_implementations.yaml",
		"    kind: component_implementation",
		"    required: true",
		"generation:",
		"  markdown_relationship_fields: true",
		"  markdown_baseline_fields: true",
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
		"  - name: realizations",
		"    file: realizations.yaml",
		"    kind: realization",
		"    required: true",
		"  - name: feature_implementations",
		"    file: feature_implementations.yaml",
		"    kind: feature_implementation",
		"    required: true",
		"  - name: component_implementations",
		"    file: component_implementations.yaml",
		"    kind: component_implementation",
		"    required: true",
		"generation:",
		"  markdown_relationship_fields: true",
		"  markdown_baseline_fields: true",
		"  stable_sort: lexical_id",
	}, "\n"))
	writeGraphFile(t, dir, "realizations.yaml", strings.Join([]string{
		"kind: realization",
		"entries:",
		"  - source: ./product/alpha.md",
		"    targets:",
		"      - model/requirements/alpha-requirement",
	}, "\n"))
	writeGraphFile(t, dir, "feature_implementations.yaml", "kind: feature_implementation\nentries: []\n")
	writeGraphFile(t, dir, "component_implementations.yaml", "kind: component_implementation\nentries: []\n")

	if _, err := Load(filepath.Join(dir, "graph.yaml")); err == nil || !strings.Contains(err.Error(), "must be normalized") {
		t.Fatalf("Load() error = %v, want normalization error", err)
	}
}

func TestNormalizeNodeID(t *testing.T) {
	got, err := NormalizeNodeID(`model\\features\\alpha.md`)
	if err != nil {
		t.Fatalf("NormalizeNodeID() error = %v", err)
	}
	if got != "model/features/alpha" {
		t.Fatalf("NormalizeNodeID() = %q, want %q", got, "model/features/alpha")
	}
}

func writeGraphFile(t *testing.T, dir string, name string, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
