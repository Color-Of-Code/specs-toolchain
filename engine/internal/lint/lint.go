// Package lint ports the framework lint checks to Go. It exposes
// modular check functions and a Result type so callers can compose modes
// (--all, --links, --style).
package lint

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Result aggregates findings from a lint run.
type Result struct {
	Errors   []string
	Warnings []string
}

// Failed reports whether any error-level issues were recorded.
func (r *Result) Failed() bool { return len(r.Errors) > 0 }

func (r *Result) errf(format string, a ...any) {
	r.Errors = append(r.Errors, fmt.Sprintf(format, a...))
}
func (r *Result) warnf(format string, a ...any) {
	r.Warnings = append(r.Warnings, fmt.Sprintf(format, a...))
}

// Excluded path components (relative to specs root) that mirror the bash
// lint script and the style config ignore list. Matches at any depth
// (e.g. extension/node_modules/...).
var excludedPathComponents = map[string]struct{}{
	".framework":   {},
	".lint":        {},
	"node_modules": {},
	".git":         {},
	"dist":         {},
}

func isExcludedRel(rel string) bool {
	rel = filepath.ToSlash(rel)
	for _, part := range strings.Split(rel, "/") {
		if _, ok := excludedPathComponents[part]; ok {
			return true
		}
	}
	return false
}

// CheckSymlinks finds broken symlinks under specsRoot, mirroring
// `find -xtype l` semantics.
func CheckSymlinks(out io.Writer, specsRoot string, r *Result) {
	fmt.Fprintln(out, "== broken symlinks ==")
	count := 0
	err := filepath.Walk(specsRoot, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			// Skip unreadable trees rather than aborting.
			return nil
		}
		rel, _ := filepath.Rel(specsRoot, path)
		if rel == "." {
			return nil
		}
		if isExcludedRel(rel) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			if _, err := os.Stat(path); err != nil {
				r.errf("broken symlink: %s", rel)
				count++
			}
		}
		return nil
	})
	if err != nil {
		r.errf("walk specs root: %v", err)
		return
	}
	if count == 0 {
		fmt.Fprintln(out, "ok")
	}
}

// linkRe extracts inline markdown link targets like ](target).
var linkRe = regexp.MustCompile(`\]\(([^)]+)\)`)

// CheckMarkdownLinks verifies that every relative markdown link target
// resolves to an existing file or directory.
func CheckMarkdownLinks(out io.Writer, specsRoot string, r *Result) {
	fmt.Fprintln(out, "== markdown link targets ==")
	count := 0
	err := filepath.Walk(specsRoot, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("%s: %w", path, walkErr)
		}
		rel, _ := filepath.Rel(specsRoot, path)
		if rel == "." {
			return nil
		}
		if isExcludedRel(rel) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		// Skip the change-request template subtree as in the bash version.
		if strings.HasPrefix(filepath.ToSlash(rel), "change-requests/_template/") {
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}
		// Skip symlinks: bash `find -type f` does not match them, and the
		// broken-symlink check is already a separate pass.
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		if filepath.Base(path) == "_template.md" {
			return nil
		}
		broken, openErr := linksInFile(path)
		if openErr != nil {
			r.errf("%s: %v", rel, openErr)
			return nil
		}
		baseDir := filepath.Dir(path)
		for _, target := range broken {
			clean := stripFragment(target)
			if clean == "" || strings.HasPrefix(clean, "/") {
				continue
			}
			abs := filepath.Join(baseDir, clean)
			if _, err := os.Stat(abs); err != nil {
				r.errf("broken link: %s -> %s", rel, target)
				count++
			}
		}
		return nil
	})
	if err != nil {
		r.errf("walk specs root: %v", err)
		return
	}
	if count == 0 {
		fmt.Fprintln(out, "ok")
	}
}

// linksInFile returns the targets of every relative inline markdown link
// in path, skipping fenced code blocks. URL-only and anchor-only targets
// are filtered out.
func linksInFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var targets []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	fenced := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			fenced = !fenced
			continue
		}
		if fenced {
			continue
		}
		for _, m := range linkRe.FindAllStringSubmatch(line, -1) {
			t := strings.TrimSpace(m[1])
			if t == "" {
				continue
			}
			switch {
			case strings.HasPrefix(t, "http://"),
				strings.HasPrefix(t, "https://"),
				strings.HasPrefix(t, "mailto:"),
				strings.HasPrefix(t, "#"),
				strings.HasPrefix(t, "<"):
				continue
			}
			targets = append(targets, t)
		}
	}
	return targets, nil
}

func stripFragment(s string) string {
	if i := strings.IndexAny(s, "#?"); i >= 0 {
		return s[:i]
	}
	return s
}

// CheckMarkdownStyle runs the built-in Go-native markdown style checker.
// configPath points to a style.yaml; if empty, compiled-in defaults are used.
func CheckMarkdownStyle(out io.Writer, specsRoot, configPath string, r *Result) {
	fmt.Fprintln(out, "== markdown style ==")

	cfg, err := LoadStyleConfig(configPath)
	if err != nil {
		r.errf("load style config: %v", err)
		return
	}

	count := 0
	walkErr := filepath.Walk(specsRoot, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("%s: %w", path, walkErr)
		}
		rel, _ := filepath.Rel(specsRoot, path)
		if rel == "." {
			return nil
		}
		if isExcludedRel(rel) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		violations := CheckFileStyle(path, &cfg.Rules)
		for _, v := range violations {
			relPath, _ := filepath.Rel(specsRoot, v.File)
			fmt.Fprintf(out, "%s:%d: [%s] %s\n", relPath, v.Line, v.Rule, v.Message)
			r.errf("%s:%d: [%s] %s", relPath, v.Line, v.Rule, v.Message)
			count++
		}
		return nil
	})
	if walkErr != nil {
		r.errf("walk specs root: %v", walkErr)
		return
	}
	if count == 0 {
		fmt.Fprintln(out, "ok")
	}
}
