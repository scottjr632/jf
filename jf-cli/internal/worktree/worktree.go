package worktree

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/scottjr632/jf-cli/internal/git"
)

var runGit = git.Run
var userHomeDir = os.UserHomeDir
var mkdirAll = os.MkdirAll

// List returns the output from `git worktree list`.
func List(ctx context.Context, repo string) (string, error) {
	return runGit(ctx, repo, "worktree", "list")
}

// Entry describes a git worktree entry from `git worktree list --porcelain`.
type Entry struct {
	Path     string
	Branch   string
	Detached bool
}

// ListEntries returns parsed worktree entries from `git worktree list --porcelain`.
func ListEntries(ctx context.Context, repo string) ([]Entry, error) {
	output, err := runGit(ctx, repo, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	return parseWorktreeEntries(output), nil
}

// Add creates a new worktree at path, optionally checking out ref.
func Add(ctx context.Context, repo, path, ref string) (string, error) {
	resolvedPath, err := resolveWorktreePath(ctx, repo, path)
	if err != nil {
		return "", err
	}
	args := []string{"worktree", "add", resolvedPath}
	if ref != "" {
		args = append(args, ref)
	}
	if _, err := runGit(ctx, repo, args...); err != nil {
		return "", err
	}
	return resolvedPath, nil
}

// Remove deletes a worktree at path.
func Remove(ctx context.Context, repo, path string) error {
	resolvedPath, err := resolveWorktreePath(ctx, repo, path)
	if err != nil {
		return err
	}
	branch, err := worktreeBranch(ctx, repo, resolvedPath)
	if err != nil {
		return err
	}
	if _, err := runGit(ctx, repo, "worktree", "remove", resolvedPath); err != nil {
		return err
	}
	if branch == "" {
		return nil
	}
	_, err = runGit(ctx, repo, "branch", "-D", branch)
	return err
}

// Prune removes stale worktree metadata.
func Prune(ctx context.Context, repo string) error {
	_, err := runGit(ctx, repo, "worktree", "prune")
	return err
}

// Merge merges the branch for the source worktree into the target branch.
func Merge(ctx context.Context, repo, sourcePath, targetBranch string) error {
	if targetBranch == "" {
		return fmt.Errorf("target branch is required")
	}
	resolvedPath, err := resolveWorktreePath(ctx, repo, sourcePath)
	if err != nil {
		return err
	}
	sourceBranch, err := worktreeBranch(ctx, repo, resolvedPath)
	if err != nil {
		return err
	}
	if sourceBranch == "" {
		return fmt.Errorf("worktree %q is detached", resolvedPath)
	}
	if sourceBranch == targetBranch {
		return fmt.Errorf("worktree %q is already on %q", resolvedPath, targetBranch)
	}
	targetPath, err := worktreePathForBranch(ctx, repo, targetBranch)
	if err != nil {
		return err
	}
	if targetPath == "" {
		return fmt.Errorf("worktree for branch %q not found", targetBranch)
	}
	_, err = runGit(ctx, targetPath, "merge", sourceBranch)
	return err
}

// PathForBranch returns the worktree path for a branch.
func PathForBranch(ctx context.Context, repo, branch string) (string, error) {
	if branch == "" {
		return "", fmt.Errorf("branch is required")
	}
	path, err := worktreePathForBranch(ctx, repo, branch)
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", fmt.Errorf("worktree for branch %q not found", branch)
	}
	return path, nil
}

// PathForBranches returns the worktree path for the first matching branch.
func PathForBranches(ctx context.Context, repo string, branches []string) (string, error) {
	for _, branch := range branches {
		if strings.TrimSpace(branch) == "" {
			continue
		}
		path, err := worktreePathForBranch(ctx, repo, branch)
		if err != nil {
			return "", err
		}
		if path != "" {
			return path, nil
		}
	}
	return "", fmt.Errorf("worktree for branches %v not found", branches)
}

func resolveWorktreePath(ctx context.Context, repo, path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("worktree path is required")
	}
	if filepath.IsAbs(path) || strings.HasPrefix(path, ".") {
		return path, nil
	}
	baseDir, err := defaultWorktreeBase(ctx, repo)
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, path), nil
}

// ResolvePath converts a named or relative path into a full worktree path.
func ResolvePath(ctx context.Context, repo, path string) (string, error) {
	return resolveWorktreePath(ctx, repo, path)
}

func defaultWorktreeBase(ctx context.Context, repo string) (string, error) {
	top, err := runGit(ctx, repo, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	top = strings.TrimSpace(top)
	if top == "" {
		return "", fmt.Errorf("git rev-parse --show-toplevel returned empty path")
	}
	repoName := filepath.Base(top)
	home, err := userHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	baseDir := filepath.Join(home, ".jf", repoName, "worktrees")
	if err := mkdirAll(baseDir, 0o755); err != nil {
		return "", fmt.Errorf("create worktree dir: %w", err)
	}
	return baseDir, nil
}

func worktreeBranch(ctx context.Context, repo, resolvedPath string) (string, error) {
	output, err := runGit(ctx, repo, "worktree", "list", "--porcelain")
	if err != nil {
		return "", err
	}
	paths := []string{resolvedPath}
	if !filepath.IsAbs(resolvedPath) {
		if absPath, err := filepath.Abs(resolvedPath); err == nil {
			paths = append(paths, absPath)
		}
	}
	lines := strings.Split(output, "\n")
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(line, "worktree ") {
			continue
		}
		wtPath := strings.TrimSpace(strings.TrimPrefix(line, "worktree "))
		if !stringSliceContains(paths, wtPath) {
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			entry := strings.TrimSpace(lines[j])
			if strings.HasPrefix(entry, "worktree ") {
				break
			}
			if strings.HasPrefix(entry, "branch ") {
				ref := strings.TrimSpace(strings.TrimPrefix(entry, "branch "))
				return strings.TrimPrefix(ref, "refs/heads/"), nil
			}
		}
		return "", nil
	}
	return "", nil
}

func worktreePathForBranch(ctx context.Context, repo, branch string) (string, error) {
	output, err := runGit(ctx, repo, "worktree", "list", "--porcelain")
	if err != nil {
		return "", err
	}
	lines := strings.Split(output, "\n")
	worktreePath := ""
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "worktree ") {
			worktreePath = strings.TrimSpace(strings.TrimPrefix(line, "worktree "))
			continue
		}
		if strings.HasPrefix(line, "branch ") {
			ref := strings.TrimSpace(strings.TrimPrefix(line, "branch "))
			if strings.TrimPrefix(ref, "refs/heads/") == branch {
				return worktreePath, nil
			}
		}
	}
	return "", nil
}

func stringSliceContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func parseWorktreeEntries(output string) []Entry {
	lines := strings.Split(output, "\n")
	var entries []Entry
	var current *Entry

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "worktree ") {
			if current != nil {
				entries = append(entries, *current)
			}
			path := strings.TrimSpace(strings.TrimPrefix(line, "worktree "))
			current = &Entry{Path: path}
			continue
		}
		if current == nil {
			continue
		}
		if strings.HasPrefix(line, "branch ") {
			ref := strings.TrimSpace(strings.TrimPrefix(line, "branch "))
			current.Branch = strings.TrimPrefix(ref, "refs/heads/")
			continue
		}
		if line == "detached" {
			current.Detached = true
		}
	}

	if current != nil {
		entries = append(entries, *current)
	}

	return entries
}
