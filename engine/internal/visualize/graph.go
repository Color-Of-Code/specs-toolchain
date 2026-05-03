// Package visualize builds traceability graphs from the model and product
// trees.
//
// The graph captures the traceability edge families used throughout the
// toolchain:
//
//	product-requirement -> requirement                  (## Realised By)
//	requirement -> feature/component/api/service        (## Implemented By)
//	feature/component/... -> requirement                (## Requirements)
//
// Output formats: DOT (graphviz) and Mermaid (`flowchart`).
package visualize

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Edge is a directed link between two model files (workspace-relative).
type Edge struct {
	From string
	To   string
}

// Node carries a stable id derived from its path plus a display label
// (the file's H1, or the basename when no H1 is present).
type Node struct {
	ID    string // sanitised, unique per file
	Path  string // workspace-relative; product-requirement nodes are prefixed with "product/"
	Label string
	Kind  string // product-requirement | requirement | feature | component | api | service
}

// Graph is the resolved traceability graph for a model tree.
type Graph struct {
	Nodes []*Node
	Edges []Edge
}

var (
	linkRe          = regexp.MustCompile(`\]\(([^)]+)\)`)
	sectionHeaderRe = regexp.MustCompile(`(?m)^##\s+(.+?)\s*$`)
	h1Re            = regexp.MustCompile(`(?m)^#\s+(.+?)\s*$`)
)

// Build walks modelDir and productDir and returns the graph. productDir may
// be empty or non-existent; in that case product-requirement nodes are
// simply absent from the result.
func Build(modelDir, productDir string) (*Graph, error) {
	if _, err := os.Stat(modelDir); err != nil {
		return nil, fmt.Errorf("model dir %s: %w", modelDir, err)
	}

	nodesByPath := map[string]*Node{}
	addModelNode := func(absPath string) {
		if _, ok := nodesByPath[absPath]; ok {
			return
		}
		rel, err := filepath.Rel(modelDir, absPath)
		if err != nil {
			return
		}
		rel = filepath.ToSlash(rel)
		kind := modelKindFor(rel)
		if kind == "" {
			return
		}
		nodesByPath[absPath] = &Node{
			ID:    sanitiseID("model/" + rel),
			Path:  rel,
			Label: readH1(absPath, rel),
			Kind:  kind,
		}
	}
	addProductNode := func(absPath string) {
		if productDir == "" {
			return
		}
		if _, ok := nodesByPath[absPath]; ok {
			return
		}
		rel, err := filepath.Rel(productDir, absPath)
		if err != nil {
			return
		}
		rel = filepath.ToSlash(rel)
		display := "product/" + rel
		nodesByPath[absPath] = &Node{
			ID:    sanitiseID(display),
			Path:  display,
			Label: readH1(absPath, rel),
			Kind:  "product-requirement",
		}
	}

	var edges []Edge
	// walkModel walks an area within modelDir for a given section header,
	// ensuring nodes on both sides of each edge.
	walkModel := func(area, section string, addTarget func(string)) {
		root := filepath.Join(modelDir, area)
		if _, err := os.Stat(root); err != nil {
			return
		}
		_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
			if err != nil || info == nil || info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(p, ".md") || filepath.Base(p) == "_index.md" {
				return nil
			}
			data, err := os.ReadFile(p)
			if err != nil {
				return nil
			}
			body := extractSection(string(data), section)
			if body == "" {
				return nil
			}
			addModelNode(p)
			for _, t := range linksIn(body) {
				abs := resolveTarget(p, t)
				if abs == "" {
					continue
				}
				addTarget(abs)
				if _, ok := nodesByPath[abs]; !ok {
					continue
				}
				edges = append(edges, Edge{From: p, To: abs})
			}
			return nil
		})
	}
	walkProduct := func(section string) {
		if productDir == "" {
			return
		}
		if _, err := os.Stat(productDir); err != nil {
			return
		}
		_ = filepath.Walk(productDir, func(p string, info os.FileInfo, err error) error {
			if err != nil || info == nil || info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(p, ".md") || filepath.Base(p) == "_index.md" {
				return nil
			}
			data, err := os.ReadFile(p)
			if err != nil {
				return nil
			}
			body := extractSection(string(data), section)
			if body == "" {
				return nil
			}
			addProductNode(p)
			for _, t := range linksIn(body) {
				abs := resolveTarget(p, t)
				if abs == "" {
					continue
				}
				addModelNode(abs)
				if _, ok := nodesByPath[abs]; !ok {
					continue
				}
				edges = append(edges, Edge{From: p, To: abs})
			}
			return nil
		})
	}

	walkProduct("Realised By")
	walkModel("requirements", "Implemented By", addModelNode)
	for _, area := range []string{"features", "components", "apis", "services"} {
		walkModel(area, "Requirements", addModelNode)
	}

	g := &Graph{}
	for _, n := range nodesByPath {
		g.Nodes = append(g.Nodes, n)
	}
	sort.Slice(g.Nodes, func(i, j int) bool { return g.Nodes[i].Path < g.Nodes[j].Path })

	// Re-key edges from absolute paths to node IDs (and dedupe).
	seen := map[string]bool{}
	for _, e := range edges {
		from, fok := nodesByPath[e.From]
		to, tok := nodesByPath[e.To]
		if !fok || !tok {
			continue
		}
		key := from.ID + "->" + to.ID
		if seen[key] {
			continue
		}
		seen[key] = true
		g.Edges = append(g.Edges, Edge{From: from.ID, To: to.ID})
	}
	sort.Slice(g.Edges, func(i, j int) bool {
		if g.Edges[i].From != g.Edges[j].From {
			return g.Edges[i].From < g.Edges[j].From
		}
		return g.Edges[i].To < g.Edges[j].To
	})
	return g, nil
}

