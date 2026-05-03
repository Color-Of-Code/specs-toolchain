package baseline

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ResolveCommit(component, repo, repoPath string, repos map[string]string, workspace string) (string, error) {
	repoRoot, err := resolveRepoRoot(component, repo, repos, workspace)
	if err != nil {
		return "", err
	}
	gitArgs := []string{"-C", repoRoot, "log", "-1", "--format=%H"}
	if repoPath != "/" {
		gitArgs = append(gitArgs, "--", repoPath)
	}
	out, err := exec.Command("git", gitArgs...).Output()
	if err != nil {
		return "", fmt.Errorf("git log failed for %s:%s: %v", repo, repoPath, err)
	}
	sha := strings.TrimSpace(string(out))
	if sha == "" {
		return "", fmt.Errorf("no git history for %s:%s", repo, repoPath)
	}
	return sha, nil
}

func IsGitRepo(dir string) bool {
	if st, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		_ = st
		return true
	}
	return false
}

func resolveRepoRoot(component, repo string, repos map[string]string, workspace string) (string, error) {
	repoPath, ok := repos[repo]
	if !ok {
		return "", fmt.Errorf("unknown repo %q (add to repos: in .specs.yaml); component=%q", repo, component)
	}
	absRepo := filepath.Join(workspace, repoPath)
	if !IsGitRepo(absRepo) {
		return "", fmt.Errorf("repo not checked out at %s", absRepo)
	}
	return absRepo, nil
}