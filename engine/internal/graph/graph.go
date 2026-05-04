package graph

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const NodeIDFormatRepoRelativeMarkdownPathWithoutExtension = "repo_relative_markdown_path_without_extension"

type PartKind string

const (
	PartKindDeriveReqt PartKind = "deriveReqt"
	PartKindSatisfy    PartKind = "satisfy"
	PartKindRefine     PartKind = "refine"
	PartKindTrace      PartKind = "trace"
	PartKindBaseline   PartKind = "baseline"
	PartKindLayout     PartKind = "layout"
)

type Manifest struct {
	SchemaVersion int              `yaml:"schema_version"`
	NodeIDFormat  string           `yaml:"node_id_format"`
	Parts         []ManifestPart   `yaml:"parts"`
	Generation    GenerationConfig `yaml:"generation"`
}

type ManifestPart struct {
	Name     string   `yaml:"name"`
	File     string   `yaml:"file"`
	Kind     PartKind `yaml:"kind"`
	Required bool     `yaml:"required"`
}

type GenerationConfig struct {
	MarkdownRelationshipFields bool   `yaml:"markdown_relationship_fields"`
	MarkdownBaselineFields     bool   `yaml:"markdown_baseline_fields"`
	StableSort                 string `yaml:"stable_sort"`
}

type RelationEntry struct {
	Source  string   `yaml:"source"`
	Targets []string `yaml:"targets"`
}

type BaselineEntry struct {
	Component string `yaml:"component"`
	Repo      string `yaml:"repo"`
	Path      string `yaml:"path"`
	Commit    string `yaml:"commit"`
}

type LayoutEntry struct {
	ID     string  `yaml:"id"`
	X      float64 `yaml:"x"`
	Y      float64 `yaml:"y"`
	Locked bool    `yaml:"locked,omitempty"`
}

type Graph struct {
	ManifestPath string
	RootDir      string
	Manifest     Manifest
	Relations    map[PartKind][]RelationEntry
	Baselines    []BaselineEntry
	Layout       []LayoutEntry
}

type relationPart struct {
	Kind    PartKind        `yaml:"kind"`
	Entries []RelationEntry `yaml:"entries"`
}

type baselinePart struct {
	Kind    PartKind        `yaml:"kind"`
	Entries []BaselineEntry `yaml:"entries"`
}

type layoutPart struct {
	Kind  PartKind      `yaml:"kind"`
	Nodes []LayoutEntry `yaml:"nodes"`
}

type partSpec struct {
	Name     string
	File     string
	Kind     PartKind
	Required bool
	// Relation-only fields (meaningful when isRelation is true).
	isRelation   bool
	sourceKind   string // storage source kind; empty = any valid node kind
	targetKind   string // storage target kind; empty = any valid node kind
	invertOnSave bool   // frontend sends (target→source); swap before storing
}

// RelationPartSpec is the exported view of a relation part spec.
type RelationPartSpec struct {
	Kind         PartKind
	Required     bool
	SourceKind   string
	TargetKind   string
	InvertOnSave bool
}

// AllRelationSpecs returns metadata for all registered relation parts in manifest order.
func AllRelationSpecs() []RelationPartSpec {
	var specs []RelationPartSpec
	for _, spec := range manifestPartSpecs {
		if spec.isRelation {
			specs = append(specs, RelationPartSpec{
				Kind:         spec.Kind,
				Required:     spec.Required,
				SourceKind:   spec.sourceKind,
				TargetKind:   spec.targetKind,
				InvertOnSave: spec.invertOnSave,
			})
		}
	}
	return specs
}

var manifestPartSpecs = []partSpec{
	{Name: "derive_reqt", File: "deriveReqt.yaml", Kind: PartKindDeriveReqt, Required: true, isRelation: true, sourceKind: "product-requirement", targetKind: "requirement", invertOnSave: true},
	{Name: "refinements", File: "refinements.yaml", Kind: PartKindRefine, Required: true, isRelation: true, sourceKind: "requirement", targetKind: "use-case", invertOnSave: true},
	{Name: "satisfactions", File: "satisfactions.yaml", Kind: PartKindSatisfy, Required: true, isRelation: true, sourceKind: "requirement", targetKind: "component", invertOnSave: true},
	{Name: "traces", File: "traces.yaml", Kind: PartKindTrace, Required: false, isRelation: true},
	{Name: "baselines", File: "baselines.yaml", Kind: PartKindBaseline, Required: false},
	{Name: "layout", File: "layout.yaml", Kind: PartKindLayout, Required: false},
}

