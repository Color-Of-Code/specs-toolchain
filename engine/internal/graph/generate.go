package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type GenerateResult struct {
	UpdatedFiles []string
}

func GenerateMarkdown(modelDir, productDir string, g *Graph, dryRun bool) (*GenerateResult, error) {
	realisesByRequirement := invertRelationEntries(g.Relations[PartKindDeriveReqt])
	implementedByByRequirement := mergeRelationEntries(
		g.Relations[PartKindRefine],
		g.Relations[PartKindSatisfy],
	)
	requirementsByUseCase := invertRelationEntries(g.Relations[PartKindRefine])
	requirementsByComponent := invertRelationEntries(g.Relations[PartKindSatisfy])
	baselinesByComponent := map[string]BaselineEntry{}
	for _, entry := range g.Baselines {
		baselinesByComponent[entry.Component] = entry
	}
	traceTargets := map[string][]string{}
	for _, entry := range g.Relations[PartKindTrace] {
		traceTargets[entry.Source] = entry.Targets
	}

	result := &GenerateResult{}
	if productDir != "" {
		if _, err := os.Stat(productDir); err == nil {
			if err := walkMarkdownFiles(productDir, func(path string) error {
				nodeID, err := nodeIDForMarkdownPath(path, modelDir, productDir)
				if err != nil {
					return err
				}
				body, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				updated, err := applyFMUpdates(string(body), []fmUpdate{
					{key: "realised_by", paths: pathsForNodeLinks(path, g.Relations[PartKindDeriveReqt], nodeID, modelDir, productDir)},
					{key: "traces", paths: pathsForTargets(path, traceTargets[nodeID], modelDir, productDir), omitIfAbsent: true},
				})
				if err != nil {
					return fmt.Errorf("update %s: %w", path, err)
				}
				return maybeWriteUpdatedFile(path, string(body), updated, dryRun, result)
			}); err != nil {
				return nil, err
			}
		}
	}

	requirementsDir := filepath.Join(modelDir, "requirements")
	if _, err := os.Stat(requirementsDir); err == nil {
		if err := walkMarkdownFiles(requirementsDir, func(path string) error {
			nodeID, err := nodeIDForMarkdownPath(path, modelDir, productDir)
			if err != nil {
				return err
			}
			body, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			updated, err := applyFMUpdates(string(body), []fmUpdate{
				{key: "realises", paths: pathsForTargets(path, realisesByRequirement[nodeID], modelDir, productDir)},
				{key: "implemented_by", paths: pathsForTargets(path, implementedByByRequirement[nodeID], modelDir, productDir)},
				{key: "traces", paths: pathsForTargets(path, traceTargets[nodeID], modelDir, productDir), omitIfAbsent: true},
			})
			if err != nil {
				return fmt.Errorf("update %s: %w", path, err)
			}
			return maybeWriteUpdatedFile(path, string(body), updated, dryRun, result)
		}); err != nil {
			return nil, err
		}
	}

	for _, area := range []struct {
		root         string
		requirements map[string][]string
		includeBase  bool
	}{
		{root: filepath.Join(modelDir, "use-cases"), requirements: requirementsByUseCase},
		{root: filepath.Join(modelDir, "components"), requirements: requirementsByComponent, includeBase: true},
	} {
		if _, err := os.Stat(area.root); err != nil {
			continue
		}
		if err := walkMarkdownFiles(area.root, func(path string) error {
			nodeID, err := nodeIDForMarkdownPath(path, modelDir, productDir)
			if err != nil {
				return err
			}
			body, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			updates := []fmUpdate{
				{key: "requirements", paths: pathsForTargets(path, area.requirements[nodeID], modelDir, productDir)},
				{key: "traces", paths: pathsForTargets(path, traceTargets[nodeID], modelDir, productDir), omitIfAbsent: true},
			}
			if area.includeBase {
				updates = append(updates, fmUpdate{
					key:      "baseline",
					scalar:   renderBaselineScalar(baselinesByComponent[nodeID]),
					isScalar: true,
				})
			}
			updated, err := applyFMUpdates(string(body), updates)
			if err != nil {
				return fmt.Errorf("update %s: %w", path, err)
			}
			return maybeWriteUpdatedFile(path, string(body), updated, dryRun, result)
		}); err != nil {
			return nil, err
		}
	}

	sort.Strings(result.UpdatedFiles)
	return result, nil
}

