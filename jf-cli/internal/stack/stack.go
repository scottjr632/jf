package stack

import (
	"context"
	"fmt"
	"strings"

	"github.com/scottjr632/jf-cli/internal/git"
)

var runGitPassthrough = git.RunPassthrough

// Commit describes a stack item derived from git history.
type Commit struct {
	SHA     string
	Short   string
	Subject string
	Body    string
}

// Stack represents the current stack of commits from trunk to head.
type Stack struct {
	Trunk   string
	Head    string
	Commits []Commit
}

// StackItem describes a stack commit with its stable id.
type StackItem struct {
	ID     string
	Commit Commit
}

// StackDetails represents the current stack with stable ids.
type StackDetails struct {
	Trunk string
	Head  string
	Items []StackItem
}

// CurrentStack returns commits for the current stack using metadata.
func CurrentStack(ctx context.Context, repo string, cfg *Config, trunkOverride string) (Stack, error) {
	resolved, err := resolveStack(ctx, repo, cfg, trunkOverride)
	if err != nil {
		return Stack{}, err
	}
	if resolved.changed {
		if err := Save(ctx, repo, *cfg); err != nil {
			return Stack{}, err
		}
	}

	headLabel := resolved.headRef
	if resolved.detached {
		label, err := currentShortSHA(ctx, repo)
		if err != nil {
			return Stack{}, err
		}
		headLabel = label
	}

	return Stack{
		Trunk:   resolved.effectiveTrunk,
		Head:    headLabel,
		Commits: stackCommits(resolved.stack),
	}, nil
}

// CurrentStackDetails returns commits for the current stack with stable ids.
func CurrentStackDetails(ctx context.Context, repo string, cfg *Config, trunkOverride string) (StackDetails, error) {
	resolved, err := resolveStack(ctx, repo, cfg, trunkOverride)
	if err != nil {
		return StackDetails{}, err
	}
	if resolved.changed {
		if err := Save(ctx, repo, *cfg); err != nil {
			return StackDetails{}, err
		}
	}

	headLabel := resolved.headRef
	if resolved.detached {
		label, err := currentShortSHA(ctx, repo)
		if err != nil {
			return StackDetails{}, err
		}
		headLabel = label
	}

	items := make([]StackItem, 0, len(resolved.stack.Order))
	for _, id := range resolved.stack.Order {
		meta, ok := resolved.stack.Commits[id]
		if !ok {
			continue
		}
		items = append(items, StackItem{
			ID: id,
			Commit: Commit{
				SHA:     meta.SHA,
				Short:   shortSHA(meta.SHA),
				Subject: meta.Subject,
				Body:    meta.Body,
			},
		})
	}

	return StackDetails{
		Trunk: resolved.effectiveTrunk,
		Head:  headLabel,
		Items: items,
	}, nil
}

// SyncCurrent updates the current stack pointer based on HEAD.
func SyncCurrent(ctx context.Context, repo string, cfg *Config, trunkOverride string) error {
	resolved, err := resolveStack(ctx, repo, cfg, trunkOverride)
	if err != nil {
		return err
	}
	if resolved.changed {
		return Save(ctx, repo, *cfg)
	}
	return nil
}

// RecordCommit appends a new commit to the current stack.
func RecordCommit(ctx context.Context, repo string, cfg *Config, trunkOverride string) error {
	resolved, err := resolveStack(ctx, repo, cfg, trunkOverride)
	if err != nil {
		return err
	}
	commit, err := readCommit(ctx, repo, "HEAD")
	if err != nil {
		return err
	}

	if id := commitIDForSHA(resolved.stack, commit.SHA); id != "" {
		resolved.stack.Current = id
		resolved.changed = true
		cfg.Stacks[resolved.name] = resolved.stack
		return Save(ctx, repo, *cfg)
	}

	id, err := newUUID()
	if err != nil {
		return err
	}
	resolved.stack.Order = append(resolved.stack.Order, id)
	resolved.stack.Commits[id] = CommitMeta{SHA: commit.SHA, Subject: commit.Subject, Body: commit.Body}
	resolved.stack.Current = id
	resolved.changed = true
	cfg.Stacks[resolved.name] = resolved.stack
	cfg.CurrentStack = resolved.name

	return Save(ctx, repo, *cfg)
}

