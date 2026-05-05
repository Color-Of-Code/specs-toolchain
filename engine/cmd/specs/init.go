package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Color-Of-Code/specs-toolchain/engine/internal/config"
)

// cmdInit configures a host repository for use with the specs toolchain.
//
// It is git-init-like: idempotent, creates the target directory if missing,
// resolves the framework source, and writes .specs.yaml.
//
// Usage:
//
//	specs init [<path>] [--framework <source>]
//	           [--with-model] [--with-vscode] [--force] [--dry-run]
//
// `<path>` defaults to the current directory.
//
// `--framework <source>` accepts:
//   - omitted              -> use local "./framework"
//   - filesystem path      -> persisted as framework_dir
//   - remote git URL       -> cloned as submodule at specs/.framework
func cmdInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	frameworkSpec := fs.String("framework", "", "framework source: local path or remote git URL. URLs are cloned as specs/.framework submodules; local paths are written to framework_dir.")
	withModel := fs.Bool("with-model", false, "create empty model/, product/, and change-requests/ skeletons")
	withVSCode := fs.Bool("with-vscode", false, "write .vscode/tasks.json")
	force := fs.Bool("force", false, "overwrite an existing .specs.yaml")
	dryRun := fs.Bool("dry-run", false, "print actions without performing them")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs init [<path>] [--framework <source>] [--with-model] [--with-vscode] [--force] [--dry-run]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Positional <path> (default: cwd).
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	specsRoot := cwd
	switch fs.NArg() {
	case 0:
		// already cwd
	case 1:
		p := fs.Arg(0)
		if filepath.IsAbs(p) {
			specsRoot = p
		} else {
			specsRoot = filepath.Join(cwd, p)
		}
	default:
		return exitWith(2, "too many positional arguments; expected at most one <path>")
	}

	// Refuse to overwrite an existing .specs.yaml without --force.
	cfgPath := filepath.Join(specsRoot, config.FileName)
	if _, err := os.Stat(cfgPath); err == nil && !*force {
		return exitWith(1, "%s already exists (use --force to overwrite)", cfgPath)
	}

	if err := ensureDir(specsRoot, *dryRun); err != nil {
		return err
	}

	frameworkInput := strings.TrimSpace(*frameworkSpec)
	if frameworkInput == "" {
		frameworkInput = "./framework"
	}

	f := &config.File{MinSpecsVersion: Version}
	if looksLikeGitURL(frameworkInput) {
		if err := setupFrameworkSubmodule(specsRoot, frameworkInput, *dryRun); err != nil {
			return err
		}
	} else {
		f.FrameworkDir = frameworkInput
	}

	if err := saveConfig(cfgPath, f, *dryRun); err != nil {
		return err
	}

	return finalizeInit(specsRoot, *withModel, *withVSCode, *dryRun)
}

func setupFrameworkSubmodule(specsRoot, frameworkURL string, dryRun bool) error {
	hostRoot := gitRepoRoot(specsRoot)
	if hostRoot == "" {
		return exitWith(2, "--framework URL requires a git host repository")
	}

	submoduleDir := filepath.Join(specsRoot, ".framework")
	if st, err := os.Stat(submoduleDir); err == nil && st.IsDir() {
		return nil
	}

	relPath, err := filepath.Rel(hostRoot, submoduleDir)
	if err != nil {
		return err
	}
	relPath = filepath.ToSlash(relPath)

	return runOrLog(dryRun, fmt.Sprintf("git -C %s submodule add %s %s", hostRoot, frameworkURL, relPath), func() error {
		cmd := exec.Command("git", "-C", hostRoot, "submodule", "add", frameworkURL, relPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return exitWith(1, "git submodule add failed: %v", err)
		}
		return nil
	})
}

func looksLikeGitURL(value string) bool {
	v := strings.ToLower(value)
	return strings.HasPrefix(v, "http://") ||
		strings.HasPrefix(v, "https://") ||
		strings.HasPrefix(v, "ssh://") ||
		strings.HasPrefix(value, "git@")
}

func gitRepoRoot(start string) string {
	cmd := exec.Command("git", "-C", start, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// ensureDir creates the specs root if it does not exist.
func ensureDir(dir string, dryRun bool) error {
	if _, err := os.Stat(dir); err == nil {
		return nil
	}
	return runOrLog(dryRun, "mkdir -p "+dir, func() error { return os.MkdirAll(dir, 0o755) })
}

func saveConfig(cfgPath string, f *config.File, dryRun bool) error {
	if f.Repos == nil {
		f.Repos = map[string]string{}
	}
	if dryRun {
		fmt.Printf("would: write %s\n", cfgPath)
		return nil
	}
	if err := config.Save(cfgPath, f); err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", cfgPath)
	return nil
}

// finalizeInit writes optional skeletons after the config is in place.
func finalizeInit(specsRoot string, withModel, withVSCode, dryRun bool) error {
	if withModel {
		for _, sub := range []string{"model", "product", "change-requests"} {
			p := filepath.Join(specsRoot, sub)
			if err := runOrLog(dryRun, "mkdir -p "+p, func() error { return os.MkdirAll(p, 0o755) }); err != nil {
				return err
			}
		}
	}
	if withVSCode {
		if dryRun {
			fmt.Printf("would: write %s/.vscode/tasks.json\n", specsRoot)
			return nil
		}
		if err := writeVSCodeTasks(specsRoot); err != nil {
			return err
		}
		fmt.Println("wrote .vscode/tasks.json")
	}
	return nil
}

// runOrLog executes fn unless dryRun is set, in which case it just prints
// the label.
func runOrLog(dryRun bool, label string, fn func() error) error {
	if dryRun {
		fmt.Println("would:", label)
		return nil
	}
	fmt.Println("$", label)
	return fn()
}
