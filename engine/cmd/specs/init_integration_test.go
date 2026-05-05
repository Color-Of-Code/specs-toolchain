// Integration tests for `specs init` covering its various framework
// resolution paths. Tests build the binary into a tempdir and run it
// with --dry-run so no network work happens; they verify the command
// emits the expected actions and exit codes.
package main_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "specs")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", bin, "./")
	cmd.Dir = mustModuleDir(t)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build: %v\n%s", err, stderr.String())
	}
	return bin
}

func mustModuleDir(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func runSpecs(t *testing.T, bin, cwd string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = cwd
	var so, se bytes.Buffer
	cmd.Stdout = &so
	cmd.Stderr = &se
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return so.String(), se.String(), ee.ExitCode()
		}
		t.Fatalf("run: %v", err)
	}
	return so.String(), se.String(), 0
}

func TestInit_URLSubmodule_DryRun(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	env := isolatedEnv(t)
	initGitRepo(t, host)

	out, se, code := runSpecsEnv(t, bin, host, env, "init",
		"--framework", "https://example.com/fw.git",
		"--dry-run")
	if code != 0 {
		t.Fatalf("exit %d\nstdout:%s\nstderr:%s", code, out, se)
	}
	for _, want := range []string{"submodule add", ".framework", "would: write"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n---\n%s", want, out)
		}
	}
}

func TestInit_LocalPath_DryRun(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	out, _, code := runSpecs(t, bin, host, "init", "--framework", "../specs-framework", "--dry-run")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	if !strings.Contains(out, "would: write") {
		t.Errorf("expected config write output:\n%s", out)
	}
}

func TestInit_DefaultFramework_DryRun(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	out, _, code := runSpecs(t, bin, host, "init", "--dry-run")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	if !strings.Contains(out, "would: write") {
		t.Errorf("expected config write line:\n%s", out)
	}
}

func TestInit_URLOutsideGitRepo_Fails(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()

	_, se, code := runSpecs(t, bin, host, "init",
		"--framework", "https://example.com/fw.git",
		"--dry-run")
	if code == 0 {
		t.Fatal("expected non-zero exit for URL framework outside a git repo")
	}
	if !strings.Contains(se, "requires a git host repository") {
		t.Errorf("expected git host repository error: %s", se)
	}
}
func TestInit_PositionalPath(t *testing.T) {
	bin := buildBinary(t)
	host := t.TempDir()
	out, _, code := runSpecs(t, bin, host, "init",
		"--framework", "./framework",
		"--dry-run",
		"specs")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	want := filepath.Join(host, "specs")
	if !strings.Contains(out, want) {
		t.Errorf("expected target path %q in output:\n%s", want, out)
	}
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, string(out))
	}
}
