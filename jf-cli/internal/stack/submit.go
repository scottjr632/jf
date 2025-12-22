package stack

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/scottjr632/jf-cli/internal/gh"
)

var runGh = gh.Run

// SubmitOptions controls stack submission behavior.
type SubmitOptions struct {
	Trunk        string
	Remote       string
	BranchPrefix string
}

// SubmitResult captures the outcome for a commit PR.
type SubmitResult struct {
	Commit Commit
	Branch string
	Base   string
	Number int
	Action SubmitAction
}

// SubmitAction describes how a PR was affected.
type SubmitAction string

const (
	SubmitCreated   SubmitAction = "created"
	SubmitUpdated   SubmitAction = "updated"
	SubmitUnchanged SubmitAction = "unchanged"
)

type prInfo struct {
	Number      int    `json:"number"`
	BaseRefName string `json:"baseRefName"`
	HeadRefName string `json:"headRefName"`
	Title       string `json:"title"`
}

// SubmitCurrent creates or updates PRs for the current stack.
func SubmitCurrent(ctx context.Context, repo string, cfg Config, opts SubmitOptions) ([]SubmitResult, error) {
	applyDefaults(&cfg)
	if strings.TrimSpace(opts.Trunk) == "" {
		opts.Trunk = cfg.Trunk
	}
	if strings.TrimSpace(opts.Remote) == "" {
		opts.Remote = cfg.Remote
	}
	if strings.TrimSpace(opts.BranchPrefix) == "" {
		opts.BranchPrefix = cfg.BranchPrefix
	}

	stackInfo, err := CurrentStack(ctx, repo, opts.Trunk)
	if err != nil {
		return nil, err
	}
	if len(stackInfo.Commits) == 0 {
		return nil, fmt.Errorf("no commits to submit")
	}

	root, err := repoRoot(ctx, repo)
	if err != nil {
		return nil, err
	}

	results := make([]SubmitResult, 0, len(stackInfo.Commits))
	base := stackInfo.Trunk

	for i, commit := range stackInfo.Commits {
		branch := branchNameForCommit(opts.BranchPrefix, i+1, commit)
		if err := createOrUpdateBranch(ctx, repo, branch, commit.SHA); err != nil {
			return nil, err
		}
		if err := pushBranch(ctx, repo, opts.Remote, branch); err != nil {
			return nil, err
		}

		pr, err := findOpenPR(ctx, root, branch)
		if err != nil {
			return nil, err
		}

		result := SubmitResult{Commit: commit, Branch: branch, Base: base}
		title := commitTitle(commit)
		body := commitBody(commit)
		if pr == nil {
			if err := createPR(ctx, root, branch, base, title, body); err != nil {
				return nil, err
			}
			pr, err = findOpenPR(ctx, root, branch)
			if err != nil {
				return nil, err
			}
			if pr == nil {
				return nil, fmt.Errorf("created PR for %q but could not find it", branch)
			}
			result.Number = pr.Number
			result.Action = SubmitCreated
		} else {
			result.Number = pr.Number
			updated := false
			if pr.BaseRefName != base {
				if err := updatePRBase(ctx, root, pr.Number, base); err != nil {
					return nil, err
				}
				updated = true
			}
			if pr.Title != title {
				if err := updatePRTitle(ctx, root, pr.Number, title); err != nil {
					return nil, err
				}
				updated = true
			}
			if updated {
				result.Action = SubmitUpdated
			} else {
				result.Action = SubmitUnchanged
			}
		}

		results = append(results, result)
		base = branch
	}

	return results, nil
}

func createOrUpdateBranch(ctx context.Context, repo, branch, sha string) error {
	_, err := runGit(ctx, repo, "branch", "-f", branch, sha)
	return err
}

func pushBranch(ctx context.Context, repo, remote, branch string) error {
	if strings.TrimSpace(remote) == "" {
		return fmt.Errorf("remote is required")
	}
	_, err := runGit(ctx, repo, "push", "-f", remote, branch)
	return err
}

func findOpenPR(ctx context.Context, repoRoot, branch string) (*prInfo, error) {
	out, err := runGh(ctx, repoRoot, "pr", "list", "--head", branch, "--state", "open", "--json", "number,baseRefName,headRefName,title")
	if err != nil {
		return nil, err
	}
	var prs []prInfo
	if err := json.Unmarshal([]byte(out), &prs); err != nil {
		return nil, fmt.Errorf("parse gh output: %w", err)
	}
	if len(prs) == 0 {
		return nil, nil
	}
	if len(prs) > 1 {
		return nil, fmt.Errorf("multiple open PRs found for %q", branch)
	}
	return &prs[0], nil
}

func createPR(ctx context.Context, repoRoot, branch, base, title, body string) error {
	_, err := runGh(ctx, repoRoot, "pr", "create", "--head", branch, "--base", base, "--title", title, "--body", body)
	return err
}

func updatePRBase(ctx context.Context, repoRoot string, number int, base string) error {
	if number == 0 {
		return fmt.Errorf("missing PR number")
	}
	_, err := runGh(ctx, repoRoot, "pr", "edit", fmt.Sprintf("%d", number), "--base", base)
	return err
}

func updatePRTitle(ctx context.Context, repoRoot string, number int, title string) error {
	if number == 0 {
		return fmt.Errorf("missing PR number")
	}
	_, err := runGh(ctx, repoRoot, "pr", "edit", fmt.Sprintf("%d", number), "--title", title)
	return err
}

func commitTitle(commit Commit) string {
	if strings.TrimSpace(commit.Subject) != "" {
		return commit.Subject
	}
	return commit.Short
}

func commitBody(commit Commit) string {
	if strings.TrimSpace(commit.Body) != "" {
		return commit.Body
	}
	return commitTitle(commit)
}

func branchNameForCommit(prefix string, index int, commit Commit) string {
	cleanPrefix := strings.Trim(prefix, "/")
	if cleanPrefix == "" {
		cleanPrefix = defaultBranchPrefix
	}
	slug := slugify(commit.Subject)
	if slug == "" {
		slug = "commit"
	}
	return fmt.Sprintf("%s/%02d-%s-%s", cleanPrefix, index, slug, commit.Short)
}

var slugPattern = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(input string) string {
	lower := strings.ToLower(strings.TrimSpace(input))
	if lower == "" {
		return ""
	}
	replaced := slugPattern.ReplaceAllString(lower, "-")
	replaced = strings.Trim(replaced, "-")
	return replaced
}
