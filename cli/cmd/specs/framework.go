package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Color-Of-Code/specs-toolchain/cli/internal/framework"
)

func cmdFramework(args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: specs framework <seed>")
		return exitWith(2, "missing subcommand")
	}
	switch args[0] {
	case "seed":
		return cmdFrameworkSeed(args[1:])
	default:
		return exitWith(2, "unknown subcommand: specs framework %s", args[0])
	}
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
