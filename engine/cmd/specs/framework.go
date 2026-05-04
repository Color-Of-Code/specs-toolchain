package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/Color-Of-Code/specs-toolchain/engine/internal/cache"
	"github.com/Color-Of-Code/specs-toolchain/engine/internal/config"
	"github.com/Color-Of-Code/specs-toolchain/engine/internal/framework"
	"github.com/Color-Of-Code/specs-toolchain/engine/internal/registry"
)

func cmdFramework(args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: specs framework <list|add|remove|seed|update>")
		return exitWith(2, "missing subcommand")
	}
	switch args[0] {
	case "list":
		return cmdFrameworkList(args[1:])
	case "add":
		return cmdFrameworkAdd(args[1:])
	case "remove", "rm":
		return cmdFrameworkRemove(args[1:])
	case "seed":
		return cmdFrameworkSeed(args[1:])
	case "update":
		return cmdFrameworkUpdate(args[1:])
	case "skills":
		return cmdFrameworkSkills(args[1:])
	case "-h", "--help", "help":
		fmt.Fprintln(os.Stderr, "Usage: specs framework <list|add|remove|seed|update|skills> [flags]")
		return nil
	default:
		return exitWith(2, "unknown subcommand: specs framework %s", args[0])
	}
}

// cmdFrameworkUpdate updates the framework content layer in place.
//
//	managed: fetch into the user cache and rewrite framework_ref
//	local:   git fetch + checkout/pull inside framework_dir (no-op for non-git checkouts)
func cmdFrameworkUpdate(args []string) error {
	fs2 := flag.NewFlagSet("framework update", flag.ContinueOnError)
	to := fs2.String("to", "", "tag/branch/commit to check out (empty = pull current branch / default branch)")
	fs2.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs framework update [--to <ref>]")
		fs2.PrintDefaults()
	}
	if err := fs2.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load("")
	if err != nil {
		return err
	}

	switch cfg.FrameworkMode {
	case config.FrameworkModeManaged:
		return updateManagedFramework(cfg, *to)
	case config.FrameworkModeLocal:
		if cfg.FrameworkDir == "" {
			return exitWith(1, "framework_dir not found; run `specs init` or set framework_dir")
		}
		if _, err := os.Stat(filepath.Join(cfg.FrameworkDir, ".git")); err != nil {
			return exitWith(2, "framework_dir %s has no .git; refresh it manually (re-clone, re-vendor, or `git submodule update`)", cfg.FrameworkDir)
		}
		if err := runGit(cfg.FrameworkDir, "fetch", "--tags"); err != nil {
			return err
		}
		if *to != "" {
			return runGit(cfg.FrameworkDir, "checkout", *to)
		}
		// pull on current branch; if detached, this fails harmlessly.
		_ = runGit(cfg.FrameworkDir, "pull", "--ff-only")
		return nil
	case config.FrameworkModeMissing:
		return exitWith(1, "framework_dir is missing on disk; run `specs init`")
	default:
		return exitWith(1, "unknown framework_mode %q", cfg.FrameworkMode)
	}
}

// runGit invokes git with the given args inside dir.
func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// fetchManaged ensures the user-cache copy of a managed framework exists at
// the requested ref. An empty ref defaults to "main". When dryRun is true,
// the action is printed and the call is skipped; the returned path is empty.
// The resolved ref is always returned so callers can persist it in
// .specs.yaml.
func fetchManaged(url, ref string, dryRun bool) (path, resolvedRef string, err error) {
	resolvedRef = ref
	if resolvedRef == "" {
		resolvedRef = "main"
	}
	if dryRun {
		fmt.Printf("would: fetch %s@%s into managed cache\n", url, resolvedRef)
		return "", resolvedRef, nil
	}
	path, err = cache.Ensure(url, resolvedRef)
	if err != nil {
		return "", resolvedRef, exitWith(1, "fetch %s@%s: %v", url, resolvedRef, err)
	}
	return path, resolvedRef, nil
}