func maybeWriteUpdatedFile(path, oldBody, newBody string, dryRun bool, result *GenerateResult) error {
	if newBody == oldBody {
		return nil
	}
	result.UpdatedFiles = append(result.UpdatedFiles, path)
	if dryRun {
		return nil
	}
	return os.WriteFile(path, []byte(newBody), 0o644)
}

// pathsForNodeLinks returns relative markdown paths for the targets of the
// entry whose Source matches nodeID in entries. Returns nil when not found.
func pathsForNodeLinks(fromPath string, entries []RelationEntry, nodeID, modelDir, productDir string) []string {
	for _, entry := range entries {
		if entry.Source == nodeID {
			return pathsForTargets(fromPath, entry.Targets, modelDir, productDir)
		}
	}
	return nil
}

// pathsForTargets converts a slice of graph node IDs to relative markdown paths
// (from fromPath's directory).  Nodes that cannot be resolved are silently skipped.
func pathsForTargets(fromPath string, targets []string, modelDir, productDir string) []string {
	if len(targets) == 0 {
		return nil
	}
	paths := make([]string, 0, len(targets))
	for _, target := range targets {
		targetPath, err := markdownPathForNodeID(target, modelDir, productDir)
		if err != nil {
			continue
		}
		rel, err := filepath.Rel(filepath.Dir(fromPath), targetPath)
		if err != nil {
			continue
		}
		paths = append(paths, filepath.ToSlash(rel))
	}
	return paths
}

func renderBaselineScalar(entry BaselineEntry) string {
	if entry.Component == "" {
		return "~"
	}
	return fmt.Sprintf("%s:%s@%s", entry.Repo, entry.Path, entry.Commit)
}

func markdownPathForNodeID(nodeID, modelDir, productDir string) (string, error) {
	switch {
	case strings.HasPrefix(nodeID, "product/"):
		if productDir == "" {
			return "", fmt.Errorf("node %q requires product dir", nodeID)
		}
		return filepath.Join(productDir, filepath.FromSlash(strings.TrimPrefix(MarkdownPath(nodeID), "product/"))), nil
	case strings.HasPrefix(nodeID, "model/"):
		return filepath.Join(modelDir, filepath.FromSlash(strings.TrimPrefix(MarkdownPath(nodeID), "model/"))), nil
	default:
		return "", fmt.Errorf("unsupported node id %q", nodeID)
	}
}

func invertRelationEntries(entries []RelationEntry) map[string][]string {
	inverted := map[string][]string{}
	for _, entry := range entries {
		for _, target := range entry.Targets {
			inverted[target] = append(inverted[target], entry.Source)
		}
	}
	for key := range inverted {
		sort.Strings(inverted[key])
	}
	return inverted
}

func mergeRelationEntries(groups ...[]RelationEntry) map[string][]string {
	merged := map[string]map[string]struct{}{}
	for _, group := range groups {
		for _, entry := range group {
			if merged[entry.Source] == nil {
				merged[entry.Source] = map[string]struct{}{}
			}
			for _, target := range entry.Targets {
				merged[entry.Source][target] = struct{}{}
			}
		}
	}
	out := map[string][]string{}
	for source, targets := range merged {
		out[source] = make([]string, 0, len(targets))
		for target := range targets {
			out[source] = append(out[source], target)
		}
		sort.Strings(out[source])
	}
	return out
}
