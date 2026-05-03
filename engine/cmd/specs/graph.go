package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Color-Of-Code/specs-toolchain/engine/internal/config"
	"github.com/Color-Of-Code/specs-toolchain/engine/internal/graph"
)

type graphValidateJSON struct {
	ManifestPath                 string `json:"manifest_path"`
	NodeCount                    int    `json:"node_count"`
	RealizationEdgeCount         int    `json:"realization_edge_count"`
	FeatureImplementationEdges   int    `json:"feature_implementation_edge_count"`
	ComponentImplementationEdges int    `json:"component_implementation_edge_count"`
	ServiceImplementationEdges   int    `json:"service_implementation_edge_count"`
	APIImplementationEdges       int    `json:"api_implementation_edge_count"`
	BaselineCount                int    `json:"baseline_count"`
	LayoutNodeCount              int    `json:"layout_node_count"`
	RepoCount                    int    `json:"repo_count"`
}

type graphImportJSON struct {
	ManifestPath                 string `json:"manifest_path"`
	RealizationEdgeCount         int    `json:"realization_edge_count"`
	FeatureImplementationEdges   int    `json:"feature_implementation_edge_count"`
	ComponentImplementationEdges int    `json:"component_implementation_edge_count"`
	ServiceImplementationEdges   int    `json:"service_implementation_edge_count"`
	APIImplementationEdges       int    `json:"api_implementation_edge_count"`
	BaselineCount                int    `json:"baseline_count"`
	DryRun                       bool   `json:"dry_run"`
}

func cmdGraph(args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: specs graph <validate|import-markdown>")
		return exitWith(2, "missing subcommand")
	}
	switch args[0] {
	case "validate":
		return cmdGraphValidate(args[1:])
	case "import-markdown":
		return cmdGraphImportMarkdown(args[1:])
	case "-h", "--help", "help":
		fmt.Fprintln(os.Stderr, "Usage: specs graph <validate|import-markdown> [flags]")
		return nil
	default:
		return exitWith(2, "unknown subcommand: specs graph %s", args[0])
	}
}

func cmdGraphImportMarkdown(args []string) error {
	fs := flag.NewFlagSet("graph import-markdown", flag.ContinueOnError)
	manifestPath := fs.String("manifest", "", "path to graph manifest to write (default: graph_manifest from .specs.yaml)")
	force := fs.Bool("force", false, "overwrite an existing graph manifest and part files")
	dryRun := fs.Bool("dry-run", false, "print what would be written without changing files")
	jsonOut := fs.Bool("json", false, "emit machine-readable import summary")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs graph import-markdown [--manifest <path>] [--force] [--dry-run] [--json]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return exitWith(2, "unexpected arguments: %v", fs.Args())
	}

	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	path, err := resolveGraphManifestPath(cfg, *manifestPath)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil && !*force {
		return exitWith(1, "graph manifest already exists at %s; rerun with --force to overwrite it", path)
	}

	g, err := graph.ImportMarkdown(cfg.ModelDir, cfg.ProductDir, cfg.BaselinesFile)
	if err != nil {
		return exitWith(1, "import markdown: %v", err)
	}
	summary := graphImportJSON{
		ManifestPath:                 path,
		RealizationEdgeCount:         relationEdgeCount(g.Realizations),
		FeatureImplementationEdges:   relationEdgeCount(g.FeatureImplementations),
		ComponentImplementationEdges: relationEdgeCount(g.ComponentImplementations),
		ServiceImplementationEdges:   relationEdgeCount(g.ServiceImplementations),
		APIImplementationEdges:       relationEdgeCount(g.APIImplementations),
		BaselineCount:                len(g.Baselines),
		DryRun:                       *dryRun,
	}
	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(summary); err != nil {
			return err
		}
	} else {
		verb := "wrote"
		if *dryRun {
			verb = "would write"
		}
		fmt.Printf("%s graph manifest: %s\n", verb, summary.ManifestPath)
		fmt.Printf("realizations:     %d edge(s)\n", summary.RealizationEdgeCount)
		fmt.Printf("features:         %d edge(s)\n", summary.FeatureImplementationEdges)
		fmt.Printf("components:       %d edge(s)\n", summary.ComponentImplementationEdges)
		fmt.Printf("services:         %d edge(s)\n", summary.ServiceImplementationEdges)
		fmt.Printf("apis:             %d edge(s)\n", summary.APIImplementationEdges)
		fmt.Printf("baselines:        %d\n", summary.BaselineCount)
	}
	if *dryRun {
		return nil
	}
	if err := graph.Write(path, g); err != nil {
		return exitWith(1, "write graph: %v", err)
	}
	return nil
}