// updateManagedFramework fetches the requested ref into the user cache and rewrites
// framework_ref in .specs.yaml so subsequent invocations resolve to it.
func updateManagedFramework(cfg *config.Resolved, to string) error {
	ref := to
	if ref == "" {
		ref = cfg.FrameworkRef
	}
	path, resolvedRef, err := fetchManaged(cfg.FrameworkURL, ref, false)
	if err != nil {
		return err
	}
	fmt.Printf("managed framework cached at %s\n", path)

	// Always persist the resolved ref so the config stays authoritative even
	// when --to was not given (e.g. defaults to "main").
	if cfg.ConfigPath != "" && cfg.Source != nil && resolvedRef != cfg.FrameworkRef {
		newFile := *cfg.Source
		newFile.FrameworkRef = resolvedRef
		if err := config.Save(cfg.ConfigPath, &newFile); err != nil {
			return exitWith(1, "write %s: %v", cfg.ConfigPath, err)
		}
		fmt.Printf("updated %s: framework_ref=%s\n", cfg.ConfigPath, resolvedRef)
	}
	return nil
}

// cmdFrameworkList prints all registered framework entries.
func cmdFrameworkList(args []string) error {
	fs2 := flag.NewFlagSet("framework list", flag.ContinueOnError)
	fs2.Usage = func() { fmt.Fprintln(os.Stderr, "Usage: specs framework list") }
	if err := fs2.Parse(args); err != nil {
		return err
	}
	reg, err := registry.Load("")
	if err != nil {
		return err
	}
	names := reg.Names()
	if len(names) == 0 {
		path, _ := registry.DefaultPath()
		fmt.Printf("No frameworks registered (%s does not exist or is empty).\n", path)
		return nil
	}
	for _, n := range names {
		e := reg.Frameworks[n]
		switch {
		case e.URL != "":
			ref := e.Ref
			if ref == "" {
				ref = "main"
			}
			fmt.Printf("%s\turl=%s\tref=%s\n", n, e.URL, ref)
		case e.Path != "":
			fmt.Printf("%s\tpath=%s\n", n, e.Path)
		}
	}
	return nil
}

// cmdFrameworkAdd inserts or replaces an entry in the registry.
func cmdFrameworkAdd(args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: specs framework add <name> (--url <URL> [--ref <ref>] | --path <dir>)")
		return exitWith(2, "missing name")
	}
	name := args[0]
	fs2 := flag.NewFlagSet("framework add", flag.ContinueOnError)
	url := fs2.String("url", "", "git URL of a remote framework source")
	ref := fs2.String("ref", "", "tag/branch/commit (only with --url; default 'main')")
	path := fs2.String("path", "", "local directory path of a framework source")
	fs2.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs framework add <name> (--url <URL> [--ref <ref>] | --path <dir>)")
		fs2.PrintDefaults()
	}
	if err := fs2.Parse(args[1:]); err != nil {
		return err
	}
	entry := registry.Entry{URL: *url, Ref: *ref, Path: *path}
	if entry.Path != "" {
		// Expand a leading ~ for convenience and store an absolute path.
		expanded, err := expandHome(entry.Path)
		if err != nil {
			return err
		}
		abs, err := filepath.Abs(expanded)
		if err != nil {
			return err
		}
		entry.Path = abs
	}
	if err := entry.Validate(); err != nil {
		return exitWith(2, "%v", err)
	}
	reg, err := registry.Load("")
	if err != nil {
		return err
	}
	if err := reg.Add(name, entry); err != nil {
		return err
	}
	if err := reg.Save(""); err != nil {
		return err
	}
	regPath, _ := registry.DefaultPath()
	fmt.Printf("registered %q in %s\n", name, regPath)
	return nil
}

// cmdFrameworkRemove deletes an entry from the registry.
func cmdFrameworkRemove(args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: specs framework remove <name>")
		return exitWith(2, "missing name")
	}
	name := args[0]
	reg, err := registry.Load("")
	if err != nil {
		return err
	}
	if err := reg.Remove(name); err != nil {
		return exitWith(1, "%v", err)
	}
	if err := reg.Save(""); err != nil {
		return err
	}
	fmt.Printf("removed %q from registry\n", name)
	return nil
}

// expandHome expands a leading "~" or "~/" segment to the user's home dir.
func expandHome(p string) (string, error) {
	if p == "~" || (len(p) >= 2 && p[:2] == "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if p == "~" {
			return home, nil
		}
		return filepath.Join(home, p[2:]), nil
	}
	return p, nil
}

