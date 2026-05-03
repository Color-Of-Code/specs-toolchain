package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Color-Of-Code/specs-toolchain/engine/internal/visualize"
)

func newTraceabilityUIHandler(start string) (http.Handler, error) {
	assets, err := visualize.WebAssetFS()
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))
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
			Title:        "Specs: Traceability",
			Hint:         "Click a node to inspect the underlying markdown artifact. The graph reloads from the canonical YAML on every request.",
			GraphURL:     "/graph.json",
			DotURL:       "/graph.dot",
			JSONURL:      "/graph.json",
			ArtifactURL:  "/artifact",
			Stylesheet:   "/assets/traceability-view.css",
			CytoscapeJS:  "/assets/cytoscape.min.js",
			AppJS:        "/assets/traceability-view.js",
			EmptyMessage: "No traceability data found.",
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(page.Bytes())
	})
	return mux, nil
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
