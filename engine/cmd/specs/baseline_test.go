package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Color-Of-Code/specs-toolchain/engine/internal/config"
	"github.com/Color-Of-Code/specs-toolchain/engine/internal/graph"
)

func TestCmdBaselineUpdateWritesCanonicalGraph(t *testing.T) {
	workspace := t.TempDir()
	specsDir := filepath.Join(workspace, "specs")
	repoDir := filepath.Join(workspace, "repos", "host")
	for _, path := range []string{
		filepath.Join(specsDir, "model", "traceability"),
		filepath.Join(specsDir, "model", "components"),
		repoDir,
	} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	initGitRepo(t, repoDir, map[string]string{"README.md": "# upstream\n"})
	actualCommit := gitOutputTest(t, repoDir, "rev-parse", "HEAD")

	if err := config.Save(filepath.Join(specsDir, config.FileName), &config.File{
		Repos: map[string]string{"host-repo": "repos/host"},
	}); err != nil {
		t.Fatal(err)
	}
	componentPath := filepath.Join(specsDir, "model", "components", "alpha-component.md")
	componentBody := strings.Join([]string{
		"# Alpha Component",
		"",
		"| Field | Value |",
		"| ----- | ----- |",
		"| Status | Draft |",
		"| Requirements | — |",
		"| Baseline | — |",
	}, "\n")
	if err := os.WriteFile(componentPath, []byte(componentBody), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := graph.Write(filepath.Join(specsDir, "model", "traceability", "graph.yaml"), &graph.Graph{
		Baselines: []graph.BaselineEntry{{
			Component: "model/components/alpha-component",
			Repo:      "host-repo",
			Path:      "/",
			Commit:    strings.Repeat("0", 40),
		}},
	}); err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()
	if err := os.Chdir(specsDir); err != nil {
		t.Fatal(err)
	}

	if err := cmdBaselineUpdate(nil); err != nil {
		t.Fatalf("cmdBaselineUpdate() error = %v", err)
	}

	reloaded, err := graph.Load(filepath.Join(specsDir, "model", "traceability", "graph.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(reloaded.Baselines) != 1 || reloaded.Baselines[0].Commit != actualCommit {
		t.Fatalf("updated baselines = %+v, want commit %s", reloaded.Baselines, actualCommit)
	}
	updatedComponent, err := os.ReadFile(componentPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(updatedComponent), "`host-repo:/@"+actualCommit+"`") {
		t.Fatalf("component baseline field not regenerated:\n%s", string(updatedComponent))
	}
}

func TestCmdBaselineUpdateDryRunDoesNotWriteGraph(t *testing.T) {
	workspace := t.TempDir()
	specsDir := filepath.Join(workspace, "specs")
	repoDir := filepath.Join(workspace, "repos", "host")
	for _, path := range []string{
		filepath.Join(specsDir, "model", "traceability"),
		filepath.Join(specsDir, "model", "components"),
		repoDir,
	} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	initGitRepo(t, repoDir, map[string]string{"README.md": "# upstream\n"})

	if err := config.Save(filepath.Join(specsDir, config.FileName), &config.File{
		Repos: map[string]string{"host-repo": "repos/host"},
	}); err != nil {
		t.Fatal(err)
	}
	componentPath := filepath.Join(specsDir, "model", "components", "alpha-component.md")
	componentBody := strings.Join([]string{
		"# Alpha Component",
		"",
		"| Field | Value |",
		"| ----- | ----- |",
		"| Status | Draft |",
		"| Requirements | — |",
		"| Baseline | — |",
	}, "\n")
	if err := os.WriteFile(componentPath, []byte(componentBody), 0o644); err != nil {
		t.Fatal(err)
	}
	oldCommit := strings.Repeat("1", 40)
	if err := graph.Write(filepath.Join(specsDir, "model", "traceability", "graph.yaml"), &graph.Graph{
		Baselines: []graph.BaselineEntry{{
			Component: "model/components/alpha-component",
			Repo:      "host-repo",
			Path:      "/",
			Commit:    oldCommit,
		}},
	}); err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()
	if err := os.Chdir(specsDir); err != nil {
		t.Fatal(err)
	}

	if err := cmdBaselineUpdate([]string{"--dry-run"}); err != nil {
		t.Fatalf("cmdBaselineUpdate() dry-run error = %v", err)
	}

	reloaded, err := graph.Load(filepath.Join(specsDir, "model", "traceability", "graph.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(reloaded.Baselines) != 1 || reloaded.Baselines[0].Commit != oldCommit {
		t.Fatalf("dry-run baselines = %+v, want unchanged commit %s", reloaded.Baselines, oldCommit)
	}
	updatedComponent, err := os.ReadFile(componentPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(updatedComponent), "host-repo:/@") {
		t.Fatalf("dry-run should not rewrite component markdown:\n%s", string(updatedComponent))
	}
}

func initGitRepo(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	runGitTest(t, dir, "init")
	runGitTest(t, dir, "config", "user.email", "specs@example.com")
	runGitTest(t, dir, "config", "user.name", "Specs Tests")
	for path, content := range files {
		fullPath := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runGitTest(t, dir, "add", ".")
	runGitTest(t, dir, "commit", "-m", "initial")
}

func gitOutputTest(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %s: %v", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out))
}

func runGitTest(t *testing.T, dir string, args ...string) {
	t.Helper()
	gitArgs := append([]string{"-C", dir, "-c", "commit.gpgsign=false"}, args...)
	cmd := exec.Command("git", gitArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, string(output))
	}
}
