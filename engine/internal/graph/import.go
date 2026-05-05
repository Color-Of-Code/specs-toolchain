package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

func ImportMarkdown(modelDir, productDir string) (*Graph, error) {
	if _, err := os.Stat(modelDir); err != nil {
		return nil, fmt.Errorf("model dir %s: %w", modelDir, err)
	}

	graphData := &Graph{
		Relations: make(map[PartKind][]RelationEntry),
	}
	realizations := map[string]map[string]struct{}{}
	satisfactions := map[string]map[string]struct{}{}
	refinements := map[string]map[string]struct{}{}
	traces := map[string]map[string]struct{}{}

	addEdge := func(edges map[string]map[string]struct{}, source, target string) {
		if edges[source] == nil {
			edges[source] = map[string]struct{}{}
		}
		edges[source][target] = struct{}{}
	}

	if productDir != "" {
		if _, err := os.Stat(productDir); err == nil {
			if err := walkMarkdownFiles(productDir, func(path string) error {
				targets, err := linkedNodeIDs(path, modelDir, productDir, "Realised By")
				if err != nil {
					return err
				}
				if len(targets) == 0 {
					return nil
				}
				source, err := nodeIDForMarkdownPath(path, modelDir, productDir)
				if err != nil {
					return err
				}
				for _, target := range targets {
					addEdge(realizations, source, target)
				}
				return nil
			}); err != nil {
				return nil, err
			}
		}
	}

	requirementsDir := filepath.Join(modelDir, "requirements")
	if _, err := os.Stat(requirementsDir); err == nil {
		if err := walkMarkdownFiles(requirementsDir, func(path string) error {
			source, err := nodeIDForMarkdownPath(path, modelDir, productDir)
			if err != nil {
				return err
			}

			realisesTargets, err := linkedNodeIDs(path, modelDir, productDir, "Realises")
			if err != nil {
				return err
			}
			for _, target := range realisesTargets {
				addEdge(realizations, target, source)
			}

			implementedByTargets, err := linkedNodeIDs(path, modelDir, productDir, "Implemented By")
			if err != nil {
				return err
			}
			for _, target := range implementedByTargets {
				switch KindForNodeID(target) {
				case "use-case":
					addEdge(refinements, source, target)
				case "component":
					addEdge(satisfactions, source, target)
				default:
					return fmt.Errorf("implemented-by target %q from %s is not a supported implementer kind", target, path)
				}
			}
			// refines: specific_req → abstract_req (stored as abstract → specific)
			refinesTargets, err := linkedNodeIDs(path, modelDir, productDir, "Refines")
			if err != nil {
				return err
			}
			for _, target := range refinesTargets {
				if KindForNodeID(target) != "requirement" {
					return fmt.Errorf("refines target %q from %s must be a requirement", target, path)
				}
				// source=current req (specific), target=abstract req
				// storage convention: source=abstract, target=specific
				addEdge(refinements, target, source)
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}

	for _, area := range []struct {
		root  string
		edges map[string]map[string]struct{}
	}{
		{root: filepath.Join(modelDir, "use-cases"), edges: refinements},
		{root: filepath.Join(modelDir, "components"), edges: satisfactions},
	} {
		if _, err := os.Stat(area.root); err != nil {
			continue
		}
		if err := walkMarkdownFiles(area.root, func(path string) error {
			source, err := nodeIDForMarkdownPath(path, modelDir, productDir)
			if err != nil {
				return err
			}
			requirements, err := linkedNodeIDs(path, modelDir, productDir, "Requirements")
			if err != nil {
				return err
			}
			for _, requirement := range requirements {
				addEdge(area.edges, requirement, source)
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}

	// Collect Trace relations from all artifact directories.
	for _, area := range []string{
		productDir,
		filepath.Join(modelDir, "requirements"),
		filepath.Join(modelDir, "use-cases"),
		filepath.Join(modelDir, "components"),
	} {
		if area == "" {
			continue
		}
		if _, err := os.Stat(area); err != nil {
			continue
		}
		if err := walkMarkdownFiles(area, func(path string) error {
			targets, err := linkedNodeIDs(path, modelDir, productDir, "Traces")
			if err != nil {
				return err
			}
			if len(targets) == 0 {
				return nil
			}
			source, err := nodeIDForMarkdownPath(path, modelDir, productDir)
			if err != nil {
				return err
			}
			for _, target := range targets {
				addEdge(traces, source, target)
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}

	graphData.Relations[PartKindDeriveReqt] = relationEntriesFromMap(realizations)
	graphData.Relations[PartKindRefine] = relationEntriesFromMap(refinements)
	graphData.Relations[PartKindSatisfy] = relationEntriesFromMap(satisfactions)
	graphData.Relations[PartKindTrace] = relationEntriesFromMap(traces)
	graphData.Manifest = manifestForGraph(graphData)
	if err := graphData.validate(); err != nil {
		return nil, err
	}
	return graphData, nil
}

func Write(manifestPath string, g *Graph) error {
	if g == nil {
		return fmt.Errorf("graph is nil")
	}
	absManifest, err := filepath.Abs(manifestPath)
	if err != nil {
		return err
	}
	manifestDir := filepath.Dir(absManifest)
	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		return err
	}

	manifest := manifestForGraph(g)
	toWrite := map[string]any{
		"graph.yaml": manifest,
	}
	for _, spec := range manifestPartSpecs {
		if !spec.isRelation {
			continue
		}
		entries := g.Relations[spec.Kind]
		if !spec.Required && len(entries) == 0 {
			continue
		}
		toWrite[spec.File] = relationPart{Kind: spec.Kind, Entries: entries}
	}
	if len(g.Layout) > 0 {
		toWrite["layout.yaml"] = layoutPart{Kind: PartKindLayout, Nodes: g.Layout}
	}

	for name, value := range toWrite {
		data, err := yaml.Marshal(value)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(manifestDir, name), data, 0o644); err != nil {
			return err
		}
	}
	for _, optional := range []string{"layout.yaml"} {
		if _, ok := toWrite[optional]; ok {
			continue
		}
		if err := os.Remove(filepath.Join(manifestDir, optional)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	// Remove optional relation files when they have no entries.
	for _, spec := range manifestPartSpecs {
		if spec.Required || !spec.isRelation {
			continue
		}
		if _, ok := toWrite[spec.File]; ok {
			continue
		}
		if err := os.Remove(filepath.Join(manifestDir, spec.File)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func manifestForGraph(g *Graph) Manifest {
	parts := make([]ManifestPart, 0, len(manifestPartSpecs))
	for _, spec := range manifestPartSpecs {
		if spec.isRelation {
			if spec.Required || len(g.Relations[spec.Kind]) > 0 {
				parts = append(parts, ManifestPart{Kind: spec.Kind, Required: spec.Required})
			}
			continue
		}
		if !spec.Required {
			switch spec.Kind {
			case PartKindLayout:
				if len(g.Layout) == 0 {
					continue
				}
			}
		}
		parts = append(parts, ManifestPart{Kind: spec.Kind, Required: spec.Required})
	}
	return Manifest{
		SchemaVersion: 1,
		NodeIDFormat:  NodeIDFormatRepoRelativeMarkdownPathWithoutExtension,
		Parts:         parts,
		Generation: GenerationConfig{
			MarkdownRelationshipFields: true,
			StableSort:                 "lexical_id",
		},
	}
}

func relationEntriesFromMap(edges map[string]map[string]struct{}) []RelationEntry {
	if len(edges) == 0 {
		return []RelationEntry{}
	}
	sources := make([]string, 0, len(edges))
	for source := range edges {
		sources = append(sources, source)
	}
	sort.Strings(sources)
	entries := make([]RelationEntry, 0, len(sources))
	for _, source := range sources {
		targets := make([]string, 0, len(edges[source]))
		for target := range edges[source] {
			targets = append(targets, target)
		}
		sort.Strings(targets)
		entries = append(entries, RelationEntry{Source: source, Targets: targets})
	}
	return entries
}

func walkMarkdownFiles(root string, visit func(path string) error) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return err
		}
		if !strings.HasSuffix(path, ".md") || filepath.Base(path) == "_index.md" {
			return nil
		}
		return visit(path)
	})
}

func linkedNodeIDs(path, modelDir, productDir, field string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	fmBytes, _, ok := splitFrontmatter(data)
	if !ok {
		// No frontmatter means no relational fields.
		return nil, nil
	}
	key, ok := fmFieldKeys[field]
	if !ok {
		return nil, nil
	}
	relPaths, err := frontmatterStringList(fmBytes, key)
	if err != nil {
		return nil, fmt.Errorf("%s: field %q: %w", path, field, err)
	}
	seen := map[string]struct{}{}
	for _, relPath := range relPaths {
		abs, err := resolveMarkdownTarget(path, relPath)
		if err != nil {
			return nil, err
		}
		nodeID, err := nodeIDForMarkdownPath(abs, modelDir, productDir)
		if err != nil {
			return nil, err
		}
		seen[nodeID] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	sort.Strings(out)
	return out, nil
}

func resolveMarkdownTarget(fromPath, target string) (string, error) {
	if filepath.IsAbs(target) {
		return "", fmt.Errorf("absolute markdown target %q in %s is not supported", target, fromPath)
	}
	abs := filepath.Clean(filepath.Join(filepath.Dir(fromPath), filepath.FromSlash(target)))
	st, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("resolve %q from %s: %w", target, fromPath, err)
	}
	if st.IsDir() || !strings.HasSuffix(abs, ".md") || filepath.Base(abs) == "_index.md" {
		return "", fmt.Errorf("target %q from %s must resolve to a markdown file", target, fromPath)
	}
	return abs, nil
}

func nodeIDForMarkdownPath(path, modelDir, productDir string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if productDir != "" {
		if productAbs, err := filepath.Abs(productDir); err == nil {
			if rel, ok := withinRoot(productAbs, abs); ok {
				return NormalizeNodeID(filepath.ToSlash(filepath.Join("product", rel[:len(rel)-len(filepath.Ext(rel))])))
			}
		}
	}
	modelAbs, err := filepath.Abs(modelDir)
	if err != nil {
		return "", err
	}
	if rel, ok := withinRoot(modelAbs, abs); ok {
		return NormalizeNodeID(filepath.ToSlash(filepath.Join("model", rel[:len(rel)-len(filepath.Ext(rel))])))
	}
	return "", fmt.Errorf("path %s is outside model and product roots", path)
}

func withinRoot(root, target string) (string, bool) {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return "", false
	}
	if rel == "." || strings.HasPrefix(rel, "..") {
		return "", false
	}
	return rel, true
}