func cmdFrameworkSeed(args []string) error {
	fs2 := flag.NewFlagSet("framework seed", flag.ContinueOnError)
	out := fs2.String("out", "", "directory to create and populate (required)")
	fs2.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs framework seed --out <dir>")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Pre-seeds an empty directory with the minimal framework skeleton.")
		fmt.Fprintln(os.Stderr, "The caller is responsible for git init and pushing to a remote.")
		fs2.PrintDefaults()
	}
	if err := fs2.Parse(args); err != nil {
		return err
	}
	if *out == "" {
		fs2.Usage()
		return exitWith(2, "--out is required")
	}

	dest, err := filepath.Abs(*out)
	if err != nil {
		return err
	}

	// Fail if target exists and is non-empty.
	if entries, err := os.ReadDir(dest); err == nil && len(entries) > 0 {
		return exitWith(1, "target directory %s is not empty", dest)
	}

	// Walk the embedded template FS and write files.
	err = fs.WalkDir(framework.Template, "template", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		// Strip the "template/" prefix to get the relative path.
		rel, err := filepath.Rel("template", path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(dest, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		data, err := fs.ReadFile(framework.Template, path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		return fmt.Errorf("seed framework: %w", err)
	}

	fmt.Printf("Framework skeleton created at %s\n", dest)
	fmt.Println("Next steps:")
	fmt.Println("  cd", dest)
	fmt.Println("  git init && git add -A && git commit -m \"initial framework skeleton\"")
	return nil
}

// SkillFrontmatter holds the YAML front-matter from a skill file under
// framework/skills/*.md. Only the fields the extension needs are parsed.
type SkillFrontmatter struct {
	ID          string                 `yaml:"id"`
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Tags        []string               `yaml:"tags"`
	InputSchema map[string]interface{} `yaml:"inputSchema"`
	EngineArgs  map[string][]string    `yaml:"engineArgs"`
}

// SkillInfo is the JSON-serialisable record emitted by `specs framework skills list`.
type SkillInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Tags        []string               `json:"tags"`
	InputSchema map[string]interface{} `json:"inputSchema,omitempty"`
	EngineArgs  map[string][]string    `json:"engineArgs,omitempty"`
	File        string                 `json:"file"`
}

// cmdFrameworkSkills handles `specs framework skills list`.
func cmdFrameworkSkills(args []string) error {
	fs2 := flag.NewFlagSet("framework skills", flag.ContinueOnError)
	fs2.Usage = func() { fmt.Fprintln(os.Stderr, "Usage: specs framework skills list") }
	if err := fs2.Parse(args); err != nil {
		return err
	}
	if len(fs2.Args()) == 0 || fs2.Args()[0] == "list" {
		return cmdFrameworkSkillsList()
	}
	fs2.Usage()
	return exitWith(2, "unknown subcommand: specs framework skills %s", fs2.Args()[0])
}

func cmdFrameworkSkillsList() error {
	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	if cfg.FrameworkDir == "" {
		fmt.Println("[]")
		return nil
	}
	skillsDir := filepath.Join(cfg.FrameworkDir, "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("[]")
			return nil
		}
		return fmt.Errorf("read skills dir: %w", err)
	}

	var skills []SkillInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		filePath := filepath.Join(skillsDir, e.Name())
		info, err := parseSkillFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skip %s: %v\n", e.Name(), err)
			continue
		}
		info.File = filePath
		skills = append(skills, *info)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if skills == nil {
		_, err = os.Stdout.WriteString("[]\n")
		return err
	}
	return enc.Encode(skills)
}

// parseSkillFile reads a Markdown file with YAML front-matter delimited by ---
// lines and returns the parsed SkillInfo.
func parseSkillFile(path string) (*SkillInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	frontmatter, err := extractFrontmatter(string(data))
	if err != nil {
		return nil, fmt.Errorf("frontmatter: %w", err)
	}
	var fm SkillFrontmatter
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}
	if fm.ID == "" {
		return nil, fmt.Errorf("missing required field: id")
	}
	return &SkillInfo{
		ID:          fm.ID,
		Name:        fm.Name,
		Description: fm.Description,
		Tags:        fm.Tags,
		InputSchema: fm.InputSchema,
		EngineArgs:  fm.EngineArgs,
	}, nil
}

// extractFrontmatter returns the content between the first pair of "---" lines.
func extractFrontmatter(content string) (string, error) {
	lines := strings.SplitN(content, "\n", -1)
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "---" {
		return "", fmt.Errorf("no front-matter found")
	}
	var buf strings.Builder
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "---" {
			return buf.String(), nil
		}
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
	return "", fmt.Errorf("front-matter not closed with ---")
}
