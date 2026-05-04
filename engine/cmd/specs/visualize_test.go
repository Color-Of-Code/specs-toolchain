package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Color-Of-Code/specs-toolchain/engine/internal/config"
)

func TestNewTraceabilityUIHandlerServesUIAndArtifacts(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	for _, path := range []string{
		filepath.Join(specsDir, "model", "traceability"),
		filepath.Join(specsDir, "product"),
		filepath.Join(specsDir, "model", "requirements"),
		filepath.Join(specsDir, "model", "features"),
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

	handler, err := newTraceabilityUIHandler(specsDir)
	if err != nil {
		t.Fatalf("newTraceabilityUIHandler() error = %v", err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()

	for _, tc := range []struct {
		path        string
		contentType string
		contains    []string
	}{
		{path: "/", contentType: "text/html", contains: []string{"Specs: Traceability", "/graph.json", "/assets/traceability-view.js", "/relations", "layout-mode", "Relayout", "Remove Selected Edge", "Add Edge", "relation-kind", "component_implementation", "id=\"details\"", "No selection"}},
		{path: "/graph.json", contentType: "application/json", contains: []string{`"id": "product/alpha"`, `"source": "product/alpha"`}},
		{path: "/assets/traceability-view.js", contentType: "text/javascript", contains: []string{"window.TraceabilityUI", "cytoscape"}},
		{path: "/artifact?path=model/features/alpha-feature.md", contentType: "text/html", contains: []string{"alpha-feature.md", "# Alpha Feature"}},
	} {
		resp, err := http.Get(server.URL + tc.path)
		if err != nil {
			t.Fatalf("GET %s: %v", tc.path, err)
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			t.Fatalf("read body %s: %v", tc.path, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET %s status = %d, want 200\n%s", tc.path, resp.StatusCode, string(body))
		}
		if got := resp.Header.Get("Content-Type"); !strings.Contains(got, tc.contentType) {
			t.Fatalf("GET %s content-type = %q, want %q", tc.path, got, tc.contentType)
		}
		for _, want := range tc.contains {
			if !strings.Contains(string(body), want) {
				t.Fatalf("GET %s body missing %q\n%s", tc.path, want, string(body))
			}
		}
	}
}

func TestNewTraceabilityUIHandlerSavesRelations(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	for _, path := range []string{
		filepath.Join(specsDir, "model", "traceability"),
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
	writeGraphFixture(t, specsDir)

	handler, err := newTraceabilityUIHandler(specsDir)
	if err != nil {
		t.Fatalf("newTraceabilityUIHandler() error = %v", err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()

	payload := `{"edges":[{"source":"model/requirements/alpha-requirement","target":"model/features/alpha-feature","kind":"feature_implementation"}]}`
	resp, err := http.Post(server.URL+"/relations", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("POST /relations: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("read POST /relations body: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("POST /relations status = %d, want 204\n%s", resp.StatusCode, string(body))
	}

	relationsData, err := os.ReadFile(filepath.Join(specsDir, "model", "traceability", "realizations.yaml"))
	if err != nil {
		t.Fatalf("read realizations.yaml: %v", err)
	}
	if !strings.Contains(string(relationsData), "entries: []") {
		t.Fatalf("realizations.yaml should be empty after removal\n%s", string(relationsData))
	}

	resp, err = http.Get(server.URL + "/graph.json")
	if err != nil {
		t.Fatalf("GET /graph.json after relation save: %v", err)
	}
	body, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("read GET /graph.json after relation save: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /graph.json after relation save status = %d, want 200\n%s", resp.StatusCode, string(body))
	}
	if strings.Contains(string(body), `"source": "product/alpha"`) {
		t.Fatalf("graph.json still contains removed realization edge\n%s", string(body))
	}
	if !strings.Contains(string(body), `"source": "model/requirements/alpha-requirement"`) {
		t.Fatalf("graph.json should still contain feature implementation edge\n%s", string(body))
	}
}

func TestNewTraceabilityUIHandlerAddsRelations(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	for _, path := range []string{
		filepath.Join(specsDir, "model", "traceability"),
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
	writeGraphFixture(t, specsDir)

	handler, err := newTraceabilityUIHandler(specsDir)
	if err != nil {
		t.Fatalf("newTraceabilityUIHandler() error = %v", err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()

	payload := `{"edges":[{"source":"product/alpha","target":"model/requirements/alpha-requirement","kind":"realization"},{"source":"model/requirements/alpha-requirement","target":"model/features/alpha-feature","kind":"feature_implementation"},{"source":"model/requirements/alpha-requirement","target":"model/components/alpha-component","kind":"component_implementation"}]}`
	resp, err := http.Post(server.URL+"/relations", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("POST /relations add: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("read POST /relations add body: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("POST /relations add status = %d, want 204\n%s", resp.StatusCode, string(body))
	}

	componentData, err := os.ReadFile(filepath.Join(specsDir, "model", "traceability", "component_implementations.yaml"))
	if err != nil {
		t.Fatalf("read component_implementations.yaml: %v", err)
	}
	if !strings.Contains(string(componentData), "model/components/alpha-component") {
		t.Fatalf("component_implementations.yaml missing added component edge\n%s", string(componentData))
	}

	resp, err = http.Get(server.URL + "/graph.json")
	if err != nil {
		t.Fatalf("GET /graph.json after add: %v", err)
	}
	body, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("read GET /graph.json after add: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /graph.json after add status = %d, want 200\n%s", resp.StatusCode, string(body))
	}
	if !strings.Contains(string(body), `"target": "model/components/alpha-component"`) {
		t.Fatalf("graph.json missing added component edge\n%s", string(body))
	}
}

func TestResolveArtifactPathRejectsEscapes(t *testing.T) {
	specsDir := t.TempDir()
	_, _, err := resolveArtifactPath(specsDir, "../outside.md")
	if err == nil || !strings.Contains(err.Error(), "escapes") {
		t.Fatalf("resolveArtifactPath() error = %v, want escape rejection", err)
	}
}
