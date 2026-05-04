package graph

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateMarkdownUpdatesArtifactFields(t *testing.T) {
	root := t.TempDir()
	modelDir := filepath.Join(root, "model")
	productDir := filepath.Join(root, "product")
	for _, dir := range []string{
		filepath.Join(modelDir, "requirements"),
		filepath.Join(modelDir, "use-cases"),
		filepath.Join(modelDir, "components"),
		productDir,
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeGenerateFile(t, filepath.Join(productDir, "alpha.md"), strings.Join([]string{
		"---",
		"status: Draft",
		"realised_by: []",
		"---",
		"",
		"# Alpha",
	}, "\n"))
	writeGenerateFile(t, filepath.Join(modelDir, "requirements", "alpha-requirement.md"), strings.Join([]string{
		"---",
		"status: Draft",
		"realises: []",
		"implemented_by: []",
		"---",
		"",
		"# Alpha Requirement",
	}, "\n"))
	writeGenerateFile(t, filepath.Join(modelDir, "use-cases", "alpha-feature.md"), strings.Join([]string{
		"---",
		"status: Draft",
		"requirements: []",
		"---",
		"",
		"# Alpha Feature",
	}, "\n"))
	writeGenerateFile(t, filepath.Join(modelDir, "components", "alpha-component.md"), strings.Join([]string{
		"---",
		"status: Draft",
		"requirements: []",
		"baseline: ~",
		"---",
		"",
		"# Alpha Component",
	}, "\n"))

	g := &Graph{
		DeriveReqt:    []RelationEntry{{Source: "product/alpha", Targets: []string{"model/requirements/alpha-requirement"}}},
		Refinements:   []RelationEntry{{Source: "model/requirements/alpha-requirement", Targets: []string{"model/use-cases/alpha-feature"}}},
		Satisfactions: []RelationEntry{{Source: "model/requirements/alpha-requirement", Targets: []string{"model/components/alpha-component"}}},
		Baselines:     []BaselineEntry{{Component: "model/components/alpha-component", Repo: "host-repo", Path: "/", Commit: "0123456789abcdef0123456789abcdef01234567"}},
	}

	result, err := GenerateMarkdown(modelDir, productDir, g, false)
	if err != nil {
		t.Fatalf("GenerateMarkdown() error = %v", err)
	}
	if len(result.UpdatedFiles) != 4 {
		t.Fatalf("len(UpdatedFiles) = %d, want 4", len(result.UpdatedFiles))
	}

	requirementBody := readGenerateFile(t, filepath.Join(modelDir, "requirements", "alpha-requirement.md"))
	for _, want := range []string{
		"realises:\n    - ../../product/alpha.md",
		"implemented_by:",
	} {
		if !strings.Contains(requirementBody, want) {
			t.Fatalf("requirement body missing %q:\n%s", want, requirementBody)
		}
	}
	componentBody := readGenerateFile(t, filepath.Join(modelDir, "components", "alpha-component.md"))
	if !strings.Contains(componentBody, "baseline: host-repo:/@0123456789abcdef0123456789abcdef01234567") {
		t.Fatalf("component baseline missing generated value:\n%s", componentBody)
	}
}

func writeGenerateFile(t *testing.T, path string, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readGenerateFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