func cmdGraphValidate(args []string) error {
	fs := flag.NewFlagSet("graph validate", flag.ContinueOnError)
	manifestPath := fs.String("manifest", "", "path to graph manifest (default: graph_manifest from .specs.yaml)")
	jsonOut := fs.Bool("json", false, "emit machine-readable validation summary")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs graph validate [--manifest <path>] [--json]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return exitWith(2, "unexpected arguments: %v", fs.Args())
	}

	cfg, err := config.Load("")
	if err != nil {
		return err
	}

	path, err := resolveGraphManifestPath(cfg, *manifestPath)
	if err != nil {
		return err
	}
	g, err := graph.Load(path)
	if err != nil {
		return exitWith(1, "validate %s: %v", path, err)
	}
	if err := validateBaselineRepos(g, cfg.Repos); err != nil {
		return exitWith(1, "validate %s: %v", path, err)
	}

	summary := graphValidateJSON{
		ManifestPath:                 path,
		NodeCount:                    len(g.NodeIDs()),
		RealizationEdgeCount:         relationEdgeCount(g.Realizations),
		FeatureImplementationEdges:   relationEdgeCount(g.FeatureImplementations),
		ComponentImplementationEdges: relationEdgeCount(g.ComponentImplementations),
		ServiceImplementationEdges:   relationEdgeCount(g.ServiceImplementations),
		APIImplementationEdges:       relationEdgeCount(g.APIImplementations),
		BaselineCount:                len(g.Baselines),
		LayoutNodeCount:              len(g.Layout),
		RepoCount:                    len(cfg.Repos),
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(summary)
	}

	fmt.Printf("graph valid:      %s\n", summary.ManifestPath)
	fmt.Printf("nodes:            %d\n", summary.NodeCount)
	fmt.Printf("realizations:     %d edge(s)\n", summary.RealizationEdgeCount)
	fmt.Printf("features:         %d edge(s)\n", summary.FeatureImplementationEdges)
	fmt.Printf("components:       %d edge(s)\n", summary.ComponentImplementationEdges)
	fmt.Printf("services:         %d edge(s)\n", summary.ServiceImplementationEdges)
	fmt.Printf("apis:             %d edge(s)\n", summary.APIImplementationEdges)
	fmt.Printf("baselines:        %d\n", summary.BaselineCount)
	fmt.Printf("layout nodes:     %d\n", summary.LayoutNodeCount)
	fmt.Printf("repos checked:    %d\n", summary.RepoCount)
	return nil
}

func resolveGraphManifestPath(cfg *config.Resolved, override string) (string, error) {
	if override == "" {
		return cfg.GraphManifest, nil
	}
	path, err := filepath.Abs(override)
	if err != nil {
		return "", err
	}
	return path, nil
}

func relationEdgeCount(entries []graph.RelationEntry) int {
	total := 0
	for _, entry := range entries {
		total += len(entry.Targets)
	}
	return total
}

func validateBaselineRepos(g *graph.Graph, repos map[string]string) error {
	for index, entry := range g.Baselines {
		if _, ok := repos[entry.Repo]; !ok {
			return fmt.Errorf("baseline entry %d repo %q is not configured in repos", index, entry.Repo)
		}
	}
	return nil
}
