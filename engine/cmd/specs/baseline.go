package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	baselineutil "github.com/Color-Of-Code/specs-toolchain/engine/internal/baseline"
	"github.com/Color-Of-Code/specs-toolchain/engine/internal/config"
	"github.com/Color-Of-Code/specs-toolchain/engine/internal/graph"
)

// cmdBaseline dispatches `specs baseline <subcommand>`.
func cmdBaseline(args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: specs baseline <subcommand>")
		fmt.Fprintln(os.Stderr, "Subcommands: update")
		return exitWith(2, "missing subcommand")
	}
	switch args[0] {
	case "update":
		return cmdBaselineUpdate(args[1:])
	case "check":
		return exitWith(2, "`specs baseline check` was removed; use `specs lint --baselines` instead")
	case "-h", "--help", "help":
		fmt.Fprintln(os.Stderr, "Usage: specs baseline update [flags]")
		return nil
	default:
		return exitWith(2, "unknown baseline subcommand %q", args[0])
	}
}

// cmdBaselineUpdate rewrites canonical baseline entries in the graph manifest
// and then regenerates the projected component Baseline field values.
//
// Flags:
//
//	--only <component-substring>   restrict the update to baseline entries
//	                               whose component id contains the substring
//	--dry-run                      print proposed changes only
func cmdBaselineUpdate(args []string) error {
	fs := flag.NewFlagSet("baseline update", flag.ContinueOnError)
	only := fs.String("only", "", "only update baseline entries whose component id contains this substring")
	dryRun := fs.Bool("dry-run", false, "print the proposed baseline updates without writing")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs baseline update [--only <substr>] [--dry-run]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	g, err := graph.Load(cfg.GraphManifest)
	if err != nil {
		return exitWith(1, "load graph %s: %v", cfg.GraphManifest, err)
	}
	if err := validateBaselineRepos(g, cfg.Repos); err != nil {
		return exitWith(1, "validate %s: %v", cfg.GraphManifest, err)
	}

	workspace := filepath.Dir(cfg.HostRoot)

	updated := 0
	skipped := 0
	for index, entry := range g.Baselines {
		if !matchesBaselineFilter(entry.Component, *only) {
			skipped++
			continue
		}
		actualCommit, err := baselineutil.ResolveCommit(entry.Component, entry.Repo, entry.Path, cfg.Repos, workspace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
			skipped++
			continue
		}
		if actualCommit == entry.Commit {
			continue
		}
		if *dryRun {
			fmt.Println("- " + formatBaselineChange(entry.Component, entry.Repo, entry.Path, entry.Commit))
			fmt.Println("+ " + formatBaselineChange(entry.Component, entry.Repo, entry.Path, actualCommit))
		}
		g.Baselines[index].Commit = actualCommit
		updated++
	}

	if *dryRun {
		fmt.Printf("would update %d baseline(s); skipped %d\n", updated, skipped)
		return nil
	}
	if updated == 0 {
		fmt.Println("no changes")
		return nil
	}
	if err := graph.Write(cfg.GraphManifest, g); err != nil {
		return exitWith(1, "write graph: %v", err)
	}
	result, err := graph.GenerateMarkdown(cfg.ModelDir, cfg.ProductDir, g, false)
	if err != nil {
		return exitWith(1, "generate markdown: %v", err)
	}
	fmt.Printf("updated %d baseline(s) in %s\n", updated, cfg.GraphManifest)
	fmt.Printf("updated markdown files: %d\n", len(result.UpdatedFiles))
	return nil
}

func matchesBaselineFilter(component, only string) bool {
	return only == "" || strings.Contains(component, only)
}

func formatBaselineChange(component, repo, repoPath, commit string) string {
	return fmt.Sprintf("%s | %s | %s | %s", component, repo, repoPath, commit)
}