// WriteDOT renders g in graphviz DOT format.
func WriteDOT(out io.Writer, g *Graph) error {
	fmt.Fprintln(out, "digraph traceability {")
	fmt.Fprintln(out, "  rankdir=LR;")
	fmt.Fprintln(out, "  node [shape=box, style=\"rounded,filled\", fontname=\"Helvetica\"];")
	for _, kind := range []string{"product-requirement", "requirement", "feature", "component", "api", "service"} {
		fmt.Fprintf(out, "  // %s nodes\n", kind)
		fmt.Fprintf(out, "  subgraph cluster_%s {\n", strings.ReplaceAll(kind, "-", "_"))
		fmt.Fprintf(out, "    label=%q; style=dashed;\n", clusterLabel(kind))
		fmt.Fprintf(out, "    node [fillcolor=%q];\n", colorFor(kind))
		for _, n := range g.Nodes {
			if n.Kind != kind {
				continue
			}
			fmt.Fprintf(out, "    %s [label=%q, tooltip=%q];\n", n.ID, n.Label, n.Path)
		}
		fmt.Fprintln(out, "  }")
	}
	for _, e := range g.Edges {
		fmt.Fprintf(out, "  %s -> %s;\n", e.From, e.To)
	}
	fmt.Fprintln(out, "}")
	return nil
}

// WriteMermaid renders g as a Mermaid flowchart.
func WriteMermaid(out io.Writer, g *Graph) error {
	fmt.Fprintln(out, "flowchart LR")
	for _, n := range g.Nodes {
		shapeOpen, shapeClose := "[", "]"
		switch n.Kind {
		case "product-requirement":
			shapeOpen, shapeClose = "([", "])"
		case "requirement":
			shapeOpen, shapeClose = "[[", "]]"
		case "component":
			shapeOpen, shapeClose = "[(", ")]"
		}
		fmt.Fprintf(out, "  %s%s%q%s\n", n.ID, shapeOpen, n.Label, shapeClose)
	}
	for _, e := range g.Edges {
		fmt.Fprintf(out, "  %s --> %s\n", e.From, e.To)
	}
	return nil
}

func modelKindFor(rel string) string {
	switch {
	case strings.HasPrefix(rel, "requirements/"):
		return "requirement"
	case strings.HasPrefix(rel, "features/"):
		return "feature"
	case strings.HasPrefix(rel, "components/"):
		return "component"
	case strings.HasPrefix(rel, "apis/"):
		return "api"
	case strings.HasPrefix(rel, "services/"):
		return "service"
	}
	return ""
}

func clusterLabel(kind string) string {
	switch kind {
	case "product-requirement":
		return "Product requirements"
	case "requirement":
		return "Requirements"
	case "feature":
		return "Features"
	case "component":
		return "Components"
	case "api":
		return "APIs"
	case "service":
		return "Services"
	}
	return kind
}

func colorFor(kind string) string {
	switch kind {
	case "product-requirement":
		return "#fce4ec"
	case "requirement":
		return "#e3f2fd"
	case "feature":
		return "#fff3e0"
	case "component":
		return "#e8f5e9"
	case "api":
		return "#f3e5f5"
	case "service":
		return "#fff8e1"
	}
	return "#ffffff"
}

func sanitiseID(rel string) string {
	rel = strings.TrimSuffix(rel, ".md")
	var b strings.Builder
	b.WriteByte('n')
	for _, r := range rel {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}

// readH1 returns the first H1 from path, or a slug-ish fallback.
func readH1(path, rel string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return rel
	}
	if m := h1Re.FindStringSubmatch(string(data)); len(m) == 2 {
		return strings.TrimSpace(m[1])
	}
	return strings.TrimSuffix(filepath.Base(rel), ".md")
}

func extractSection(body, name string) string {
	matches := sectionHeaderRe.FindAllStringSubmatchIndex(body, -1)
	for i, m := range matches {
		title := strings.TrimSpace(body[m[2]:m[3]])
		if !strings.EqualFold(title, name) {
			continue
		}
		end := len(body)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		return body[m[1]:end]
	}
	return ""
}

func linksIn(body string) []string {
	var out []string
	for _, m := range linkRe.FindAllStringSubmatch(body, -1) {
		t := strings.TrimSpace(m[1])
		if t == "" {
			continue
		}
		switch {
		case strings.HasPrefix(t, "http://"),
			strings.HasPrefix(t, "https://"),
			strings.HasPrefix(t, "mailto:"),
			strings.HasPrefix(t, "#"):
			continue
		}
		if i := strings.IndexAny(t, "#?"); i >= 0 {
			t = t[:i]
		}
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func resolveTarget(fromFile, target string) string {
	if filepath.IsAbs(target) {
		return ""
	}
	abs := filepath.Clean(filepath.Join(filepath.Dir(fromFile), target))
	st, err := os.Stat(abs)
	if err != nil || st.IsDir() {
		return ""
	}
	if !strings.HasSuffix(abs, ".md") || filepath.Base(abs) == "_index.md" {
		return ""
	}
	return abs
}
