package main

import (
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Color-Of-Code/specs-toolchain/engine/internal/config"
	"github.com/Color-Of-Code/specs-toolchain/engine/internal/graph"
)

func TestCmdGraphValidateJSON(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	if err := os.MkdirAll(filepath.Join(specsDir, "model", "traceability"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := config.Save(filepath.Join(specsDir, config.FileName), &config.File{
		Repos: map[string]string{"host-repo": "repos/host"},
	}); err != nil {
		t.Fatal(err)
	}
	writeGraphFixture(t, specsDir)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()
	if err := os.Chdir(specsDir); err != nil {
		t.Fatal(err)
	}

	stdout, restore := captureStdout(t)
	defer restore()

	if err := cmdGraphValidate([]string{"--json"}); err != nil {
		t.Fatalf("cmdGraphValidate() error = %v", err)
	}
	out := stdout()
	for _, want := range []string{
		`"manifest_path":`,
		`"node_count": 4`,
		`"realization_edge_count": 1`,
		`"feature_implementation_edge_count": 1`,
		`"baseline_count": 1`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestValidateBaselineReposRejectsUnknownRepo(t *testing.T) {
	g := &graph.Graph{
		Baselines: []graph.BaselineEntry{{
			Component: "model/components/alpha-component",
			Repo:      "missing-repo",
			Path:      "/",
			Commit:    "0123456789abcdef0123456789abcdef01234567",
		}},
	}
	err := validateBaselineRepos(g, map[string]string{"known": "repos/known"})
	if err == nil || !strings.Contains(err.Error(), `missing-repo`) {
		t.Fatalf("validateBaselineRepos() error = %v, want missing repo error", err)
	}
}

func TestCmdGraphImportMarkdownWritesCanonicalGraph(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	for _, path := range []string{
		filepath.Join(specsDir, "product"),
		filepath.Join(specsDir, "model", "requirements"),
		filepath.Join(specsDir, "model", "features"),
		filepath.Join(specsDir, "model", "components"),
		filepath.Join(specsDir, "model", "baselines"),
	} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := config.Save(filepath.Join(specsDir, config.FileName), &config.File{
		Repos: map[string]string{"host-repo": "repos/host"},
	}); err != nil {
		t.Fatal(err)
	}
	writeGraphImportFixture(t, specsDir)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()
	if err := os.Chdir(specsDir); err != nil {
		t.Fatal(err)
	}

	if err := cmdGraphImportMarkdown(nil); err != nil {
		t.Fatalf("cmdGraphImportMarkdown() error = %v", err)
	}
	reloaded, err := graph.Load(filepath.Join(specsDir, "model", "traceability", "graph.yaml"))
	if err != nil {
		t.Fatalf("graph.Load() error = %v", err)
	}
	if len(reloaded.Realizations) != 1 || len(reloaded.FeatureImplementations) != 1 || len(reloaded.ComponentImplementations) != 1 {
		t.Fatalf("unexpected imported graph sizes: %+v", reloaded)
	}
	if len(reloaded.Baselines) != 1 {
		t.Fatalf("len(Baselines) = %d, want 1", len(reloaded.Baselines))
	}
}

func TestCmdGraphGenerateMarkdownUpdatesFiles(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	for _, path := range []string{
		filepath.Join(specsDir, "product"),
		filepath.Join(specsDir, "model", "requirements"),
		filepath.Join(specsDir, "model", "features"),
		filepath.Join(specsDir, "model", "components"),
	} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := config.Save(filepath.Join(specsDir, config.FileName), &config.File{}); err != nil {
		t.Fatal(err)
	}
	writeGraphGenerateFixture(t, specsDir)
	graphData := &graph.Graph{
		Realizations:             []graph.RelationEntry{{Source: "product/alpha", Targets: []string{"model/requirements/alpha-requirement"}}},
		FeatureImplementations:   []graph.RelationEntry{{Source: "model/requirements/alpha-requirement", Targets: []string{"model/features/alpha-feature"}}},
		ComponentImplementations: []graph.RelationEntry{{Source: "model/requirements/alpha-requirement", Targets: []string{"model/components/alpha-component"}}},
	}
	if err := graph.Write(filepath.Join(specsDir, "model", "traceability", "graph.yaml"), graphData); err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()
	if err := os.Chdir(specsDir); err != nil {
		t.Fatal(err)
	}

	if err := cmdGraphGenerateMarkdown(nil); err != nil {
		t.Fatalf("cmdGraphGenerateMarkdown() error = %v", err)
	}
	requirementBody, err := os.ReadFile(filepath.Join(specsDir, "model", "requirements", "alpha-requirement.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(requirementBody), "| Realises | [Alpha](../../product/alpha.md) |") {
		t.Fatalf("missing generated Realises row:\n%s", string(requirementBody))
	}
}

func TestCmdGraphRebuildCacheWritesSQLite(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	if err := os.MkdirAll(filepath.Join(specsDir, "model", "traceability"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := config.Save(filepath.Join(specsDir, config.FileName), &config.File{
		Repos: map[string]string{"host-repo": "repos/host"},
	}); err != nil {
		t.Fatal(err)
	}
	writeGraphFixture(t, specsDir)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()
	if err := os.Chdir(specsDir); err != nil {
		t.Fatal(err)
	}

	if err := cmdGraphRebuildCache(nil); err != nil {
		t.Fatalf("cmdGraphRebuildCache() error = %v", err)
	}
	db, err := sql.Open("sqlite", filepath.Join(specsDir, ".specs-cache", "traceability.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var edges int
	if err := db.QueryRow(`SELECT COUNT(*) FROM edges`).Scan(&edges); err != nil {
		t.Fatal(err)
	}
	if edges != 2 {
		t.Fatalf("edges = %d, want 2", edges)
	}
	var baselines int
	if err := db.QueryRow(`SELECT COUNT(*) FROM baselines`).Scan(&baselines); err != nil {
		t.Fatal(err)
	}
	if baselines != 1 {
		t.Fatalf("baselines = %d, want 1", baselines)
	}
}

func writeGraphFixture(t *testing.T, specsDir string) {
	t.Helper()
	traceabilityDir := filepath.Join(specsDir, "model", "traceability")
	files := map[string]string{
		"graph.yaml": strings.Join([]string{
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
			"  - name: service_implementations",
			"    file: service_implementations.yaml",
			"    kind: service_implementation",
			"    required: true",
			"  - name: api_implementations",
			"    file: api_implementations.yaml",
			"    kind: api_implementation",
			"    required: true",
			"  - name: baselines",
			"    file: baselines.yaml",
			"    kind: baseline",
			"    required: false",
			"generation:",
			"  markdown_relationship_fields: true",
			"  markdown_baseline_fields: true",
			"  stable_sort: lexical_id",
		}, "\n"),
		"realizations.yaml": strings.Join([]string{
			"kind: realization",
			"entries:",
			"  - source: product/alpha",
			"    targets:",
			"      - model/requirements/alpha-requirement",
		}, "\n"),
		"feature_implementations.yaml": strings.Join([]string{
			"kind: feature_implementation",
			"entries:",
			"  - source: model/requirements/alpha-requirement",
			"    targets:",
			"      - model/features/alpha-feature",
		}, "\n"),
		"component_implementations.yaml": "kind: component_implementation\nentries: []\n",
		"service_implementations.yaml":   "kind: service_implementation\nentries: []\n",
		"api_implementations.yaml":       "kind: api_implementation\nentries: []\n",
		"baselines.yaml": strings.Join([]string{
			"kind: baseline",
			"entries:",
			"  - component: model/components/alpha-component",
			"    repo: host-repo",
			"    path: /",
			"    commit: 0123456789abcdef0123456789abcdef01234567",
		}, "\n"),
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(traceabilityDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func writeGraphImportFixture(t *testing.T, specsDir string) {
	t.Helper()
	files := map[string]string{
		filepath.Join(specsDir, "product", "alpha.md"): strings.Join([]string{
			"# Alpha",
			"",
			"| Field | Value |",
			"| ----- | ----- |",
			"| Realised By | [Alpha Requirement](../model/requirements/alpha-requirement.md) |",
		}, "\n"),
		filepath.Join(specsDir, "model", "requirements", "alpha-requirement.md"): strings.Join([]string{
			"# Alpha Requirement",
			"",
			"| Field | Value |",
			"| ----- | ----- |",
			"| Implemented By | [Alpha Feature](../features/alpha-feature.md), [Alpha Component](../components/alpha-component.md) |",
		}, "\n"),
		filepath.Join(specsDir, "model", "features", "alpha-feature.md"): strings.Join([]string{
			"# Alpha Feature",
			"",
			"## Requirements",
			"- [Alpha Requirement](../requirements/alpha-requirement.md)",
		}, "\n"),
		filepath.Join(specsDir, "model", "components", "alpha-component.md"): strings.Join([]string{
			"# Alpha Component",
			"",
			"| Field | Value |",
			"| ----- | ----- |",
			"| Requirements | [Alpha Requirement](../requirements/alpha-requirement.md) |",
		}, "\n"),
		filepath.Join(specsDir, "model", "baselines", "repo-baseline.md"): strings.Join([]string{
			"# Baselines",
			"",
			"## Components",
			"| Component | Repo | Path | SHA | Date |",
			"| --------- | ---- | ---- | --- | ---- |",
			"| [Alpha Component](../components/alpha-component.md) | `host-repo` | `/` | `0123456789abcdef0123456789abcdef01234567` | 2026-05-03 |",
		}, "\n"),
	}
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func writeGraphGenerateFixture(t *testing.T, specsDir string) {
	t.Helper()
	files := map[string]string{
		filepath.Join(specsDir, "product", "alpha.md"): strings.Join([]string{
			"# Alpha",
			"",
			"| Field | Value |",
			"| ----- | ----- |",
			"| Status | Draft |",
			"| Realised By | — |",
		}, "\n"),
		filepath.Join(specsDir, "model", "requirements", "alpha-requirement.md"): strings.Join([]string{
			"# Alpha Requirement",
			"",
			"| Field | Value |",
			"| ----- | ----- |",
			"| Status | Draft |",
			"| Implemented By | — |",
		}, "\n"),
		filepath.Join(specsDir, "model", "features", "alpha-feature.md"): strings.Join([]string{
			"# Alpha Feature",
			"",
			"| Field | Value |",
			"| ----- | ----- |",
			"| Status | Draft |",
			"| Requirements | — |",
		}, "\n"),
		filepath.Join(specsDir, "model", "components", "alpha-component.md"): strings.Join([]string{
			"# Alpha Component",
			"",
			"| Field | Value |",
			"| ----- | ----- |",
			"| Status | Draft |",
			"| Requirements | — |",
			"| Baseline | — |",
		}, "\n"),
	}
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func captureStdout(t *testing.T) (func() string, func()) {
	t.Helper()
	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = writer

	return func() string {
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			data, err := io.ReadAll(reader)
			if err != nil {
				t.Fatal(err)
			}
			return string(data)
		}, func() {
			os.Stdout = original
			_ = reader.Close()
			_ = writer.Close()
		}
}
