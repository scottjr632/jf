package stack

import (
	"context"
	"fmt"
	"strings"
)

type resolvedStack struct {
	name           string
	stack          StackMeta
	headRef        string
	detached       bool
	stackHead      string
	effectiveTrunk string
	changed        bool
}

func resolveStack(ctx context.Context, repo string, cfg *Config, trunkOverride string) (resolvedStack, error) {
	applyDefaults(cfg)

	headRef, detached, err := currentHeadRef(ctx, repo)
	if err != nil {
		return resolvedStack{}, err
	}

	stackName := ""
	if detached {
		if cfg.CurrentStack != "" {
			if _, ok := cfg.Stacks[cfg.CurrentStack]; ok {
				stackName = cfg.CurrentStack
			}
		}
		if stackName == "" {
			name, err := findContainingBranch(ctx, repo, cfg.Trunk)
			if err != nil {
				return resolvedStack{}, err
			}
			if name == "" {
				name = "detached"
			}
			stackName = name
		}
	} else {
		stackName = headRef
	}

	stack, ok := cfg.Stacks[stackName]
	if !ok {
		stack = StackMeta{Trunk: cfg.Trunk, Order: []string{}, Commits: map[string]CommitMeta{}}
	}

	effectiveTrunk := strings.TrimSpace(trunkOverride)
	if effectiveTrunk == "" {
		effectiveTrunk = strings.TrimSpace(stack.Trunk)
	}
	if effectiveTrunk == "" {
		effectiveTrunk = cfg.Trunk
	}

	stackHeadRef, err := resolveStackHeadRef(ctx, repo, effectiveTrunk, headRef, detached)
	if err != nil {
		return resolvedStack{}, err
	}

	changed := false
	if !ok {
		commits, err := listCommitsRange(ctx, repo, effectiveTrunk, stackHeadRef)
		if err != nil {
			return resolvedStack{}, err
		}
		order := make([]string, 0, len(commits))
		commitMap := make(map[string]CommitMeta, len(commits))
		for _, commit := range commits {
			id, err := newUUID()
			if err != nil {
				return resolvedStack{}, err
			}
			order = append(order, id)
			commitMap[id] = CommitMeta{SHA: commit.SHA, Subject: commit.Subject, Body: commit.Body}
		}
		stack = StackMeta{Trunk: effectiveTrunk, Order: order, Commits: commitMap}
		changed = true
	}

	resolved := resolvedStack{
		name:           stackName,
		stack:          stack,
		headRef:        headRef,
		detached:       detached,
		stackHead:      stackHeadRef,
		effectiveTrunk: effectiveTrunk,
		changed:        changed,
	}

	updated, err := syncStackCurrent(ctx, repo, &resolved.stack)
	if err != nil {
		return resolvedStack{}, err
	}
	resolved.changed = resolved.changed || updated

	if resolved.changed {
		cfg.Stacks[stackName] = resolved.stack
		cfg.CurrentStack = stackName
	}

	return resolved, nil
}

func syncStackCurrent(ctx context.Context, repo string, stack *StackMeta) (bool, error) {
	if stack == nil {
		return false, nil
	}
	headSHA, err := currentSHA(ctx, repo)
	if err != nil {
		return false, err
	}

	if id := commitIDForSHA(*stack, headSHA); id != "" {
		if stack.Current != id {
			stack.Current = id
			return true, nil
		}
		return false, nil
	}

	origSHA, err := resolveOptionalSHA(ctx, repo, "ORIG_HEAD")
	if err != nil {
		return false, err
	}
	if origSHA == "" {
		return false, nil
	}

	if id := commitIDForSHA(*stack, origSHA); id != "" {
		commit, err := readCommit(ctx, repo, "HEAD")
		if err != nil {
			return false, err
		}
		stack.Commits[id] = CommitMeta{SHA: commit.SHA, Subject: commit.Subject, Body: commit.Body}
		stack.Current = id
		return true, nil
	}

	return false, nil
}

func commitIDForSHA(stack StackMeta, sha string) string {
	for id, meta := range stack.Commits {
		if meta.SHA == sha {
			return id
		}
	}
	return ""
}

func stackCommits(stack StackMeta) []Commit {
	commits := make([]Commit, 0, len(stack.Order))
	for _, id := range stack.Order {
		meta, ok := stack.Commits[id]
		if !ok {
			continue
		}
		commits = append(commits, Commit{
			SHA:     meta.SHA,
			Short:   shortSHA(meta.SHA),
			Subject: meta.Subject,
			Body:    meta.Body,
		})
	}
	return commits
}

func shortSHA(sha string) string {
	sha = strings.TrimSpace(sha)
	if len(sha) <= 7 {
		return sha
	}
	return sha[:7]
}

func refreshStackFromGit(ctx context.Context, repo, trunk, head string, stack *StackMeta) error {
	commits, err := listCommitsRange(ctx, repo, trunk, head)
	if err != nil {
		return err
	}
	if head == "" {
		return fmt.Errorf("stack head ref is empty")
	}
	if len(commits) != len(stack.Order) {
		return fmt.Errorf("stack size mismatch after rebase")
	}
	for i, commit := range commits {
		id := stack.Order[i]
		stack.Commits[id] = CommitMeta{SHA: commit.SHA, Subject: commit.Subject, Body: commit.Body}
	}
	if headSHA, err := currentSHA(ctx, repo); err == nil {
		if id := commitIDForSHA(*stack, headSHA); id != "" {
			stack.Current = id
		}
	}
	return nil
}