var fullSHARe = regexp.MustCompile(`^[0-9a-f]{40}$`)

func Load(manifestPath string) (*Graph, error) {
	absManifest, err := filepath.Abs(manifestPath)
	if err != nil {
		return nil, err
	}

	manifestData, err := os.ReadFile(absManifest)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	var manifest Manifest
	if err := yaml.Unmarshal(manifestData, &manifest); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	if err := validateManifest(manifest); err != nil {
		return nil, err
	}

	g := &Graph{
		ManifestPath: absManifest,
		RootDir:      filepath.Dir(absManifest),
		Manifest:     manifest,
		Relations:    make(map[PartKind][]RelationEntry),
	}

	for _, part := range manifest.Parts {
		if err := g.loadPart(part); err != nil {
			return nil, err
		}
	}

	if err := g.validate(); err != nil {
		return nil, err
	}
	return g, nil
}

func NormalizeNodeID(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("node id is empty")
	}
	trimmed = strings.ReplaceAll(trimmed, "\\", "/")
	trimmed = strings.TrimPrefix(trimmed, "./")
	trimmed = strings.TrimSuffix(trimmed, ".md")
	if strings.HasPrefix(trimmed, "/") {
		return "", fmt.Errorf("node id %q must be repo-relative", raw)
	}
	normalized := path.Clean(trimmed)
	if normalized == "." || normalized == "" {
		return "", fmt.Errorf("node id %q is invalid", raw)
	}
	if strings.HasPrefix(normalized, "../") || normalized == ".." {
		return "", fmt.Errorf("node id %q escapes the repo", raw)
	}
	if kind := KindForNodeID(normalized); kind == "" {
		return "", fmt.Errorf("node id %q has unsupported prefix", raw)
	}
	return normalized, nil
}

func KindForNodeID(id string) string {
	switch {
	case strings.HasPrefix(id, "product/"):
		return "product-requirement"
	case strings.HasPrefix(id, "model/requirements/"):
		return "requirement"
	case strings.HasPrefix(id, "model/use-cases/"):
		return "use-case"
	case strings.HasPrefix(id, "model/components/"):
		return "component"
	default:
		return ""
	}
}

func MarkdownPath(nodeID string) string {
	return nodeID + ".md"
}

