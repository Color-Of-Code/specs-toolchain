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
		{path: "/", contentType: "text/html", contains: []string{"Specs: Traceability", "/graph.json", "/assets/traceability-view.js"}},
		{path: "/graph.json", contentType: "application/json", contains: []string{`"id": "product/alpha"`, `"source": "product/alpha"`}},
		{path: "/graph.dot", contentType: "text/vnd.graphviz", contains: []string{"digraph traceability", "nproduct_alpha -> nmodel_requirements_alpha_requirement;"}},
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

func TestResolveArtifactPathRejectsEscapes(t *testing.T) {
	specsDir := t.TempDir()
	_, _, err := resolveArtifactPath(specsDir, "../outside.md")
	if err == nil || !strings.Contains(err.Error(), "escapes") {
		t.Fatalf("resolveArtifactPath() error = %v, want escape rejection", err)
	}
}
