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

type stackPR struct {
	Number   int
	Title    string
	Position int
}

type prComment struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

const stackCommentMarker = "<!-- jf-stack-info -->"

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

	resolved, err := resolveStack(ctx, repo, &cfg, opts.Trunk)
	if err != nil {
		return nil, err
	}
	if resolved.changed {
		if err := Save(ctx, repo, cfg); err != nil {
			return nil, err
		}
	}
	if len(resolved.stack.Order) == 0 {
		return nil, fmt.Errorf("no commits to submit")
	}

	root, err := repoRoot(ctx, repo)
	if err != nil {
		return nil, err
	}

	results := make([]SubmitResult, 0, len(resolved.stack.Order))
	stackPRs := make([]stackPR, 0, len(resolved.stack.Order))
	base := resolved.effectiveTrunk

	for i, id := range resolved.stack.Order {
		meta, ok := resolved.stack.Commits[id]
		if !ok {
			return nil, fmt.Errorf("missing metadata for stack commit")
		}
		commit := Commit{
			SHA:     meta.SHA,
			Short:   shortSHA(meta.SHA),
			Subject: meta.Subject,
			Body:    meta.Body,
		}
		position := i + 1
		branch := BranchNameForCommit(opts.BranchPrefix, i+1, id, commit)
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

		stackPRs = append(stackPRs, stackPR{Number: pr.Number, Title: title, Position: position})
		results = append(results, result)
		base = branch
	}

	if err := updateStackComments(ctx, root, resolved.name, Stack{Trunk: resolved.effectiveTrunk, Head: resolved.headRef}, stackPRs); err != nil {
		return nil, err
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

func BranchNameForCommit(prefix string, index int, id string, commit Commit) string {
	cleanPrefix := strings.Trim(prefix, "/")
	if cleanPrefix == "" {
		cleanPrefix = defaultBranchPrefix
	}
	suffix := stableBranchSuffix(id, commit)
	return fmt.Sprintf("%s/%02d-%s", cleanPrefix, index, suffix)
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

func stableBranchSuffix(id string, commit Commit) string {
	clean := strings.ReplaceAll(strings.TrimSpace(id), "-", "")
	if clean != "" {
		if len(clean) > 8 {
			return clean[:8]
		}
		return clean
	}
	slug := slugify(commit.Subject)
	if slug == "" {
		return "commit"
	}
	if len(slug) > 16 {
		return slug[:16]
	}
	return slug
}

func updateStackComments(ctx context.Context, repoRoot string, stackName string, stackInfo Stack, prs []stackPR) error {
	if len(prs) == 0 {
		return nil
	}
	repoName, err := repoNameWithOwner(ctx, repoRoot)
	if err != nil {
		return err
	}
	if strings.TrimSpace(repoName) == "" {
		return fmt.Errorf("missing repo name")
	}
	total := len(prs)
	for _, pr := range prs {
		message := stackCommentBody(stackName, stackInfo, pr.Position, total, prs)
		if err := upsertStackComment(ctx, repoRoot, repoName, pr.Number, message); err != nil {
			return err
		}
	}
	return nil
}

func repoNameWithOwner(ctx context.Context, repoRoot string) (string, error) {
	out, err := runGh(ctx, repoRoot, "repo", "view", "--json", "nameWithOwner", "-q", ".nameWithOwner")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func stackCommentBody(stackName string, stackInfo Stack, position, total int, prs []stackPR) string {
	name := strings.TrimSpace(stackName)
	if name == "" {
		name = stackInfo.Head
	}
	var builder strings.Builder
	builder.WriteString(stackCommentMarker)
	builder.WriteString("\n")
	fmt.Fprintf(&builder, "Stack: %s (trunk: %s) [%d/%d]\n\n", name, stackInfo.Trunk, position, total)
	builder.WriteString("PRs:\n")
	for _, pr := range prs {
		line := fmt.Sprintf("%d. #%d - %s", pr.Position, pr.Number, strings.TrimSpace(pr.Title))
		if pr.Position == position {
			line += " (current)"
		}
		builder.WriteString(line)
		builder.WriteString("\n")
	}
	return strings.TrimSpace(builder.String())
}

func upsertStackComment(ctx context.Context, repoRoot, repoName string, number int, message string) error {
	if number == 0 {
		return fmt.Errorf("missing PR number")
	}
	commentID, err := findStackCommentID(ctx, repoRoot, repoName, number)
	if err != nil {
		return err
	}
	if commentID != 0 {
		return updateStackComment(ctx, repoRoot, repoName, commentID, message)
	}
	_, err = runGh(ctx, repoRoot, "pr", "comment", fmt.Sprintf("%d", number), "--body", message)
	return err
}

func findStackCommentID(ctx context.Context, repoRoot, repoName string, number int) (int, error) {
	if strings.TrimSpace(repoName) == "" {
		return 0, fmt.Errorf("missing repo name")
	}
	out, err := runGh(ctx, repoRoot, "api", fmt.Sprintf("repos/%s/issues/%d/comments", repoName, number))
	if err != nil {
		return 0, err
	}
	var comments []prComment
	if err := json.Unmarshal([]byte(out), &comments); err != nil {
		return 0, fmt.Errorf("parse gh output: %w", err)
	}
	for _, comment := range comments {
		if strings.Contains(comment.Body, stackCommentMarker) {
			return comment.ID, nil
		}
	}
	return 0, nil
}

func updateStackComment(ctx context.Context, repoRoot, repoName string, commentID int, message string) error {
	if commentID == 0 {
		return fmt.Errorf("missing comment id")
	}
	_, err := runGh(ctx, repoRoot, "api", "-X", "PATCH", fmt.Sprintf("repos/%s/issues/comments/%d", repoName, commentID), "-f", "body="+message)
	return err
}