func (g *Graph) NodeIDs() []string {
	seen := map[string]struct{}{}
	for _, spec := range manifestPartSpecs {
		if !spec.isRelation {
			continue
		}
		for _, entry := range g.Relations[spec.Kind] {
			seen[entry.Source] = struct{}{}
			for _, target := range entry.Targets {
				seen[target] = struct{}{}
			}
		}
	}
	for _, baseline := range g.Baselines {
		seen[baseline.Component] = struct{}{}
	}
	for _, layout := range g.Layout {
		seen[layout.ID] = struct{}{}
	}
	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func validateManifest(manifest Manifest) error {
	if manifest.SchemaVersion != 1 {
		return fmt.Errorf("schema_version must be 1, got %d", manifest.SchemaVersion)
	}
	if manifest.NodeIDFormat != NodeIDFormatRepoRelativeMarkdownPathWithoutExtension {
		return fmt.Errorf("node_id_format must be %q", NodeIDFormatRepoRelativeMarkdownPathWithoutExtension)
	}
	if !manifest.Generation.MarkdownRelationshipFields {
		return fmt.Errorf("generation.markdown_relationship_fields must be true")
	}
	if !manifest.Generation.MarkdownBaselineFields {
		return fmt.Errorf("generation.markdown_baseline_fields must be true")
	}
	if manifest.Generation.StableSort != "lexical_id" {
		return fmt.Errorf("generation.stable_sort must be %q", "lexical_id")
	}

	seen := map[PartKind]struct{}{}
	maxIndex := -1
	for _, part := range manifest.Parts {
		spec, index, ok := specForKind(part.Kind)
		if !ok {
			return fmt.Errorf("unknown part kind %q", part.Kind)
		}
		if _, exists := seen[part.Kind]; exists {
			return fmt.Errorf("duplicate part kind %q", part.Kind)
		}
		seen[part.Kind] = struct{}{}
		if index <= maxIndex {
			return fmt.Errorf("part %q is out of order", part.Kind)
		}
		maxIndex = index
		if part.Name != spec.Name {
			return fmt.Errorf("part %q must use name %q", part.Kind, spec.Name)
		}
		if part.File != spec.File {
			return fmt.Errorf("part %q must use file %q", part.Kind, spec.File)
		}
		if part.Required != spec.Required {
			return fmt.Errorf("part %q must use required=%t", part.Kind, spec.Required)
		}
		if filepath.Base(part.File) != part.File {
			return fmt.Errorf("part %q file must not contain directories", part.Kind)
		}
	}
	for _, spec := range manifestPartSpecs {
		if spec.Required {
			if _, ok := seen[spec.Kind]; !ok {
				return fmt.Errorf("missing required part %q", spec.Kind)
			}
		}
	}
	return nil
}

func specForKind(kind PartKind) (partSpec, int, bool) {
	for index, spec := range manifestPartSpecs {
		if spec.Kind == kind {
			return spec, index, true
		}
	}
	return partSpec{}, -1, false
}

func (g *Graph) loadPart(part ManifestPart) error {
	partPath := filepath.Join(g.RootDir, part.File)
	data, err := os.ReadFile(partPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", part.File, err)
	}

	for _, spec := range manifestPartSpecs {
		if spec.isRelation && spec.Kind == part.Kind {
			entries, err := loadRelationEntries(data, part.Kind)
			if err != nil {
				return fmt.Errorf("load %s: %w", part.File, err)
			}
			g.Relations[part.Kind] = entries
			return nil
		}
	}
	switch part.Kind {
	case PartKindBaseline:
		entries, err := loadBaselineEntries(data)
		if err != nil {
			return fmt.Errorf("load %s: %w", part.File, err)
		}
		g.Baselines = entries
	case PartKindLayout:
		entries, err := loadLayoutEntries(data)
		if err != nil {
			return fmt.Errorf("load %s: %w", part.File, err)
		}
		g.Layout = entries
	default:
		return fmt.Errorf("unsupported part kind %q", part.Kind)
	}
	return nil
}

func loadRelationEntries(data []byte, kind PartKind) ([]RelationEntry, error) {
	var part relationPart
	if err := yaml.Unmarshal(data, &part); err != nil {
		return nil, err
	}
	if part.Kind != kind {
		return nil, fmt.Errorf("kind must be %q", kind)
	}
	return part.Entries, nil
}

func loadBaselineEntries(data []byte) ([]BaselineEntry, error) {
	var part baselinePart
	if err := yaml.Unmarshal(data, &part); err != nil {
		return nil, err
	}
	if part.Kind != PartKindBaseline {
		return nil, fmt.Errorf("kind must be %q", PartKindBaseline)
	}
	return part.Entries, nil
}

func loadLayoutEntries(data []byte) ([]LayoutEntry, error) {
	var part layoutPart
	if err := yaml.Unmarshal(data, &part); err != nil {
		return nil, err
	}
	if part.Kind != PartKindLayout {
		return nil, fmt.Errorf("kind must be %q", PartKindLayout)
	}
	return part.Nodes, nil
}

func (g *Graph) validate() error {
	for _, spec := range manifestPartSpecs {
		if !spec.isRelation {
			continue
		}
		if err := validateRelationEntries(g.Relations[spec.Kind], spec.Kind, spec.sourceKind, spec.targetKind); err != nil {
			return err
		}
	}
	if err := validateBaselines(g.Baselines); err != nil {
		return err
	}
	if err := validateLayout(g.Layout); err != nil {
		return err
	}
	return nil
}

func validateRelationEntries(entries []RelationEntry, partKind PartKind, sourceKind string, targetKind string) error {
	seenSources := map[string]struct{}{}
	for index, entry := range entries {
		normalizedSource, err := NormalizeNodeID(entry.Source)
		if err != nil {
			return fmt.Errorf("%s entry %d source: %w", partKind, index, err)
		}
		if normalizedSource != entry.Source {
			return fmt.Errorf("%s entry %d source must be normalized as %q", partKind, index, normalizedSource)
		}
		if sourceKind != "" && KindForNodeID(entry.Source) != sourceKind {
			return fmt.Errorf("%s entry %d source %q must be a %s", partKind, index, entry.Source, sourceKind)
		}
		if _, exists := seenSources[entry.Source]; exists {
			return fmt.Errorf("%s entry %d duplicates source %q", partKind, index, entry.Source)
		}
		seenSources[entry.Source] = struct{}{}
		if index > 0 && entries[index-1].Source > entry.Source {
			return fmt.Errorf("%s entries must be sorted by source", partKind)
		}
		if len(entry.Targets) == 0 {
			return fmt.Errorf("%s entry %d targets must not be empty", partKind, index)
		}
		seenTargets := map[string]struct{}{}
		for targetIndex, target := range entry.Targets {
			normalizedTarget, err := NormalizeNodeID(target)
			if err != nil {
				return fmt.Errorf("%s entry %d target %d: %w", partKind, index, targetIndex, err)
			}
			if normalizedTarget != target {
				return fmt.Errorf("%s entry %d target %d must be normalized as %q", partKind, index, targetIndex, normalizedTarget)
			}
			if targetKind != "" && KindForNodeID(target) != targetKind {
				return fmt.Errorf("%s entry %d target %q must be a %s", partKind, index, target, targetKind)
			}
			if _, exists := seenTargets[target]; exists {
				return fmt.Errorf("%s entry %d contains duplicate target %q", partKind, index, target)
			}
			seenTargets[target] = struct{}{}
			if targetIndex > 0 && entry.Targets[targetIndex-1] > target {
				return fmt.Errorf("%s entry %d targets must be sorted", partKind, index)
			}
		}
	}
	return nil
}

func validateBaselines(entries []BaselineEntry) error {
	seenComponents := map[string]struct{}{}
	for index, entry := range entries {
		normalizedComponent, err := NormalizeNodeID(entry.Component)
		if err != nil {
			return fmt.Errorf("baseline entry %d component: %w", index, err)
		}
		if normalizedComponent != entry.Component {
			return fmt.Errorf("baseline entry %d component must be normalized as %q", index, normalizedComponent)
		}
		if KindForNodeID(entry.Component) != "component" {
			return fmt.Errorf("baseline entry %d component %q must be a component", index, entry.Component)
		}
		if _, exists := seenComponents[entry.Component]; exists {
			return fmt.Errorf("baseline entry %d duplicates component %q", index, entry.Component)
		}
		seenComponents[entry.Component] = struct{}{}
		if index > 0 && entries[index-1].Component > entry.Component {
			return fmt.Errorf("baseline entries must be sorted by component")
		}
		if strings.TrimSpace(entry.Repo) == "" {
			return fmt.Errorf("baseline entry %d repo is empty", index)
		}
		if _, err := normalizeBaselinePath(entry.Path); err != nil {
			return fmt.Errorf("baseline entry %d path: %w", index, err)
		}
		if !fullSHARe.MatchString(entry.Commit) {
			return fmt.Errorf("baseline entry %d commit must be a full lowercase SHA", index)
		}
	}
	return nil
}

func validateLayout(entries []LayoutEntry) error {
	seen := map[string]struct{}{}
	for index, entry := range entries {
		normalizedID, err := NormalizeNodeID(entry.ID)
		if err != nil {
			return fmt.Errorf("layout entry %d id: %w", index, err)
		}
		if normalizedID != entry.ID {
			return fmt.Errorf("layout entry %d id must be normalized as %q", index, normalizedID)
		}
		if _, exists := seen[entry.ID]; exists {
			return fmt.Errorf("layout entry %d duplicates id %q", index, entry.ID)
		}
		seen[entry.ID] = struct{}{}
		if index > 0 && entries[index-1].ID > entry.ID {
			return fmt.Errorf("layout entries must be sorted by id")
		}
	}
	return nil
}

func normalizeBaselinePath(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("baseline path is empty")
	}
	trimmed = strings.ReplaceAll(trimmed, "\\", "/")
	if trimmed == "/" {
		return "/", nil
	}
	if strings.HasPrefix(trimmed, "/") {
		return "", fmt.Errorf("baseline path %q must be repo-relative or /", raw)
	}
	normalized := path.Clean(trimmed)
	if normalized == "." || normalized == "" {
		return "", fmt.Errorf("baseline path %q is invalid", raw)
	}
	if strings.HasPrefix(normalized, "../") || normalized == ".." {
		return "", fmt.Errorf("baseline path %q escapes the repo", raw)
	}
	return normalized, nil
}
