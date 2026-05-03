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

	tracegraph "github.com/Color-Of-Code/specs-toolchain/engine/internal/graph"
	"github.com/Color-Of-Code/specs-toolchain/engine/internal/visualize"
)

type layoutSaveRequest struct {
	Nodes []layoutSaveNode `json:"nodes"`
}

type layoutSaveNode struct {
	ID     string  `json:"id"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Locked bool    `json:"locked,omitempty"`
}

func newTraceabilityUIHandler(start string) (http.Handler, error) {
	assets, err := visualize.WebAssetFS()
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))
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
			Title:         "Specs: Traceability",
			Hint:          "Click a node to inspect the underlying markdown artifact. Drag nodes, then save layout back into the canonical YAML.",
			GraphURL:      "/graph.json",
			SaveLayoutURL: "/layout",
			DotURL:        "/graph.dot",
			JSONURL:       "/graph.json",
			ArtifactURL:   "/artifact",
			Stylesheet:    "/assets/traceability-view.css",
			CytoscapeJS:   "/assets/cytoscape.min.js",
			AppJS:         "/assets/traceability-view.js",
			EmptyMessage:  "No traceability data found.",
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
	cfg, graphView, err := loadTraceabilityVisualization(start)
	if err != nil {
		return err
	}
	traceability, err := tracegraph.Load(cfg.GraphManifest)
	if err != nil {
		return exitWith(1, "load graph %s: %v", cfg.GraphManifest, err)
	}
	allowed := make(map[string]struct{}, len(graphView.Nodes))
	for _, node := range graphView.Nodes {
		allowed[node.NodeID] = struct{}{}
	}
	layouts := make([]tracegraph.LayoutEntry, 0, len(nodes))
	seen := make(map[string]struct{}, len(nodes))
	for index, current := range nodes {
		normalizedID, err := tracegraph.NormalizeNodeID(current.ID)
		if err != nil {
			return fmt.Errorf("layout node %d id: %w", index, err)
		}
		if _, ok := allowed[normalizedID]; !ok {
			return fmt.Errorf("layout node %d id %q is not in the traceability graph", index, normalizedID)
		}
		if _, exists := seen[normalizedID]; exists {
			return fmt.Errorf("layout node %d duplicates id %q", index, normalizedID)
		}
		seen[normalizedID] = struct{}{}
		layouts = append(layouts, tracegraph.LayoutEntry{ID: normalizedID, X: current.X, Y: current.Y, Locked: current.Locked})
	}
	sort.Slice(layouts, func(i, j int) bool {
		return layouts[i].ID < layouts[j].ID
	})
	traceability.Layout = layouts
	if err := tracegraph.Write(cfg.GraphManifest, traceability); err != nil {
		return fmt.Errorf("write graph %s: %w", cfg.GraphManifest, err)
	}
	return nil
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
