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
		filepath.Join(modelDir, "features"),
		filepath.Join(modelDir, "components"),
		filepath.Join(modelDir, "services"),
		filepath.Join(modelDir, "apis"),
		productDir,
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeGenerateFile(t, filepath.Join(productDir, "alpha.md"), strings.Join([]string{
		"# Alpha",
		"",
		"| Field | Value |",
		"| ----- | ----- |",
		"| Status | Draft |",
		"| Realised By | — |",
	}, "\n"))
	writeGenerateFile(t, filepath.Join(modelDir, "requirements", "alpha-requirement.md"), strings.Join([]string{
		"# Alpha Requirement",
		"",
		"| Field | Value |",
		"| ----- | ----- |",
		"| Status | Draft |",
		"| Implemented By | — |",
	}, "\n"))
	writeGenerateFile(t, filepath.Join(modelDir, "features", "alpha-feature.md"), strings.Join([]string{
		"# Alpha Feature",
		"",
		"| Field | Value |",
		"| ----- | ----- |",
		"| Status | Draft |",
		"| Requirements | — |",
	}, "\n"))
	writeGenerateFile(t, filepath.Join(modelDir, "components", "alpha-component.md"), strings.Join([]string{
		"# Alpha Component",
		"",
		"| Field | Value |",
		"| ----- | ----- |",
		"| Status | Draft |",
		"| Requirements | — |",
		"| Baseline | — |",
	}, "\n"))
	writeGenerateFile(t, filepath.Join(modelDir, "services", "alpha-service.md"), strings.Join([]string{
		"# Alpha Service",
		"",
		"| Field | Value |",
		"| ----- | ----- |",
		"| Status | Draft |",
		"| Requirements | — |",
	}, "\n"))
	writeGenerateFile(t, filepath.Join(modelDir, "apis", "alpha-api.md"), strings.Join([]string{
		"# Alpha API",
		"",
		"| Field | Value |",
		"| ----- | ----- |",
		"| Status | Draft |",
		"| Requirements | — |",
	}, "\n"))

	g := &Graph{
		Realizations:             []RelationEntry{{Source: "product/alpha", Targets: []string{"model/requirements/alpha-requirement"}}},
		FeatureImplementations:   []RelationEntry{{Source: "model/requirements/alpha-requirement", Targets: []string{"model/features/alpha-feature"}}},
		ComponentImplementations: []RelationEntry{{Source: "model/requirements/alpha-requirement", Targets: []string{"model/components/alpha-component"}}},
		ServiceImplementations:   []RelationEntry{{Source: "model/requirements/alpha-requirement", Targets: []string{"model/services/alpha-service"}}},
		APIImplementations:       []RelationEntry{{Source: "model/requirements/alpha-requirement", Targets: []string{"model/apis/alpha-api"}}},
		Baselines:                []BaselineEntry{{Component: "model/components/alpha-component", Repo: "host-repo", Path: "/", Commit: "0123456789abcdef0123456789abcdef01234567"}},
	}

	result, err := GenerateMarkdown(modelDir, productDir, g, false)
	if err != nil {
		t.Fatalf("GenerateMarkdown() error = %v", err)
	}
	if len(result.UpdatedFiles) != 6 {
		t.Fatalf("len(UpdatedFiles) = %d, want 6", len(result.UpdatedFiles))
	}

	requirementBody := readGenerateFile(t, filepath.Join(modelDir, "requirements", "alpha-requirement.md"))
	for _, want := range []string{
		"| Realises | [Alpha](../../product/alpha.md) |",
		"[Alpha API](../apis/alpha-api.md), [Alpha Component](../components/alpha-component.md), [Alpha Feature](../features/alpha-feature.md), [Alpha Service](../services/alpha-service.md)",
	} {
		if !strings.Contains(requirementBody, want) {
			t.Fatalf("requirement body missing %q:\n%s", want, requirementBody)
		}
	}
	componentBody := readGenerateFile(t, filepath.Join(modelDir, "components", "alpha-component.md"))
	if !strings.Contains(componentBody, "| Baseline | `host-repo:/@0123456789abcdef0123456789abcdef01234567` |") {
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
