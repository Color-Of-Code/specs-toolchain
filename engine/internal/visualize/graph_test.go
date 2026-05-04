package visualize

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tracegraph "github.com/Color-Of-Code/specs-toolchain/engine/internal/graph"
)

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func sampleGraph(t *testing.T) (string, string, *tracegraph.Graph) {
	root := t.TempDir()
	model := filepath.Join(root, "model")
	product := filepath.Join(root, "product")
	write(t, filepath.Join(model, "requirements", "core", "001-foo.md"), "# Foo Requirement\n")
	write(t, filepath.Join(model, "use-cases", "core", "foo.md"), "# Foo Feature\n")
	write(t, filepath.Join(model, "components", "core", "comp.md"), "# Comp\n")
	write(t, filepath.Join(product, "core", "001-login.md"), "# Login PR\n")
	return model, product, &tracegraph.Graph{
		DeriveReqt: []tracegraph.RelationEntry{{
			Source:  "product/core/001-login",
			Targets: []string{"model/requirements/core/001-foo"},
		}},
		Refinements: []tracegraph.RelationEntry{{
			Source:  "model/requirements/core/001-foo",
			Targets: []string{"model/use-cases/core/foo"},
		}},
		Satisfactions: []tracegraph.RelationEntry{{
			Source:  "model/requirements/core/001-foo",
			Targets: []string{"model/components/core/comp"},
		}},
	}
}

func sampleModelGraph(t *testing.T) (string, *tracegraph.Graph) {
	model, _, traceability := sampleGraph(t)
	return model, &tracegraph.Graph{
		Refinements:   traceability.Refinements,
		Satisfactions: traceability.Satisfactions,
		Layout: []tracegraph.LayoutEntry{{
			ID:     "model/use-cases/core/foo",
			X:      12.5,
			Y:      8.75,
			Locked: true,
		}},
	}
}

func TestBuild(t *testing.T) {
	model, traceability := sampleModelGraph(t)
	g, err := Build(model, "", traceability)
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Nodes) != 3 {
		t.Errorf("want 3 nodes, got %d", len(g.Nodes))
	}
	if len(g.Edges) != 2 {
		t.Errorf("want 2 edges, got %d: %+v", len(g.Edges), g.Edges)
	}

	kinds := map[string]int{}
	for _, n := range g.Nodes {
		kinds[n.Kind]++
	}
	for _, want := range []string{"requirement", "use-case", "component"} {
		if kinds[want] != 1 {
			t.Errorf("expected one %s node, got %d", want, kinds[want])
		}
	}

	// Labels use the H1 text.
	for _, n := range g.Nodes {
		if n.Label == "" || n.Label == filepath.Base(n.Path) {
			t.Errorf("node %s has empty/basename label %q", n.Path, n.Label)
		}
	}
}

func TestBuild_NilGraph(t *testing.T) {
	if _, err := Build(t.TempDir(), "", nil); err == nil {
		t.Fatal("expected error for nil graph")
	}
}

func TestWriteMermaid(t *testing.T) {
	model, traceability := sampleModelGraph(t)
	g, err := Build(model, "", traceability)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := WriteMermaid(&buf, g); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"flowchart", "-->"} {
		if !strings.Contains(out, want) {
			t.Errorf("Mermaid output missing %q\n---\n%s", want, out)
		}
	}
	// Requirement uses [[..]] subroutine shape; component uses [(..)] cylinder.
	if !strings.Contains(out, "[[") || !strings.Contains(out, "[(") {
		t.Errorf("Mermaid output missing per-kind shapes\n---\n%s", out)
	}
}

func TestWriteJSON(t *testing.T) {
	model, traceability := sampleModelGraph(t)
	g, err := Build(model, "", traceability)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := WriteJSON(&out, g); err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Nodes []struct {
			ID     string `json:"id"`
			Path   string `json:"path"`
			Label  string `json:"label"`
			Kind   string `json:"kind"`
			Layout *struct {
				X      float64 `json:"x"`
				Y      float64 `json:"y"`
				Locked bool    `json:"locked"`
			} `json:"layout"`
		} `json:"nodes"`
		Edges []struct {
			Source string `json:"source"`
			Target string `json:"target"`
			Kind   string `json:"kind"`
		} `json:"edges"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Nodes) != 3 || len(payload.Edges) != 2 {
		t.Fatalf("unexpected payload sizes: %+v", payload)
	}
	var sawFeatureLayout bool
	for _, node := range payload.Nodes {
		if node.ID != "model/use-cases/core/foo" {
			continue
		}
		if node.Layout == nil || node.Layout.X != 12.5 || node.Layout.Y != 8.75 || !node.Layout.Locked {
			t.Fatalf("use-case layout missing from JSON node: %+v", node)
		}
		sawFeatureLayout = true
	}
	if !sawFeatureLayout {
		t.Fatalf("use-case node missing from JSON payload: %+v", payload.Nodes)
	}
}

func TestBuild_WithProductTree(t *testing.T) {
	model, product, traceability := sampleGraph(t)
	g, err := Build(model, product, traceability)
	if err != nil {
		t.Fatal(err)
	}

	kinds := map[string]int{}
	for _, n := range g.Nodes {
		kinds[n.Kind]++
	}
	if kinds["product-requirement"] != 1 {
		t.Errorf("want 1 product-requirement node, got %d", kinds["product-requirement"])
	}
	if kinds["requirement"] != 1 || kinds["use-case"] != 1 || kinds["component"] != 1 {
		t.Errorf("want 1 req+1 use-case+1 component, got kinds=%v", kinds)
	}

	// PR -> req edge present; req -> use-case edge present.
	var sawPRtoReq, sawReqToFeat bool
	for _, e := range g.Edges {
		if strings.Contains(e.Source, "requirements") && strings.Contains(e.Target, "product") {
			sawPRtoReq = true
		}
		if strings.Contains(e.Source, "use-cases") && strings.Contains(e.Target, "requirements") {
			sawReqToFeat = true
		}
	}
	if !sawPRtoReq {
		t.Errorf("expected a req->PR edge, got %+v", g.Edges)
	}
	if !sawReqToFeat {
		t.Errorf("expected a use-case->req edge, got %+v", g.Edges)
	}
}
