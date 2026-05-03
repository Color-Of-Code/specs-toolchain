// Package visualize builds rendered traceability graphs from the canonical
// traceability graph and the current model and product trees.
//
// The graph captures the traceability edge families used throughout the
// toolchain:
//
//	product-requirement -> requirement
//	requirement -> feature/component/api/service
//
// Output formats: DOT (graphviz), Mermaid (`flowchart`), and JSON.
package visualize

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	tracegraph "github.com/Color-Of-Code/specs-toolchain/engine/internal/graph"
)

// Edge is a directed link between two model files (workspace-relative).
type Edge struct {
	From   string
	To     string
	Source string
	Target string
	Kind   string
}

// Node carries a stable id derived from its path plus a display label
// (the file's H1, or the basename when no H1 is present).
type Node struct {
	ID        string // sanitised, unique per file
	NodeID    string
	Path      string // workspace-relative markdown file path
	Label     string
	Summary   string // body text of the first ## section after the title
	Kind      string // product-requirement | requirement | feature | component | api | service
	HasLayout bool
	X         float64
	Y         float64
	Locked    bool
}

// Graph is the resolved traceability graph for a model tree.
type Graph struct {
	Nodes []*Node
	Edges []Edge
}

var (
	h1Re      = regexp.MustCompile(`(?m)^#\s+(.+?)\s*$`)
	sectionRe = regexp.MustCompile(`(?ms)^##\s+[^\n]+\n(.*?)(?:\n##\s|\z)`)
)

