package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Color-Of-Code/specs-toolchain/engine/internal/config"
	tracegraph "github.com/Color-Of-Code/specs-toolchain/engine/internal/graph"
	"github.com/Color-Of-Code/specs-toolchain/engine/internal/visualize"
)

type layoutSaveRequest struct {
	Nodes []layoutSaveNode `json:"nodes"`
}

type relationSaveRequest struct {
	Edges []relationSaveEdge `json:"edges"`
}

type layoutSaveNode struct {
	ID     string  `json:"id"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Locked bool    `json:"locked,omitempty"`
}

type relationSaveEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Kind   string `json:"kind"`
}

func newTraceabilityUIHandler(start string) (http.Handler, error) {
	assets, err := visualize.WebAssetFS()
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))
	mux.HandleFunc("/relations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		defer r.Body.Close()
		var req relationSaveRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("decode relation request: %v", err), http.StatusBadRequest)
			return
		}
		if err := saveTraceabilityRelations(start, req.Edges); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/layout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		defer r.Body.Close()
		var req layoutSaveRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("decode layout request: %v", err), http.StatusBadRequest)
			return
		}
		if err := saveTraceabilityLayout(start, req.Nodes); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/graph.json", func(w http.ResponseWriter, _ *http.Request) {
		_, graphView, err := loadTraceabilityVisualization(start)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := visualize.WriteJSON(w, graphView); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/graph.dot", func(w http.ResponseWriter, _ *http.Request) {
		_, graphView, err := loadTraceabilityVisualization(start)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/vnd.graphviz; charset=utf-8")
		if err := visualize.WriteDOT(w, graphView); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/artifact", func(w http.ResponseWriter, r *http.Request) {
		cfg, _, err := loadTraceabilityVisualization(start)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		relPath, absPath, err := resolveArtifactPath(cfg.SpecsRoot, r.URL.Query().Get("path"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		body, err := os.ReadFile(absPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("read %s: %v", relPath, err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := visualize.WriteArtifactPage(w, visualize.ArtifactPageData{
			Title: filepath.Base(relPath),
			Path:  filepath.ToSlash(relPath),
			Body:  string(body),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		var page bytes.Buffer
		if err := visualize.WriteTraceabilityPage(&page, visualize.TraceabilityPageData{
			Title:            "Specs: Traceability",
			Hint:             "Click a node to inspect the underlying markdown artifact. Drag nodes, then save layout back into the canonical YAML.",
			GraphURL:         "/graph.json",
			SaveRelationsURL: "/relations",
			SaveLayoutURL:    "/layout",
			DotURL:           "/graph.dot",
			JSONURL:          "/graph.json",
			ArtifactURL:      "/artifact",
			Stylesheet:       "/assets/traceability-view.css",
			CytoscapeJS:      "/assets/cytoscape.min.js",
			AppJS:            "/assets/traceability-view.js",
			EmptyMessage:     "No traceability data found.",
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(page.Bytes())
	})
	return mux, nil
}

func saveTraceabilityLayout(start string, nodes []layoutSaveNode) error {
	cfg, err := config.Load(start)
	if err != nil {
		return err
	}
	traceability, err := tracegraph.Load(cfg.GraphManifest)
	if err != nil {
		return exitWith(1, "load graph %s: %v", cfg.GraphManifest, err)
	}
	layouts, err := layoutEntriesFromSaveNodes(nodes, traceabilityAllowedNodeIDs(traceability))
	if err != nil {
		return err
	}
	traceability.Layout = layouts
	if err := tracegraph.Write(cfg.GraphManifest, traceability); err != nil {
		return fmt.Errorf("write graph %s: %w", cfg.GraphManifest, err)
	}
	return nil
}

func saveTraceabilityRelations(start string, edges []relationSaveEdge) error {
	cfg, err := config.Load(start)
	if err != nil {
		return err
	}
	traceability, err := tracegraph.Load(cfg.GraphManifest)
	if err != nil {
		return exitWith(1, "load graph %s: %v", cfg.GraphManifest, err)
	}
	relations, err := relationEntriesFromSaveEdges(edges, traceabilityAllowedNodeIDs(traceability))
	if err != nil {
		return err
	}
	traceability.Realizations = relations.realizations
	traceability.FeatureImplementations = relations.featureImplementations
	traceability.ComponentImplementations = relations.componentImplementations
	traceability.ServiceImplementations = relations.serviceImplementations
	traceability.APIImplementations = relations.apiImplementations
	if err := tracegraph.Write(cfg.GraphManifest, traceability); err != nil {
		return fmt.Errorf("write graph %s: %w", cfg.GraphManifest, err)
	}
	return nil
}

func traceabilityAllowedNodeIDs(g *tracegraph.Graph) map[string]struct{} {
	allowed := map[string]struct{}{}
	for _, entries := range [][]tracegraph.RelationEntry{
		g.Realizations,
		g.FeatureImplementations,
		g.ComponentImplementations,
		g.ServiceImplementations,
		g.APIImplementations,
	} {
		for _, entry := range entries {
			allowed[entry.Source] = struct{}{}
			for _, target := range entry.Targets {
				allowed[target] = struct{}{}
			}
		}
	}
	for _, baseline := range g.Baselines {
		allowed[baseline.Component] = struct{}{}
	}
	return allowed
}

func layoutEntriesFromSaveNodes(nodes []layoutSaveNode, allowed map[string]struct{}) ([]tracegraph.LayoutEntry, error) {
	layouts := make([]tracegraph.LayoutEntry, 0, len(nodes))
	seen := make(map[string]struct{}, len(nodes))
	for index, current := range nodes {
		normalizedID, err := tracegraph.NormalizeNodeID(current.ID)
		if err != nil {
			return nil, fmt.Errorf("layout node %d id: %w", index, err)
		}
		if _, ok := allowed[normalizedID]; !ok {
			return nil, fmt.Errorf("layout node %d id %q is not in the traceability graph", index, normalizedID)
		}
		if _, exists := seen[normalizedID]; exists {
			return nil, fmt.Errorf("layout node %d duplicates id %q", index, normalizedID)
		}
		seen[normalizedID] = struct{}{}
		layouts = append(layouts, tracegraph.LayoutEntry{ID: normalizedID, X: current.X, Y: current.Y, Locked: current.Locked})
	}
	sort.Slice(layouts, func(i, j int) bool {
		return layouts[i].ID < layouts[j].ID
	})
	return layouts, nil
}

type relationSaveFamilies struct {
	realizations             []tracegraph.RelationEntry
	featureImplementations   []tracegraph.RelationEntry
	componentImplementations []tracegraph.RelationEntry
	serviceImplementations   []tracegraph.RelationEntry
	apiImplementations       []tracegraph.RelationEntry
}

func relationEntriesFromSaveEdges(edges []relationSaveEdge, allowed map[string]struct{}) (*relationSaveFamilies, error) {
	realizations := map[string]map[string]struct{}{}
	featureImplementations := map[string]map[string]struct{}{}
	componentImplementations := map[string]map[string]struct{}{}
	serviceImplementations := map[string]map[string]struct{}{}
	apiImplementations := map[string]map[string]struct{}{}
	seen := map[string]struct{}{}
	for index, current := range edges {
		normalizedSource, err := tracegraph.NormalizeNodeID(current.Source)
		if err != nil {
			return nil, fmt.Errorf("relation edge %d source: %w", index, err)
		}
		normalizedTarget, err := tracegraph.NormalizeNodeID(current.Target)
		if err != nil {
			return nil, fmt.Errorf("relation edge %d target: %w", index, err)
		}
		if _, ok := allowed[normalizedSource]; !ok {
			return nil, fmt.Errorf("relation edge %d source %q is not in the traceability graph", index, normalizedSource)
		}
		if _, ok := allowed[normalizedTarget]; !ok {
			return nil, fmt.Errorf("relation edge %d target %q is not in the traceability graph", index, normalizedTarget)
		}
		sourceKind, targetKind, ok := relationKindsForSave(current.Kind)
		if !ok {
			return nil, fmt.Errorf("relation edge %d kind %q is unsupported", index, current.Kind)
		}
		if tracegraph.KindForNodeID(normalizedSource) != sourceKind {
			return nil, fmt.Errorf("relation edge %d source %q must be a %s", index, normalizedSource, sourceKind)
		}
		if tracegraph.KindForNodeID(normalizedTarget) != targetKind {
			return nil, fmt.Errorf("relation edge %d target %q must be a %s", index, normalizedTarget, targetKind)
		}
		key := current.Kind + "\x00" + normalizedSource + "\x00" + normalizedTarget
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("relation edge %d duplicates %s -> %s (%s)", index, normalizedSource, normalizedTarget, current.Kind)
		}
		seen[key] = struct{}{}
		switch current.Kind {
		case string(tracegraph.PartKindRealization):
			addRelationTarget(realizations, normalizedSource, normalizedTarget)
		case string(tracegraph.PartKindFeatureImplementation):
			addRelationTarget(featureImplementations, normalizedSource, normalizedTarget)
		case string(tracegraph.PartKindComponentImplementation):
			addRelationTarget(componentImplementations, normalizedSource, normalizedTarget)
		case string(tracegraph.PartKindServiceImplementation):
			addRelationTarget(serviceImplementations, normalizedSource, normalizedTarget)
		case string(tracegraph.PartKindAPIImplementation):
			addRelationTarget(apiImplementations, normalizedSource, normalizedTarget)
		}
	}
	return &relationSaveFamilies{
		realizations:             relationEntriesFromTargetMap(realizations),
		featureImplementations:   relationEntriesFromTargetMap(featureImplementations),
		componentImplementations: relationEntriesFromTargetMap(componentImplementations),
		serviceImplementations:   relationEntriesFromTargetMap(serviceImplementations),
		apiImplementations:       relationEntriesFromTargetMap(apiImplementations),
	}, nil
}

func relationKindsForSave(kind string) (string, string, bool) {
	switch kind {
	case string(tracegraph.PartKindRealization):
		return "product-requirement", "requirement", true
	case string(tracegraph.PartKindFeatureImplementation):
		return "requirement", "feature", true
	case string(tracegraph.PartKindComponentImplementation):
		return "requirement", "component", true
	case string(tracegraph.PartKindServiceImplementation):
		return "requirement", "service", true
	case string(tracegraph.PartKindAPIImplementation):
		return "requirement", "api", true
	default:
		return "", "", false
	}
}

func addRelationTarget(edges map[string]map[string]struct{}, source, target string) {
	targets, ok := edges[source]
	if !ok {
		targets = map[string]struct{}{}
		edges[source] = targets
	}
	targets[target] = struct{}{}
}

func relationEntriesFromTargetMap(edges map[string]map[string]struct{}) []tracegraph.RelationEntry {
	if len(edges) == 0 {
		return []tracegraph.RelationEntry{}
	}
	sources := make([]string, 0, len(edges))
	for source := range edges {
		sources = append(sources, source)
	}
	sort.Strings(sources)
	entries := make([]tracegraph.RelationEntry, 0, len(sources))
	for _, source := range sources {
		targets := make([]string, 0, len(edges[source]))
		for target := range edges[source] {
			targets = append(targets, target)
		}
		sort.Strings(targets)
		entries = append(entries, tracegraph.RelationEntry{Source: source, Targets: targets})
	}
	return entries
}

func resolveArtifactPath(specsRoot, requested string) (string, string, error) {
	relPath := filepath.ToSlash(strings.TrimSpace(requested))
	if relPath == "" {
		return "", "", fmt.Errorf("missing artifact path")
	}
	if strings.HasPrefix(relPath, "/") {
		return "", "", fmt.Errorf("artifact path %q must be repo-relative", requested)
	}
	absPath := filepath.Clean(filepath.Join(specsRoot, filepath.FromSlash(relPath)))
	relFromRoot, err := filepath.Rel(specsRoot, absPath)
	if err != nil {
		return "", "", err
	}
	if relFromRoot == ".." || strings.HasPrefix(relFromRoot, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("artifact path %q escapes the specs root", requested)
	}
	if filepath.Ext(absPath) != ".md" {
		return "", "", fmt.Errorf("artifact path %q must point to a markdown file", requested)
	}
	return filepath.ToSlash(relFromRoot), absPath, nil
}
