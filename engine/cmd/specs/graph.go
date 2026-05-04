package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
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

type graphGenerateJSON struct {
	ManifestPath string `json:"manifest_path"`
	UpdatedFiles int    `json:"updated_files"`
	DryRun       bool   `json:"dry_run"`
}

type graphCacheJSON struct {
	ManifestPath  string `json:"manifest_path"`
	CachePath     string `json:"cache_path"`
	NodeCount     int    `json:"node_count"`
	EdgeCount     int    `json:"edge_count"`
	BaselineCount int    `json:"baseline_count"`
	LayoutCount   int    `json:"layout_count"`
	DryRun        bool   `json:"dry_run"`
}

type graphSaveRelationsJSON struct {
	ManifestPath string `json:"manifest_path"`
	EdgeCount    int    `json:"edge_count"`
}

func cmdGraph(args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: specs graph <validate|import-markdown|generate-markdown|rebuild-cache|save-relations>")
		return exitWith(2, "missing subcommand")
	}
	switch args[0] {
	case "validate":
		return cmdGraphValidate(args[1:])
	case "import-markdown":
		return cmdGraphImportMarkdown(args[1:])
	case "generate-markdown":
		return cmdGraphGenerateMarkdown(args[1:])
	case "rebuild-cache":
		return cmdGraphRebuildCache(args[1:])
	case "save-relations":
		return cmdGraphSaveRelations(args[1:])
	case "-h", "--help", "help":
		fmt.Fprintln(os.Stderr, "Usage: specs graph <validate|import-markdown|generate-markdown|rebuild-cache|save-relations> [flags]")
		return nil
	default:
		return exitWith(2, "unknown subcommand: specs graph %s", args[0])
	}
}

func cmdGraphSaveRelations(args []string) error {
	fs := flag.NewFlagSet("graph save-relations", flag.ContinueOnError)
	manifestPath := fs.String("manifest", "", "path to graph manifest to update (default: graph_manifest from .specs.yaml)")
	inputPath := fs.String("in", "-", "path to JSON relation payload (default: stdin)")
	jsonOut := fs.Bool("json", false, "emit machine-readable save summary")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs graph save-relations [--manifest <path>] [--in <path>|-] [--json]")
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
	traceability, err := graph.Load(path)
	if err != nil {
		return exitWith(1, "load graph %s: %v", path, err)
	}
	payloadData, err := readGraphPayload(*inputPath)
	if err != nil {
		return err
	}
	request, err := decodeRelationSaveRequest(payloadData)
	if err != nil {
		return exitWith(1, "decode relation payload: %v", err)
	}
	relations, err := relationEntriesFromSaveEdges(request.Edges, traceabilityAllowedNodeIDs(traceability))
	if err != nil {
		return exitWith(1, "save relations: %v", err)
	}
	traceability.Realizations = relations.realizations
	traceability.FeatureImplementations = relations.featureImplementations
	traceability.ComponentImplementations = relations.componentImplementations
	traceability.ServiceImplementations = relations.serviceImplementations
	traceability.APIImplementations = relations.apiImplementations
	if err := graph.Write(path, traceability); err != nil {
		return exitWith(1, "write graph: %v", err)
	}
	summary := graphSaveRelationsJSON{ManifestPath: path, EdgeCount: relationEdgeCount(traceability.Realizations) + relationEdgeCount(traceability.FeatureImplementations) + relationEdgeCount(traceability.ComponentImplementations) + relationEdgeCount(traceability.ServiceImplementations) + relationEdgeCount(traceability.APIImplementations)}
	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(summary)
	}
	fmt.Printf("saved relations: %s\n", summary.ManifestPath)
	fmt.Printf("edges:           %d\n", summary.EdgeCount)
	return nil
}
func cmdGraphRebuildCache(args []string) error {
	fs := flag.NewFlagSet("graph rebuild-cache", flag.ContinueOnError)
	manifestPath := fs.String("manifest", "", "path to graph manifest to read (default: graph_manifest from .specs.yaml)")
	cachePath := fs.String("cache", "", "path to SQLite cache file to write (default: graph_cache from .specs.yaml)")
	dryRun := fs.Bool("dry-run", false, "report the cache rebuild summary without writing")
	jsonOut := fs.Bool("json", false, "emit machine-readable cache summary")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs graph rebuild-cache [--manifest <path>] [--cache <path>] [--dry-run] [--json]")
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
	cache, err := resolveGraphCachePath(cfg, *cachePath)
	if err != nil {
		return err
	}
	g, err := graph.Load(path)
	if err != nil {
		return exitWith(1, "load graph %s: %v", path, err)
	}
	if err := validateBaselineRepos(g, cfg.Repos); err != nil {
		return exitWith(1, "validate %s: %v", path, err)
	}
	stats, err := graph.RebuildCache(cache, g, *dryRun)
	if err != nil {
		return exitWith(1, "rebuild cache %s: %v", cache, err)
	}
	summary := graphCacheJSON{
		ManifestPath:  path,
		CachePath:     stats.CachePath,
		NodeCount:     stats.NodeCount,
		EdgeCount:     stats.EdgeCount,
		BaselineCount: stats.BaselineCount,
		LayoutCount:   stats.LayoutCount,
		DryRun:        *dryRun,
	}
	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(summary)
	}
	verb := "rebuilt"
	if *dryRun {
		verb = "would rebuild"
	}
	fmt.Printf("%s cache:     %s\n", verb, summary.CachePath)
	fmt.Printf("graph manifest:   %s\n", summary.ManifestPath)
	fmt.Printf("nodes:            %d\n", summary.NodeCount)
	fmt.Printf("edges:            %d\n", summary.EdgeCount)
	fmt.Printf("baselines:        %d\n", summary.BaselineCount)
	fmt.Printf("layout nodes:     %d\n", summary.LayoutCount)
	return nil
}

