package stack

import (
	"context"
	"fmt"
	"strings"
)

// Commit describes a stack item derived from git history.
type Commit struct {
	SHA     string
	Short   string
	Subject string
	Body    string
}

// Stack represents the current stack of commits from trunk to HEAD.
type Stack struct {
	Trunk   string
	Head    string
	Commits []Commit
}

// CurrentStack returns commits between trunk and HEAD in order.
func CurrentStack(ctx context.Context, repo, trunk string) (Stack, error) {
	resolvedTrunk := strings.TrimSpace(trunk)
	if resolvedTrunk == "" {
		resolvedTrunk = defaultTrunk
	}
	if err := ensureBranchExists(ctx, repo, resolvedTrunk); err != nil {
		return Stack{}, fmt.Errorf("trunk branch %q not found: %w", resolvedTrunk, err)
	}

	head, err := currentBranch(ctx, repo)
	if err != nil {
		return Stack{}, err
	}

	if err := ensureAncestor(ctx, repo, resolvedTrunk, "HEAD"); err != nil {
		return Stack{}, err
	}

	commits, err := listCommits(ctx, repo, resolvedTrunk)
	if err != nil {
		return Stack{}, err
	}

	return Stack{Trunk: resolvedTrunk, Head: head, Commits: commits}, nil
}

func ensureBranchExists(ctx context.Context, repo, branch string) error {
	if strings.TrimSpace(branch) == "" {
		return fmt.Errorf("branch name is required")
	}
	_, err := runGit(ctx, repo, "rev-parse", "--verify", "refs/heads/"+branch)
	return err
}

func ensureAncestor(ctx context.Context, repo, base, head string) error {
	_, err := runGit(ctx, repo, "merge-base", "--is-ancestor", base, head)
	if err != nil {
		return fmt.Errorf("%q is not an ancestor of %q", base, head)
	}
	return nil
}

func currentBranch(ctx context.Context, repo string) (string, error) {
	out, err := runGit(ctx, repo, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	branch := strings.TrimSpace(out)
	if branch == "" {
		return "", fmt.Errorf("git rev-parse --abbrev-ref HEAD returned empty branch")
	}
	if branch == "HEAD" {
		return "", fmt.Errorf("repository is in a detached HEAD state")
	}
	return branch, nil
}

func listCommits(ctx context.Context, repo, trunk string) ([]Commit, error) {
	format := "%H%x1f%h%x1f%s%x1f%b%x1e"
	out, err := runGit(ctx, repo, "log", "--reverse", "--format="+format, trunk+"..HEAD")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(out) == "" {
		return nil, nil
	}

	records := strings.Split(out, "\x1e")
	commits := make([]Commit, 0, len(records))
	for _, record := range records {
		record = strings.TrimSpace(record)
		if record == "" {
			continue
		}
		fields := strings.Split(record, "\x1f")
		if len(fields) < 4 {
			return nil, fmt.Errorf("unexpected git log output")
		}
		body := strings.TrimSpace(fields[3])
		commits = append(commits, Commit{
			SHA:     strings.TrimSpace(fields[0]),
			Short:   strings.TrimSpace(fields[1]),
			Subject: strings.TrimSpace(fields[2]),
			Body:    body,
		})
	}

	return commits, nil
}
