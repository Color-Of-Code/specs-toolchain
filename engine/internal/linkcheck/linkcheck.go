// Package linkcheck verifies bidirectional consistency between requirements
// and the features/components/APIs/services that implement them, and between
// product requirements and the model requirements that realise them.
//
// Conventions in the framework:
//
//   - A product-requirement file (product/<area>/*.md) has a `## Realised By`
//     section listing the model requirements that realise it.
//   - A requirement file (model/requirements/<area>/*.md) has a `## Realises`
//     section listing the product requirements it realises, and an
//     `## Implemented By` section listing the documents that implement it.
//   - A feature/component/api/service file has a `## Requirements` section
//     listing the requirements it implements.
//
// This package finds asymmetric links: a forward edge that lacks a matching
// reverse edge.
package linkcheck

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Result aggregates findings.
type Result struct {
	Errors   []string
	Warnings []string
}

func (r *Result) Failed() bool { return len(r.Errors) > 0 }

func (r *Result) errf(format string, a ...any) {
	r.Errors = append(r.Errors, fmt.Sprintf(format, a...))
}

func (r *Result) warnf(format string, a ...any) {
	r.Warnings = append(r.Warnings, fmt.Sprintf(format, a...))
}

// linkRe matches inline markdown link targets.
var linkRe = regexp.MustCompile(`\]\(([^)]+)\)`)

// sectionHeaderRe matches a level-2 heading.
var sectionHeaderRe = regexp.MustCompile(`(?m)^##\s+(.+?)\s*$`)

// CheckBidirectional walks modelDir and productDir and reports edges that
// lack a reverse:
//
//	product/.../*.md                   ## Realised By      -> requirements
//	model/requirements/.../*.md        ## Realises         -> product
//	model/requirements/.../*.md        ## Implemented By   -> impls
//	model/{features,components,apis,services}/.../*.md  ## Requirements -> reqs
//
// productDir may be empty or non-existent; in that case the product axis is
// skipped silently and only the requirement <-> implementer pair is checked.
func CheckBidirectional(out io.Writer, modelDir, productDir string, r *Result) {
	fmt.Fprintln(out, "== requirement <-> implementer cross-links ==")
	if _, err := os.Stat(modelDir); err != nil {
		r.warnf("model dir %s missing; skipping link check", modelDir)
		return
	}

	type edge struct {
		from string // absolute path to source file
		to   string // absolute path to target file (resolved)
	}
	var fromReqs, fromImpls, fromPRs, fromRealises []edge

	walkRoot := func(root, section string, sink *[]edge) {
		if root == "" {
			return
		}
		if _, err := os.Stat(root); err != nil {
			return
		}
		_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
			if err != nil || info == nil || info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(p, ".md") || filepath.Base(p) == "_index.md" {
				return nil
			}
			data, err := os.ReadFile(p)
			if err != nil {
				return nil
			}
			body := extractSection(string(data), section)
			if body == "" {
				return nil
			}
			for _, target := range linksIn(body) {
				abs := resolveTarget(p, target)
				if abs == "" {
					continue
				}
				*sink = append(*sink, edge{from: p, to: abs})
			}
			return nil
		})
	}

	walkRoot(filepath.Join(modelDir, "requirements"), "Implemented By", &fromReqs)
	walkRoot(filepath.Join(modelDir, "requirements"), "Realises", &fromRealises)
	for _, area := range []string{"features", "components", "apis", "services"} {
		walkRoot(filepath.Join(modelDir, area), "Requirements", &fromImpls)
	}
	walkRoot(productDir, "Realised By", &fromPRs)

	// Build reverse-lookup sets for O(1) symmetry checks.
	indexBy := func(es []edge) map[string]map[string]bool {
		m := make(map[string]map[string]bool)
		for _, e := range es {
			if m[e.from] == nil {
				m[e.from] = map[string]bool{}
			}
			m[e.from][e.to] = true
		}
		return m
	}
	reqHasImpl := indexBy(fromReqs)   // req -> set of impls
	implHasReq := indexBy(fromImpls)  // impl -> set of reqs
	prHasReq := indexBy(fromPRs)      // PR  -> set of reqs
	reqHasPR := indexBy(fromRealises) // req -> set of PRs

	count := 0
	for _, e := range fromReqs {
		// Only enforce reverse edges when target lives inside the model tree
		// and is a known impl-kind (features/components/apis/services).
		if !inImplArea(modelDir, e.to) {
			continue
		}
		if !implHasReq[e.to][e.from] {
			r.errf("missing reverse: %s\n  -> Implemented By links to %s\n  but %s has no `## Requirements` link back to it",
				rel(modelDir, e.from), rel(modelDir, e.to), rel(modelDir, e.to))
			count++
		}
	}
	for _, e := range fromImpls {
		if !inReqArea(modelDir, e.to) {
			continue
		}
		if !reqHasImpl[e.to][e.from] {
			r.errf("missing reverse: %s\n  -> Requirements links to %s\n  but %s has no `## Implemented By` link back to it",
				rel(modelDir, e.from), rel(modelDir, e.to), rel(modelDir, e.to))
			count++
		}
	}
	for _, e := range fromPRs {
		// PR -> req: target must be a model requirement.
		if !inReqArea(modelDir, e.to) {
			continue
		}
		if !reqHasPR[e.to][e.from] {
			r.errf("missing reverse: %s\n  -> Realised By links to %s\n  but %s has no `## Realises` link back to it",
				relAny(productDir, e.from), rel(modelDir, e.to), rel(modelDir, e.to))
			count++
		}
	}
	for _, e := range fromRealises {
		// req -> PR: target must be in productDir.
		if productDir == "" || !inProductArea(productDir, e.to) {
			continue
		}
		if !prHasReq[e.to][e.from] {
			r.errf("missing reverse: %s\n  -> Realises links to %s\n  but %s has no `## Realised By` link back to it",
				rel(modelDir, e.from), relAny(productDir, e.to), relAny(productDir, e.to))
			count++
		}
	}
	if count == 0 {
		fmt.Fprintln(out, "ok")
	}
}

