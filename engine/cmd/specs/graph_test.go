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
	for _, path := range []string{
		filepath.Join(specsDir, "model", "traceability"),
		filepath.Join(specsDir, "product"),
		filepath.Join(specsDir, "model", "requirements"),
		filepath.Join(specsDir, "model", "use-cases"),
		filepath.Join(specsDir, "model", "components"),
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
		`"derive_reqt_edge_count": 1`,
		`"refine_edge_count": 1`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestCmdGraphValidateRejectsMissingNodeFile(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	for _, path := range []string{
		filepath.Join(specsDir, "model", "traceability"),
		filepath.Join(specsDir, "product"),
		filepath.Join(specsDir, "model", "requirements"),
		filepath.Join(specsDir, "model", "use-cases"),
		filepath.Join(specsDir, "model", "components"),
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

	if err := os.Remove(filepath.Join(specsDir, "model", "use-cases", "alpha-feature.md")); err != nil {
		t.Fatal(err)
	}

	err = cmdGraphValidate(nil)
	if err == nil || !strings.Contains(err.Error(), `model/use-cases/alpha-feature`) {
		t.Fatalf("cmdGraphValidate() error = %v, want missing node file error", err)
	}
}

func TestCmdGraphImportMarkdownWritesCanonicalGraph(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	for _, path := range []string{
		filepath.Join(specsDir, "product"),
		filepath.Join(specsDir, "model", "requirements"),
		filepath.Join(specsDir, "model", "use-cases"),
		filepath.Join(specsDir, "model", "components"),
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
	if len(reloaded.Relations[graph.PartKindDeriveReqt]) != 1 || len(reloaded.Relations[graph.PartKindSatisfy]) != 1 || len(reloaded.Relations[graph.PartKindRefine]) != 1 {
		t.Fatalf("unexpected imported graph sizes: %+v", reloaded)
	}
}

func TestCmdGraphGenerateMarkdownUpdatesFiles(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	for _, path := range []string{
		filepath.Join(specsDir, "product"),
		filepath.Join(specsDir, "model", "requirements"),
		filepath.Join(specsDir, "model", "use-cases"),
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
		Relations: map[graph.PartKind][]graph.RelationEntry{
			graph.PartKindDeriveReqt: {{Source: "product/alpha", Targets: []string{"model/requirements/alpha-requirement"}}},
			graph.PartKindRefine:     {{Source: "model/requirements/alpha-requirement", Targets: []string{"model/use-cases/alpha-feature"}}},
			graph.PartKindSatisfy:    {{Source: "model/requirements/alpha-requirement", Targets: []string{"model/components/alpha-component"}}},
		},
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
	if !strings.Contains(string(requirementBody), "realises:\n    - ../../product/alpha.md") {
		t.Fatalf("missing generated realises field:\n%s", string(requirementBody))
	}
}

func TestCmdGraphRebuildCacheWritesSQLite(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	for _, path := range []string{
		filepath.Join(specsDir, "model", "traceability"),
		filepath.Join(specsDir, "product"),
		filepath.Join(specsDir, "model", "requirements"),
		filepath.Join(specsDir, "model", "use-cases"),
		filepath.Join(specsDir, "model", "components"),
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
	if edges != 3 {
		t.Fatalf("edges = %d, want 3", edges)
	}
}

func TestCmdGraphSaveRelationsWritesCanonicalRelations(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	for _, path := range []string{
		filepath.Join(specsDir, "model", "traceability"),
		filepath.Join(specsDir, "product"),
		filepath.Join(specsDir, "model", "requirements"),
		filepath.Join(specsDir, "model", "use-cases"),
		filepath.Join(specsDir, "model", "components"),
	} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := config.Save(filepath.Join(specsDir, config.FileName), &config.File{}); err != nil {
		t.Fatal(err)
	}
	writeGraphFixture(t, specsDir)
	inputPath := filepath.Join(dir, "relations.json")
	if err := os.WriteFile(inputPath, []byte(`{"edges":[{"source":"model/use-cases/alpha-feature","target":"model/requirements/alpha-requirement","kind":"refine"}]}`), 0o644); err != nil {
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

	if err := cmdGraphSaveRelations([]string{"--in", inputPath}); err != nil {
		t.Fatalf("cmdGraphSaveRelations() error = %v", err)
	}
	reloaded, err := graph.Load(filepath.Join(specsDir, "model", "traceability", "graph.yaml"))
	if err != nil {
		t.Fatalf("graph.Load() error = %v", err)
	}
	if len(reloaded.Relations[graph.PartKindDeriveReqt]) != 0 {
		t.Fatalf("len(Relations[deriveReqt]) = %d, want 0", len(reloaded.Relations[graph.PartKindDeriveReqt]))
	}
	if len(reloaded.Relations[graph.PartKindRefine]) != 1 {
		t.Fatalf("len(Relations[refine]) = %d, want 1", len(reloaded.Relations[graph.PartKindRefine]))
	}
	if reloaded.Relations[graph.PartKindRefine][0].Source != "model/requirements/alpha-requirement" || len(reloaded.Relations[graph.PartKindRefine][0].Targets) != 1 || reloaded.Relations[graph.PartKindRefine][0].Targets[0] != "model/use-cases/alpha-feature" {
		t.Fatalf("unexpected feature implementation entry: %+v", reloaded.Relations[graph.PartKindRefine][0])
	}
}

func TestCmdGraphSaveRelationsAllowsUnlinkedArtifactNode(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	for _, path := range []string{
		filepath.Join(specsDir, "model", "traceability"),
		filepath.Join(specsDir, "product"),
		filepath.Join(specsDir, "model", "requirements"),
		filepath.Join(specsDir, "model", "use-cases"),
		filepath.Join(specsDir, "model", "components"),
	} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := config.Save(filepath.Join(specsDir, config.FileName), &config.File{}); err != nil {
		t.Fatal(err)
	}
	writeGraphFixture(t, specsDir)
	if err := os.WriteFile(filepath.Join(specsDir, "model", "components", "beta-component.md"), []byte("# Beta Component\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	inputPath := filepath.Join(dir, "relations.json")
	if err := os.WriteFile(inputPath, []byte(`{"edges":[{"source":"model/components/beta-component","target":"model/requirements/alpha-requirement","kind":"satisfy"}]}`), 0o644); err != nil {
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

	if err := cmdGraphSaveRelations([]string{"--in", inputPath}); err != nil {
		t.Fatalf("cmdGraphSaveRelations() error = %v", err)
	}
	reloaded, err := graph.Load(filepath.Join(specsDir, "model", "traceability", "graph.yaml"))
	if err != nil {
		t.Fatalf("graph.Load() error = %v", err)
	}
	if len(reloaded.Relations[graph.PartKindSatisfy]) != 1 {
		t.Fatalf("len(Relations[satisfy]) = %d, want 1", len(reloaded.Relations[graph.PartKindSatisfy]))
	}
	if reloaded.Relations[graph.PartKindSatisfy][0].Source != "model/requirements/alpha-requirement" || len(reloaded.Relations[graph.PartKindSatisfy][0].Targets) != 1 || reloaded.Relations[graph.PartKindSatisfy][0].Targets[0] != "model/components/beta-component" {
		t.Fatalf("unexpected component implementation entry: %+v", reloaded.Relations[graph.PartKindSatisfy][0])
	}
}

func writeGraphFixture(t *testing.T, specsDir string) {
	t.Helper()
	traceabilityDir := filepath.Join(specsDir, "model", "traceability")
	artifacts := map[string]string{
		filepath.Join(specsDir, "product", "alpha.md"):                           "# Alpha\n",
		filepath.Join(specsDir, "model", "requirements", "alpha-requirement.md"): "# Alpha Requirement\n",
		filepath.Join(specsDir, "model", "use-cases", "alpha-feature.md"):        "# Alpha Feature\n",
		filepath.Join(specsDir, "model", "components", "alpha-component.md"):     "# Alpha Component\n",
	}
	for path, content := range artifacts {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string]string{
		"graph.yaml": strings.Join([]string{
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
		}, "\n"),
		"deriveReqt.yaml": strings.Join([]string{
			"kind: deriveReqt",
			"entries:",
			"  - source: product/alpha",
			"    targets:",
			"      - model/requirements/alpha-requirement",
		}, "\n"),
		"refinements.yaml": strings.Join([]string{
			"kind: refine",
			"entries:",
			"  - source: model/requirements/alpha-requirement",
			"    targets:",
			"      - model/use-cases/alpha-feature",
		}, "\n"),
		"satisfactions.yaml": strings.Join([]string{
			"kind: satisfy",
			"entries:",
			"  - source: model/requirements/alpha-requirement",
			"    targets:",
			"      - model/components/alpha-component",
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
			"---",
			"realised_by:",
			"    - ../model/requirements/alpha-requirement.md",
			"---",
			"",
			"# Alpha",
		}, "\n"),
		filepath.Join(specsDir, "model", "requirements", "alpha-requirement.md"): strings.Join([]string{
			"---",
			"implemented_by:",
			"    - ../use-cases/alpha-feature.md",
			"    - ../components/alpha-component.md",
			"---",
			"",
			"# Alpha Requirement",
		}, "\n"),
		filepath.Join(specsDir, "model", "use-cases", "alpha-feature.md"): strings.Join([]string{
			"---",
			"requirements:",
			"    - ../requirements/alpha-requirement.md",
			"---",
			"",
			"# Alpha Feature",
		}, "\n"),
		filepath.Join(specsDir, "model", "components", "alpha-component.md"): strings.Join([]string{
			"---",
			"requirements:",
			"    - ../requirements/alpha-requirement.md",
			"---",
			"",
			"# Alpha Component",
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
			"---",
			"status: Draft",
			"realised_by: []",
			"---",
			"",
			"# Alpha",
		}, "\n"),
		filepath.Join(specsDir, "model", "requirements", "alpha-requirement.md"): strings.Join([]string{
			"---",
			"status: Draft",
			"realises: []",
			"implemented_by: []",
			"---",
			"",
			"# Alpha Requirement",
		}, "\n"),
		filepath.Join(specsDir, "model", "use-cases", "alpha-feature.md"): strings.Join([]string{
			"---",
			"status: Draft",
			"requirements: []",
			"---",
			"",
			"# Alpha Feature",
		}, "\n"),
		filepath.Join(specsDir, "model", "components", "alpha-component.md"): strings.Join([]string{
			"---",
			"status: Draft",
			"requirements: []",
			"---",
			"",
			"# Alpha Component",
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