// Build projects the canonical traceability graph into the visualization DTO.
// It includes ALL artifact files found under modelDir and productDir, not just
// those already referenced by a graph relation, so unconnected artifacts are
// visible and can be wired up interactively.
func Build(modelDir, productDir string, g *tracegraph.Graph) (*Graph, error) {
	if g == nil {
		return nil, fmt.Errorf("graph is nil")
	}

	result := &Graph{}
	nodesByNodeID := make(map[string]*Node, len(g.NodeIDs()))
	layoutByID := make(map[string]tracegraph.LayoutEntry, len(g.Layout))
	for _, entry := range g.Layout {
		layoutByID[entry.ID] = entry
	}
	for _, nodeID := range g.NodeIDs() {
		path := tracegraph.MarkdownPath(nodeID)
		absPath, err := markdownPathForNodeID(nodeID, modelDir, productDir)
		if err != nil {
			return nil, err
		}
		node := &Node{
			ID:      sanitiseID(nodeID),
			NodeID:  nodeID,
			Path:    path,
			Label:   readMarkdownTitle(absPath, path),
			Summary: readMarkdownSummary(absPath),
			Kind:    tracegraph.KindForNodeID(nodeID),
		}
		if layout, ok := layoutByID[nodeID]; ok {
			node.HasLayout = true
			node.X = layout.X
			node.Y = layout.Y
			node.Locked = layout.Locked
		}
		result.Nodes = append(result.Nodes, node)
		nodesByNodeID[nodeID] = node
	}

	// Walk model and product trees to include artifacts not yet in any relation.
	for _, root := range []struct {
		dir    string
		prefix string
	}{
		{modelDir, "model/"},
		{productDir, "product/"},
	} {
		if root.dir == "" {
			continue
		}
		_ = filepath.Walk(root.dir, func(p string, info os.FileInfo, err error) error {
			if err != nil || info == nil || info.IsDir() {
				return err
			}
			if !strings.HasSuffix(p, ".md") || filepath.Base(p) == "_index.md" {
				return nil
			}
			rel, err := filepath.Rel(root.dir, p)
			if err != nil {
				return nil
			}
			nodeID := root.prefix + filepath.ToSlash(strings.TrimSuffix(rel, ".md"))
			if _, exists := nodesByNodeID[nodeID]; exists {
				return nil
			}
			node := &Node{
				ID:      sanitiseID(nodeID),
				NodeID:  nodeID,
				Path:    nodeID + ".md",
				Label:   readMarkdownTitle(p, nodeID+".md"),
				Summary: readMarkdownSummary(p),
				Kind:    tracegraph.KindForNodeID(nodeID),
			}
			result.Nodes = append(result.Nodes, node)
			nodesByNodeID[nodeID] = node
			return nil
		})
	}

	sort.Slice(result.Nodes, func(i, j int) bool { return result.Nodes[i].Path < result.Nodes[j].Path })

	seen := map[string]bool{}
	appendEdges := func(kind string, entries []tracegraph.RelationEntry) {
		for _, entry := range entries {
			from, ok := nodesByNodeID[entry.Source]
			if !ok {
				continue
			}
			for _, target := range entry.Targets {
				to, ok := nodesByNodeID[target]
				if !ok {
					continue
				}
				key := from.ID + "->" + to.ID
				if seen[key] {
					continue
				}
				seen[key] = true
				result.Edges = append(result.Edges, Edge{From: from.ID, To: to.ID, Source: from.NodeID, Target: to.NodeID, Kind: kind})
			}
		}
	}
	appendEdges(string(tracegraph.PartKindRealization), g.Realizations)
	appendEdges(string(tracegraph.PartKindFeatureImplementation), g.FeatureImplementations)
	appendEdges(string(tracegraph.PartKindComponentImplementation), g.ComponentImplementations)
	appendEdges(string(tracegraph.PartKindServiceImplementation), g.ServiceImplementations)
	appendEdges(string(tracegraph.PartKindAPIImplementation), g.APIImplementations)
	sort.Slice(result.Edges, func(i, j int) bool {
		if result.Edges[i].From != result.Edges[j].From {
			return result.Edges[i].From < result.Edges[j].From
		}
		return result.Edges[i].To < result.Edges[j].To
	})
	return result, nil
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

func WriteJSON(out io.Writer, g *Graph) error {
	type layout struct {
		X      float64 `json:"x"`
		Y      float64 `json:"y"`
		Locked bool    `json:"locked,omitempty"`
	}
	type node struct {
		ID      string  `json:"id"`
		Path    string  `json:"path"`
		Label   string  `json:"label"`
		Kind    string  `json:"kind"`
		Summary string  `json:"summary,omitempty"`
		Layout  *layout `json:"layout,omitempty"`
	}
	type edge struct {
		Source string `json:"source"`
		Target string `json:"target"`
		Kind   string `json:"kind"`
	}
	payload := struct {
		Nodes []node `json:"nodes"`
		Edges []edge `json:"edges"`
	}{
		Nodes: make([]node, 0, len(g.Nodes)),
		Edges: make([]edge, 0, len(g.Edges)),
	}
	for _, current := range g.Nodes {
		jsonNode := node{
			ID:      current.NodeID,
			Path:    current.Path,
			Label:   current.Label,
			Kind:    current.Kind,
			Summary: current.Summary,
		}
		if current.HasLayout {
			jsonNode.Layout = &layout{X: current.X, Y: current.Y, Locked: current.Locked}
		}
		payload.Nodes = append(payload.Nodes, jsonNode)
	}
	for _, current := range g.Edges {
		payload.Edges = append(payload.Edges, edge{Source: current.Source, Target: current.Target, Kind: current.Kind})
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(payload)
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

func readMarkdownTitle(path, fallback string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return strings.TrimSuffix(filepath.Base(fallback), ".md")
	}
	if m := h1Re.FindStringSubmatch(string(data)); len(m) == 2 {
		return strings.TrimSpace(m[1])
	}
	return strings.TrimSuffix(filepath.Base(fallback), ".md")
}

// readMarkdownSummary returns the trimmed body of the first ## section.
func readMarkdownSummary(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	if m := sectionRe.FindStringSubmatch(string(data)); len(m) == 2 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func markdownPathForNodeID(nodeID, modelDir, productDir string) (string, error) {
	switch {
	case strings.HasPrefix(nodeID, "product/"):
		if productDir == "" {
			return "", nil
		}
		return filepath.Join(productDir, filepath.FromSlash(strings.TrimPrefix(tracegraph.MarkdownPath(nodeID), "product/"))), nil
	case strings.HasPrefix(nodeID, "model/"):
		if modelDir == "" {
			return "", nil
		}
		return filepath.Join(modelDir, filepath.FromSlash(strings.TrimPrefix(tracegraph.MarkdownPath(nodeID), "model/"))), nil
	default:
		return "", fmt.Errorf("unsupported node id %q", nodeID)
	}
}
