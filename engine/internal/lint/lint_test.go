package lint

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Color-Of-Code/specs-toolchain/engine/internal/config"
	"github.com/Color-Of-Code/specs-toolchain/engine/internal/graph"
)

func TestLinksInFile_FiltersURLsAndAnchors(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "doc.md")
	body := strings.Join([]string{
		"# Doc",
		"",
		"See [external](https://example.com) and [anchor](#section).",
		"Also [neighbour](./neighbour.md) and [parent](../up.md#frag).",
		"",
		"```markdown",
		"[ignored](should-not-match.md)",
		"```",
		"",
		"And [mailto](mailto:a@b.c).",
	}, "\n")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := linksInFile(p)
	if err != nil {
		t.Fatalf("linksInFile: %v", err)
	}
	want := []string{"./neighbour.md", "../up.md#frag"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("link[%d]=%q want %q", i, got[i], want[i])
		}
	}
}

func TestCheckBaselinesReportsStaleCanonicalBaseline(t *testing.T) {
	workspace := t.TempDir()
	specsDir := filepath.Join(workspace, "specs")
	repoDir := filepath.Join(workspace, "repos", "host")
	for _, path := range []string{filepath.Join(specsDir, "model", "traceability"), repoDir} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	initLintGitRepo(t, repoDir, map[string]string{"README.md": "# upstream\n"})
	actual := lintGitOutput(t, repoDir, "rev-parse", "HEAD")
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
	cfg := &config.Resolved{
		GraphManifest: filepath.Join(specsDir, "model", "traceability", "graph.yaml"),
		HostRoot:      specsDir,
		Repos:         map[string]string{"host-repo": "repos/host"},
	}
	var out bytes.Buffer
	r := &Result{}
	CheckBaselines(&out, cfg, r)
	if !r.Failed() {
		t.Fatalf("expected stale baseline error, output=%s", out.String())
	}
	joined := strings.Join(r.Errors, "\n")
	if !strings.Contains(joined, actual) || !strings.Contains(joined, "baseline stale") {
		t.Fatalf("unexpected errors: %s", joined)
	}
}

func TestStripFragment(t *testing.T) {
	cases := map[string]string{
		"a/b.md":         "a/b.md",
		"a/b.md#section": "a/b.md",
		"a/b.md?x=1":     "a/b.md",
		"a/b.md#s?x":     "a/b.md",
	}
	for in, want := range cases {
		if got := stripFragment(in); got != want {
			t.Errorf("stripFragment(%q)=%q want %q", in, got, want)
		}
	}
}

func TestCheckBaselinesWarnsForUnknownRepo(t *testing.T) {
	specsDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(specsDir, "model", "traceability"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := graph.Write(filepath.Join(specsDir, "model", "traceability", "graph.yaml"), &graph.Graph{
		Baselines: []graph.BaselineEntry{{
			Component: "model/components/alpha-component",
			Repo:      "missing-repo",
			Path:      "/",
			Commit:    strings.Repeat("0", 40),
		}},
	}); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Resolved{
		GraphManifest: filepath.Join(specsDir, "model", "traceability", "graph.yaml"),
		HostRoot:      specsDir,
		Repos:         map[string]string{},
	}
	var out bytes.Buffer
	r := &Result{}
	CheckBaselines(&out, cfg, r)
	if r.Failed() {
		t.Fatalf("unexpected errors: %v", r.Errors)
	}
	if len(r.Warnings) == 0 || !strings.Contains(strings.Join(r.Warnings, "\n"), "unknown repo") {
		t.Fatalf("warnings = %v, want unknown repo warning", r.Warnings)
	}
}

func initLintGitRepo(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	runLintGit(t, dir, "init")
	runLintGit(t, dir, "config", "user.email", "specs@example.com")
	runLintGit(t, dir, "config", "user.name", "Specs Tests")
	for path, content := range files {
		fullPath := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runLintGit(t, dir, "add", ".")
	runLintGit(t, dir, "commit", "-m", "initial")
}

func lintGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %s: %v", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out))
}

func runLintGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	gitArgs := append([]string{"-C", dir, "-c", "commit.gpgsign=false"}, args...)
	cmd := exec.Command("git", gitArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, string(output))
	}
}