func cmdGraphGenerateMarkdown(args []string) error {
	fs := flag.NewFlagSet("graph generate-markdown", flag.ContinueOnError)
	manifestPath := fs.String("manifest", "", "path to graph manifest to read (default: graph_manifest from .specs.yaml)")
	dryRun := fs.Bool("dry-run", false, "report files that would change without writing")
	jsonOut := fs.Bool("json", false, "emit machine-readable generation summary")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs graph generate-markdown [--manifest <path>] [--dry-run] [--json]")
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
		return exitWith(1, "load graph %s: %v", path, err)
	}
	result, err := graph.GenerateMarkdown(cfg.ModelDir, cfg.ProductDir, g, *dryRun)
	if err != nil {
		return exitWith(1, "generate markdown: %v", err)
	}
	summary := graphGenerateJSON{ManifestPath: path, UpdatedFiles: len(result.UpdatedFiles), DryRun: *dryRun}
	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(summary)
	}
	verb := "updated"
	if *dryRun {
		verb = "would update"
	}
	fmt.Printf("%s markdown files: %d\n", verb, summary.UpdatedFiles)
	fmt.Printf("graph manifest:          %s\n", summary.ManifestPath)
	return nil
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
	if err := validateGraphNodeFiles(g, cfg.SpecsRoot); err != nil {
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

func resolveGraphCachePath(cfg *config.Resolved, override string) (string, error) {
	if override == "" {
		return cfg.GraphCache, nil
	}
	path, err := filepath.Abs(override)
	if err != nil {
		return "", err
	}
	return path, nil
}

func readGraphPayload(inputPath string) ([]byte, error) {
	if inputPath == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, exitWith(1, "read graph payload from stdin: %v", err)
		}
		return data, nil
	}
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, exitWith(1, "read graph payload %s: %v", inputPath, err)
	}
	return data, nil
}

func decodeRelationSaveRequest(data []byte) (*relationSaveRequest, error) {
	var request relationSaveRequest
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return nil, err
	}
	return &request, nil
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

func validateGraphNodeFiles(g *graph.Graph, specsRoot string) error {
	for _, nodeID := range g.NodeIDs() {
		path := filepath.Join(specsRoot, filepath.FromSlash(graph.MarkdownPath(nodeID)))
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("node %q points to missing markdown file %s", nodeID, path)
			}
			return fmt.Errorf("stat node %q markdown file %s: %w", nodeID, path, err)
		}
		if info.IsDir() {
			return fmt.Errorf("node %q points to directory %s, want markdown file", nodeID, path)
		}
	}
	return nil
}
