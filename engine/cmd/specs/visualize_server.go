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

type relationSaveRequest struct {
	Edges []relationSaveEdge `json:"edges"`
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
			Hint:             "Click a node to inspect the underlying markdown artifact. Use the layout controls to switch between layered, organic, and grid arrangements.",
			GraphURL:         "/graph.json",
			SaveRelationsURL: "/relations",
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

func saveTraceabilityRelations(start string, edges []relationSaveEdge) error {
	cfg, err := config.Load(start)
	if err != nil {
		return err
	}
	traceability, err := tracegraph.Load(cfg.GraphManifest)
	if err != nil {
		return exitWith(1, "load graph %s: %v", cfg.GraphManifest, err)
	}
	allowed, err := traceabilityAllowedNodeIDs(cfg.ModelDir, cfg.ProductDir, traceability)
	if err != nil {
		return err
	}
	relations, err := relationEntriesFromSaveEdges(edges, allowed)
	if err != nil {
		return err
	}
	// Replace required relation kinds unconditionally; update optional kinds only
	// when they appeared in the payload (preserves e.g. trace relations the UI
	// does not yet know how to edit).
	for _, spec := range tracegraph.AllRelationSpecs() {
		entries, inPayload := relations[spec.Kind]
		if spec.Required {
			if !inPayload {
				entries = []tracegraph.RelationEntry{}
			}
			traceability.Relations[spec.Kind] = entries
		} else if inPayload {
			traceability.Relations[spec.Kind] = entries
		}
	}
	if err := tracegraph.Write(cfg.GraphManifest, traceability); err != nil {
		return fmt.Errorf("write graph %s: %w", cfg.GraphManifest, err)
	}
	return nil
}

func traceabilityAllowedNodeIDs(modelDir, productDir string, g *tracegraph.Graph) (map[string]struct{}, error) {
	allowed := map[string]struct{}{}
	for _, entries := range g.Relations {
		for _, entry := range entries {
			allowed[entry.Source] = struct{}{}
			for _, target := range entry.Targets {
				allowed[target] = struct{}{}
			}
		}
	}
	for _, root := range []struct {
		dir    string
		prefix string
	}{
		{productDir, "product"},
		{filepath.Join(modelDir, "requirements"), "model/requirements"},
		{filepath.Join(modelDir, "use-cases"), "model/use-cases"},
		{filepath.Join(modelDir, "components"), "model/components"},
	} {
		if err := collectArtifactNodeIDs(allowed, root.dir, root.prefix); err != nil {
			return nil, err
		}
	}
	return allowed, nil
}

func collectArtifactNodeIDs(allowed map[string]struct{}, dir, prefix string) error {
	if dir == "" {
		return nil
	}
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat %s: %w", dir, err)
	}
	if !info.IsDir() {
		return nil
	}
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info == nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") || filepath.Base(path) == "_index.md" {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		nodeID, err := tracegraph.NormalizeNodeID(filepath.ToSlash(filepath.Join(prefix, strings.TrimSuffix(rel, filepath.Ext(rel)))))
		if err != nil {
			return nil
		}
		allowed[nodeID] = struct{}{}
		return nil
	})
}

func relationEntriesFromSaveEdges(edges []relationSaveEdge, allowed map[string]struct{}) (map[tracegraph.PartKind][]tracegraph.RelationEntry, error) {
	byKind := map[tracegraph.PartKind]map[string]map[string]struct{}{}
	seen := map[string]struct{}{}
	specs := tracegraph.AllRelationSpecs()
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
		kind := tracegraph.PartKind(current.Kind)
		var spec *tracegraph.RelationPartSpec
		for i := range specs {
			if specs[i].Kind == kind {
				spec = &specs[i]
				break
			}
		}
		if spec == nil {
			return nil, fmt.Errorf("relation edge %d kind %q is unsupported", index, current.Kind)
		}
		// Derive the expected source/target node kinds as the frontend sends them.
		frontendSrc, frontendTgt := spec.SourceKind, spec.TargetKind
		if spec.InvertOnSave {
			frontendSrc, frontendTgt = spec.TargetKind, spec.SourceKind
		}
		if frontendSrc != "" && tracegraph.KindForNodeID(normalizedSource) != frontendSrc {
			return nil, fmt.Errorf("relation edge %d source %q must be a %s", index, normalizedSource, frontendSrc)
		}
		if frontendTgt != "" && tracegraph.KindForNodeID(normalizedTarget) != frontendTgt {
			return nil, fmt.Errorf("relation edge %d target %q must be a %s", index, normalizedTarget, frontendTgt)
		}
		key := current.Kind + "\x00" + normalizedSource + "\x00" + normalizedTarget
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("relation edge %d duplicates %s -> %s (%s)", index, normalizedSource, normalizedTarget, current.Kind)
		}
		seen[key] = struct{}{}
		if byKind[kind] == nil {
			byKind[kind] = map[string]map[string]struct{}{}
		}
		if spec.InvertOnSave {
			addRelationTarget(byKind[kind], normalizedTarget, normalizedSource)
		} else {
			addRelationTarget(byKind[kind], normalizedSource, normalizedTarget)
		}
	}
	result := map[tracegraph.PartKind][]tracegraph.RelationEntry{}
	for kind, targets := range byKind {
		result[kind] = relationEntriesFromTargetMap(targets)
	}
	return result, nil
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
