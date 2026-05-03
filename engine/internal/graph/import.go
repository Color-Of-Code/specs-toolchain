package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	inlineLinkRe      = regexp.MustCompile(`\]\(([^)]+)\)`)
	sectionHeaderRe   = regexp.MustCompile(`(?m)^##\s+(.+?)\s*$`)
	baselineRowLinkRe = regexp.MustCompile(`^\|\s*\[`)
)

func ImportMarkdown(modelDir, productDir, baselinesFile string) (*Graph, error) {
	if _, err := os.Stat(modelDir); err != nil {
		return nil, fmt.Errorf("model dir %s: %w", modelDir, err)
	}

	graphData := &Graph{}
	realizations := map[string]map[string]struct{}{}
	featureImplementations := map[string]map[string]struct{}{}
	componentImplementations := map[string]map[string]struct{}{}
	serviceImplementations := map[string]map[string]struct{}{}
	apiImplementations := map[string]map[string]struct{}{}

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
				case "feature":
					addEdge(featureImplementations, source, target)
				case "component":
					addEdge(componentImplementations, source, target)
				case "service":
					addEdge(serviceImplementations, source, target)
				case "api":
					addEdge(apiImplementations, source, target)
				default:
					return fmt.Errorf("implemented-by target %q from %s is not a supported implementer kind", target, path)
				}
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
		{root: filepath.Join(modelDir, "features"), edges: featureImplementations},
		{root: filepath.Join(modelDir, "components"), edges: componentImplementations},
		{root: filepath.Join(modelDir, "services"), edges: serviceImplementations},
		{root: filepath.Join(modelDir, "apis"), edges: apiImplementations},
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

	if baselinesFile != "" {
		entries, err := importBaselinesMarkdown(baselinesFile, modelDir, productDir)
		if err != nil {
			return nil, err
		}
		graphData.Baselines = entries
	}

	graphData.Realizations = relationEntriesFromMap(realizations)
	graphData.FeatureImplementations = relationEntriesFromMap(featureImplementations)
	graphData.ComponentImplementations = relationEntriesFromMap(componentImplementations)
	graphData.ServiceImplementations = relationEntriesFromMap(serviceImplementations)
	graphData.APIImplementations = relationEntriesFromMap(apiImplementations)
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
		"realizations.yaml": relationPart{
			Kind:    PartKindRealization,
			Entries: g.Realizations,
		},
		"feature_implementations.yaml": relationPart{
			Kind:    PartKindFeatureImplementation,
			Entries: g.FeatureImplementations,
		},
		"component_implementations.yaml": relationPart{
			Kind:    PartKindComponentImplementation,
			Entries: g.ComponentImplementations,
		},
		"service_implementations.yaml": relationPart{
			Kind:    PartKindServiceImplementation,
			Entries: g.ServiceImplementations,
		},
		"api_implementations.yaml": relationPart{
			Kind:    PartKindAPIImplementation,
			Entries: g.APIImplementations,
		},
	}
	if len(g.Baselines) > 0 {
		toWrite["baselines.yaml"] = baselinePart{Kind: PartKindBaseline, Entries: g.Baselines}
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
	for _, optional := range []string{"baselines.yaml", "layout.yaml"} {
		if _, ok := toWrite[optional]; ok {
			continue
		}
		if err := os.Remove(filepath.Join(manifestDir, optional)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func manifestForGraph(g *Graph) Manifest {
	parts := make([]ManifestPart, 0, len(manifestPartSpecs))
	for _, spec := range manifestPartSpecs {
		if !spec.Required {
			switch spec.Kind {
			case PartKindBaseline:
				if len(g.Baselines) == 0 {
					continue
				}
			case PartKindLayout:
				if len(g.Layout) == 0 {
					continue
				}
			}
		}
		parts = append(parts, ManifestPart{Name: spec.Name, File: spec.File, Kind: spec.Kind, Required: spec.Required})
	}
	return Manifest{
		SchemaVersion: 1,
		NodeIDFormat:  NodeIDFormatRepoRelativeMarkdownPathWithoutExtension,
		Parts:         parts,
		Generation: GenerationConfig{
			MarkdownRelationshipFields: true,
			MarkdownBaselineFields:     true,
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
	body := string(data)
	seen := map[string]struct{}{}
	values := []string{extractFieldValue(body, field), extractSectionBody(body, field)}
	for _, value := range values {
		for _, target := range inlineLinkTargets(value) {
			abs, err := resolveMarkdownTarget(path, target)
			if err != nil {
				return nil, err
			}
			nodeID, err := nodeIDForMarkdownPath(abs, modelDir, productDir)
			if err != nil {
				return nil, err
			}
			seen[nodeID] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for target := range seen {
		out = append(out, target)
	}
	sort.Strings(out)
	return out, nil
}

func extractFieldValue(body, field string) string {
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "|") {
			continue
		}
		parts := strings.Split(trimmed, "|")
		if len(parts) < 4 {
			continue
		}
		key := strings.TrimSpace(parts[1])
		if !strings.EqualFold(key, field) {
			continue
		}
		return strings.TrimSpace(parts[2])
	}
	return ""
}

func extractSectionBody(body, field string) string {
	matches := sectionHeaderRe.FindAllStringSubmatchIndex(body, -1)
	for index, match := range matches {
		title := strings.TrimSpace(body[match[2]:match[3]])
		if !strings.EqualFold(title, field) {
			continue
		}
		end := len(body)
		if index+1 < len(matches) {
			end = matches[index+1][0]
		}
		return body[match[1]:end]
	}
	return ""
}

func inlineLinkTargets(body string) []string {
	var targets []string
	for _, match := range inlineLinkRe.FindAllStringSubmatch(body, -1) {
		target := strings.TrimSpace(match[1])
		if target == "" {
			continue
		}
		if i := strings.IndexAny(target, "#?"); i >= 0 {
			target = target[:i]
		}
		if target == "" || strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") || strings.HasPrefix(target, "mailto:") {
			continue
		}
		targets = append(targets, target)
	}
	return targets
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

func importBaselinesMarkdown(path, modelDir, productDir string) ([]BaselineEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var entries []BaselineEntry
	seen := map[string]struct{}{}
	inComponents := false
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			inComponents = strings.Contains(trimmed, "Components")
			continue
		}
		if !inComponents || !baselineRowLinkRe.MatchString(line) {
			continue
		}
		cells := strings.Split(strings.TrimRight(line, "\r"), "|")
		if len(cells) < 6 {
			continue
		}
		componentLinks := inlineLinkTargets(cells[1])
		if len(componentLinks) == 0 {
			continue
		}
		componentPath, err := resolveMarkdownTarget(path, componentLinks[0])
		if err != nil {
			return nil, err
		}
		componentID, err := nodeIDForMarkdownPath(componentPath, modelDir, productDir)
		if err != nil {
			return nil, err
		}
		repo := strings.Trim(strings.TrimSpace(cells[2]), "`")
		repoPath := strings.Trim(strings.TrimSpace(cells[3]), "`")
		commit := strings.Trim(strings.TrimSpace(cells[4]), "`")
		if repoPath == "" || strings.HasPrefix(repoPath, "_") || strings.HasPrefix(repoPath, "(") {
			continue
		}
		entry := BaselineEntry{Component: componentID, Repo: repo, Path: repoPath, Commit: commit}
		key := entry.Component + "\x00" + entry.Repo + "\x00" + entry.Path
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Component < entries[j].Component })
	return entries, nil
}