// RecordAmend updates stack metadata after an amend and rebases descendants.
func RecordAmend(ctx context.Context, repo string, cfg *Config, trunkOverride string) error {
	resolved, err := resolveStack(ctx, repo, cfg, trunkOverride)
	if err != nil {
		return err
	}

	headCommit, err := readCommit(ctx, repo, "HEAD")
	if err != nil {
		return err
	}

	origSHA, err := resolveOptionalSHA(ctx, repo, "ORIG_HEAD")
	if err != nil {
		return err
	}
	if origSHA == "" {
		return SyncCurrent(ctx, repo, cfg, trunkOverride)
	}

	amendedID := commitIDForSHA(resolved.stack, origSHA)
	if amendedID == "" {
		amendedID = commitIDForSHA(resolved.stack, headCommit.SHA)
	}
	if amendedID == "" {
		return SyncCurrent(ctx, repo, cfg, trunkOverride)
	}

	resolved.stack.Commits[amendedID] = CommitMeta{SHA: headCommit.SHA, Subject: headCommit.Subject, Body: headCommit.Body}
	resolved.stack.Current = amendedID

	amendedIndex := indexOfStackID(resolved.stack.Order, amendedID)
	if amendedIndex != -1 && amendedIndex < len(resolved.stack.Order)-1 {
		if err := rebaseDescendants(ctx, repo, origSHA, headCommit.SHA, resolved.stackHead); err != nil {
			return err
		}
		if err := refreshStackFromGit(ctx, repo, resolved.effectiveTrunk, resolved.stackHead, &resolved.stack); err != nil {
			return err
		}
	}

	cfg.Stacks[resolved.name] = resolved.stack
	cfg.CurrentStack = resolved.name
	return Save(ctx, repo, *cfg)
}

func rebaseDescendants(ctx context.Context, repo, oldSHA, newSHA, headRef string) error {
	return runGitPassthrough(ctx, repo, "rebase", "--onto", newSHA, oldSHA, headRef)
}

func indexOfStackID(order []string, id string) int {
	for i, candidate := range order {
		if candidate == id {
			return i
		}
	}
	return -1
}

func currentHeadRef(ctx context.Context, repo string) (string, bool, error) {
	out, err := runGit(ctx, repo, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", false, err
	}
	branch := strings.TrimSpace(out)
	if branch == "" {
		return "", false, fmt.Errorf("git rev-parse --abbrev-ref HEAD returned empty branch")
	}
	if branch == "HEAD" || strings.HasPrefix(branch, "(") || strings.Contains(branch, "detached") {
		return "HEAD", true, nil
	}
	return branch, false, nil
}

func currentSHA(ctx context.Context, repo string) (string, error) {
	out, err := runGit(ctx, repo, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	sha := strings.TrimSpace(out)
	if sha == "" {
		return "", fmt.Errorf("git rev-parse HEAD returned empty output")
	}
	return sha, nil
}

func currentShortSHA(ctx context.Context, repo string) (string, error) {
	out, err := runGit(ctx, repo, "rev-parse", "--short", "HEAD")
	if err != nil {
		return "", err
	}
	sha := strings.TrimSpace(out)
	if sha == "" {
		return "", fmt.Errorf("git rev-parse --short HEAD returned empty output")
	}
	return sha, nil
}

func resolveOptionalSHA(ctx context.Context, repo, ref string) (string, error) {
	out, err := runGit(ctx, repo, "rev-parse", "--verify", ref)
	if err != nil {
		return "", nil
	}
	sha := strings.TrimSpace(out)
	return sha, nil
}

func readCommit(ctx context.Context, repo, ref string) (Commit, error) {
	format := "%H%x1f%h%x1f%s%x1f%b%x1e"
	out, err := runGit(ctx, repo, "log", "-1", "--format="+format, ref)
	if err != nil {
		return Commit{}, err
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return Commit{}, fmt.Errorf("git log returned empty output")
	}
	parts := strings.Split(out, "\x1f")
	if len(parts) < 4 {
		return Commit{}, fmt.Errorf("unexpected git log output")
	}
	return Commit{
		SHA:     strings.TrimSpace(parts[0]),
		Short:   strings.TrimSpace(parts[1]),
		Subject: strings.TrimSpace(parts[2]),
		Body:    strings.TrimSpace(parts[3]),
	}, nil
}

func resolveStackHeadRef(ctx context.Context, repo, trunk, headRef string, detached bool) (string, error) {
	if !detached {
		return headRef, nil
	}
	branch, err := findContainingBranch(ctx, repo, trunk)
	if err != nil {
		return "", err
	}
	if branch == "" {
		return "HEAD", nil
	}
	return branch, nil
}

func findContainingBranch(ctx context.Context, repo, trunk string) (string, error) {
	out, err := runGit(ctx, repo, "branch", "--contains", "HEAD", "--sort=-committerdate", "--format=%(refname:short)")
	if err != nil {
		return "", err
	}
	lines := strings.Split(out, "\n")
	trunk = strings.TrimSpace(trunk)
	candidates := make([]string, 0, len(lines))
	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		if name == "HEAD" || strings.HasPrefix(name, "(") || strings.Contains(name, "detached") {
			continue
		}
		candidates = append(candidates, name)
	}
	if len(candidates) == 0 {
		return "", nil
	}
	for _, name := range candidates {
		if trunk != "" && name == trunk {
			continue
		}
		return name, nil
	}
	return trunk, nil
}
