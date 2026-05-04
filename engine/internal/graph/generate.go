package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type GenerateResult struct {
	UpdatedFiles []string
}

type fieldUpdate struct {
	Name  string
	Value string
}

var h1LineRe = regexp.MustCompile(`(?m)^#\s+(.+?)\s*$`)

func GenerateMarkdown(modelDir, productDir string, g *Graph, dryRun bool) (*GenerateResult, error) {
	labels, err := graphLabels(modelDir, productDir, g.NodeIDs())
	if err != nil {
		return nil, err
	}

	realisesByRequirement := invertRelationEntries(g.DeriveReqt)
	implementedByByRequirement := mergeRelationEntries(
		g.Refinements,
		g.Satisfactions,
	)
	requirementsByUseCase := invertRelationEntries(g.Refinements)
	requirementsByComponent := invertRelationEntries(g.Satisfactions)
	baselinesByComponent := map[string]BaselineEntry{}
	for _, entry := range g.Baselines {
		baselinesByComponent[entry.Component] = entry
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
				updated, err := applyTableFieldUpdates(string(body), []fieldUpdate{{
					Name:  "Realised By",
					Value: renderNodeLinks(path, g.DeriveReqt, nodeID, modelDir, productDir, labels),
				}})
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
			updated, err := applyTableFieldUpdates(string(body), []fieldUpdate{
				{Name: "Realises", Value: renderLinksForTargets(path, realisesByRequirement[nodeID], modelDir, productDir, labels)},
				{Name: "Implemented By", Value: renderLinksForTargets(path, implementedByByRequirement[nodeID], modelDir, productDir, labels)},
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
			updates := []fieldUpdate{{
				Name:  "Requirements",
				Value: renderLinksForTargets(path, area.requirements[nodeID], modelDir, productDir, labels),
			}}
			if area.includeBase {
				updates = append(updates, fieldUpdate{Name: "Baseline", Value: renderBaselineField(baselinesByComponent[nodeID])})
			}
			updated, err := applyTableFieldUpdates(string(body), updates)
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

func graphLabels(modelDir, productDir string, nodeIDs []string) (map[string]string, error) {
	labels := make(map[string]string, len(nodeIDs))
	for _, nodeID := range nodeIDs {
		path, err := markdownPathForNodeID(nodeID, modelDir, productDir)
		if err != nil {
			return nil, err
		}
		labels[nodeID] = readMarkdownTitle(path, filepath.Base(path))
	}
	return labels, nil
}

func renderNodeLinks(path string, entries []RelationEntry, nodeID, modelDir, productDir string, labels map[string]string) string {
	for _, entry := range entries {
		if entry.Source == nodeID {
			return renderLinksForTargets(path, entry.Targets, modelDir, productDir, labels)
		}
	}
	return emDashValue()
}

func renderLinksForTargets(path string, targets []string, modelDir, productDir string, labels map[string]string) string {
	if len(targets) == 0 {
		return emDashValue()
	}
	links := make([]string, 0, len(targets))
	for _, target := range targets {
		targetPath, err := markdownPathForNodeID(target, modelDir, productDir)
		if err != nil {
			return emDashValue()
		}
		rel, err := filepath.Rel(filepath.Dir(path), targetPath)
		if err != nil {
			return emDashValue()
		}
		links = append(links, fmt.Sprintf("[%s](%s)", labels[target], filepath.ToSlash(rel)))
	}
	return strings.Join(links, ", ")
}

func renderBaselineField(entry BaselineEntry) string {
	if entry.Component == "" {
		return emDashValue()
	}
	return fmt.Sprintf("`%s:%s@%s`", entry.Repo, entry.Path, entry.Commit)
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

func applyTableFieldUpdates(body string, updates []fieldUpdate) (string, error) {
	lines := strings.Split(body, "\n")
	tableStart := -1
	tableEnd := -1
	for index, line := range lines {
		parts, ok := splitTableLine(line)
		if !ok {
			continue
		}
		if len(parts) >= 2 && strings.EqualFold(strings.TrimSpace(parts[1]), "Field") {
			tableStart = index
			tableEnd = index + 1
			for tableEnd < len(lines) {
				if _, ok := splitTableLine(lines[tableEnd]); !ok {
					break
				}
				tableEnd++
			}
			break
		}
	}
	if tableStart < 0 || tableEnd-tableStart < 2 {
		return body, fmt.Errorf("field table not found")
	}

	rowIndices := map[string]int{}
	for index := tableStart + 2; index < tableEnd; index++ {
		parts, ok := splitTableLine(lines[index])
		if !ok || len(parts) < 3 {
			continue
		}
		rowIndices[strings.TrimSpace(parts[1])] = index
	}

	for updateIndex, update := range updates {
		if rowIndex, ok := rowIndices[update.Name]; ok {
			lines[rowIndex] = replaceTableValue(lines[rowIndex], update.Value)
			continue
		}
		insertAt := tableEnd
		for nextIndex := updateIndex + 1; nextIndex < len(updates); nextIndex++ {
			if rowIndex, ok := rowIndices[updates[nextIndex].Name]; ok {
				insertAt = rowIndex
				break
			}
		}
		newLine := fmt.Sprintf("| %s | %s |", update.Name, update.Value)
		lines = append(lines[:insertAt], append([]string{newLine}, lines[insertAt:]...)...)
		for name, rowIndex := range rowIndices {
			if rowIndex >= insertAt {
				rowIndices[name] = rowIndex + 1
			}
		}
		rowIndices[update.Name] = insertAt
		tableEnd++
	}

	return strings.Join(lines, "\n"), nil
}

func splitTableLine(line string) ([]string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "|") {
		return nil, false
	}
	parts := strings.Split(line, "|")
	if len(parts) < 4 {
		return nil, false
	}
	return parts, true
}

func replaceTableValue(line, value string) string {
	parts, ok := splitTableLine(line)
	if !ok {
		return line
	}
	parts[2] = replaceCellValue(parts[2], value)
	return strings.Join(parts, "|")
}

func replaceCellValue(cell, value string) string {
	leading := ""
	for i := 0; i < len(cell) && (cell[i] == ' ' || cell[i] == '\t'); i++ {
		leading += string(cell[i])
	}
	trailing := ""
	for i := len(cell) - 1; i >= 0 && (cell[i] == ' ' || cell[i] == '\t'); i-- {
		trailing = string(cell[i]) + trailing
	}
	if leading == "" {
		leading = " "
	}
	if trailing == "" {
		trailing = " "
	}
	return leading + value + trailing
}

func readMarkdownTitle(path, fallback string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return strings.TrimSuffix(fallback, ".md")
	}
	if match := h1LineRe.FindStringSubmatch(string(data)); len(match) == 2 {
		return strings.TrimSpace(match[1])
	}
	return strings.TrimSuffix(fallback, ".md")
}

func emDashValue() string {
	return "—"
}