// extractSection returns the body between a `## <name>` heading and the
// next `## ` heading (or EOF). Returns "" if not found.
func extractSection(body, name string) string {
	matches := sectionHeaderRe.FindAllStringSubmatchIndex(body, -1)
	for i, m := range matches {
		title := strings.TrimSpace(body[m[2]:m[3]])
		if !strings.EqualFold(title, name) {
			continue
		}
		end := len(body)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		return body[m[1]:end]
	}
	return ""
}

// linksIn returns inline-link targets in body, skipping URLs and anchors.
func linksIn(body string) []string {
	var out []string
	for _, m := range linkRe.FindAllStringSubmatch(body, -1) {
		t := strings.TrimSpace(m[1])
		if t == "" {
			continue
		}
		switch {
		case strings.HasPrefix(t, "http://"),
			strings.HasPrefix(t, "https://"),
			strings.HasPrefix(t, "mailto:"),
			strings.HasPrefix(t, "#"):
			continue
		}
		// strip fragment / query
		if i := strings.IndexAny(t, "#?"); i >= 0 {
			t = t[:i]
		}
		if t == "" {
			continue
		}
		out = append(out, t)
	}
	return out
}

// resolveTarget makes target absolute relative to fromFile and verifies it
// exists. Returns "" on failure or when the target isn't a regular .md file
// (directories and _index.md files don't carry per-spec sections to verify).
func resolveTarget(fromFile, target string) string {
	if filepath.IsAbs(target) {
		return ""
	}
	abs := filepath.Clean(filepath.Join(filepath.Dir(fromFile), target))
	st, err := os.Stat(abs)
	if err != nil {
		return ""
	}
	if st.IsDir() {
		return ""
	}
	if !strings.HasSuffix(abs, ".md") {
		return ""
	}
	if filepath.Base(abs) == "_index.md" {
		return ""
	}
	return abs
}

func inReqArea(modelDir, p string) bool {
	rel, err := filepath.Rel(modelDir, p)
	if err != nil {
		return false
	}
	return strings.HasPrefix(filepath.ToSlash(rel), "requirements/")
}

func inImplArea(modelDir, p string) bool {
	rel, err := filepath.Rel(modelDir, p)
	if err != nil {
		return false
	}
	r := filepath.ToSlash(rel)
	return strings.HasPrefix(r, "features/") ||
		strings.HasPrefix(r, "components/") ||
		strings.HasPrefix(r, "apis/") ||
		strings.HasPrefix(r, "services/")
}

// inProductArea returns true when p lives anywhere under productDir.
func inProductArea(productDir, p string) bool {
	if productDir == "" {
		return false
	}
	rel, err := filepath.Rel(productDir, p)
	if err != nil {
		return false
	}
	r := filepath.ToSlash(rel)
	return r != "" && !strings.HasPrefix(r, "../")
}

func rel(base, p string) string {
	r, err := filepath.Rel(base, p)
	if err != nil {
		return p
	}
	return r
}

// relAny formats p relative to base when possible, otherwise returns p
// unchanged. Used for product-tree paths in error messages.
func relAny(base, p string) string {
	if base == "" {
		return p
	}
	return rel(base, p)
}