// cmdCRDrain interactively migrates files from a CR's local working tree
// into their canonical homes under model/. Each file is shown with its
// proposed destination; the user accepts (y), edits the destination (e),
// or skips (n). Accepted moves use `git mv` so the rename is tracked.
//
// Usage:
//
//	specs cr drain --id <NNN> [--yes] [--dry-run]
//
// --yes accepts every default destination without prompting.
func cmdCRDrain(args []string) error {
	fs := flag.NewFlagSet("cr drain", flag.ContinueOnError)
	id := fs.String("id", "", "CR id (required)")
	yes := fs.Bool("yes", false, "accept every default destination")
	dryRun := fs.Bool("dry-run", false, "print proposed moves only")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs cr drain --id <NNN> [--yes] [--dry-run]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id == "" {
		return exitWith(2, "--id is required")
	}

	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	crDir, err := findCRDir(cfg.ChangeRequestsDir, *id)
	if err != nil {
		return err
	}

	type move struct{ src, dst string }
	var moves []move

	// Per-area drain destinations. Most CR-local subtrees mirror the model/
	// layout 1:1; product-requirements/ flattens into the canonical product/
	// tree (the kind is the tree, no nested folder).
	type drainArea struct {
		name string
		dst  string
	}
	areas := []drainArea{
		{"product-requirements", cfg.ProductDir},
		{"requirements", filepath.Join(cfg.ModelDir, "requirements")},
		{"use-cases", filepath.Join(cfg.ModelDir, "use-cases")},
		{"components", filepath.Join(cfg.ModelDir, "components")},
		{"architecture", filepath.Join(cfg.ModelDir, "architecture")},
	}
	for _, a := range areas {
		srcRoot := filepath.Join(crDir, a.name)
		dstRoot := a.dst
		_ = filepath.Walk(srcRoot, func(p string, info os.FileInfo, walkErr error) error {
			if walkErr != nil || info == nil || info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(p, ".md") {
				return nil
			}
			rel, err := filepath.Rel(srcRoot, p)
			if err != nil {
				return nil
			}
			moves = append(moves, move{src: p, dst: filepath.Join(dstRoot, rel)})
			return nil
		})
	}

	if len(moves) == 0 {
		fmt.Println("no files to drain")
		return nil
	}

	reader := bufio.NewReader(os.Stdin)
	accepted := 0
	for _, m := range moves {
		fmt.Println("")
		relSrc, _ := filepath.Rel(cfg.SpecsRoot, m.src)
		relDst, _ := filepath.Rel(cfg.SpecsRoot, m.dst)
		fmt.Printf("  %s\n  -> %s\n", relSrc, relDst)
		dst := m.dst
		if !*yes {
			fmt.Print("  [y]es / [n]o / [e]dit dest: ")
			ans, _ := reader.ReadString('\n')
			ans = strings.TrimSpace(ans)
			switch ans {
			case "n", "no":
				continue
			case "e", "edit":
				fmt.Print("  new destination (relative to specs root): ")
				newRel, _ := reader.ReadString('\n')
				newRel = strings.TrimSpace(newRel)
				if newRel == "" {
					continue
				}
				dst = filepath.Join(cfg.SpecsRoot, newRel)
			}
		}
		if _, err := os.Stat(dst); err == nil {
			fmt.Printf("  destination exists; skipping\n")
			continue
		}
		if *dryRun {
			fmt.Printf("  would: git mv %s %s\n", m.src, dst)
			accepted++
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "  error: mkdir %s: %v\n", filepath.Dir(dst), err)
			continue
		}
		gitDir := cfg.HostRoot
		if gitDir == "" {
			gitDir = cfg.SpecsRoot
		}
		cmd := exec.Command("git", "-C", gitDir, "mv", m.src, dst)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			// Fall back to a plain rename if git mv refuses (untracked file).
			if rerr := os.Rename(m.src, dst); rerr != nil {
				fmt.Fprintf(os.Stderr, "  error: %v / %v\n", err, rerr)
				continue
			}
			fmt.Println("  moved (untracked, plain rename)")
		} else {
			fmt.Println("  moved")
		}
		accepted++
	}
	fmt.Printf("\ndrained %d file(s); review with `git status` and commit.\n", accepted)
	return nil
}
