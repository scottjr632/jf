package stack

import (
	"context"
	"fmt"
	"strings"
)

// NextCommit returns the next commit in the current stack.
func NextCommit(ctx context.Context, repo string, cfg *Config, trunkOverride string) (Commit, error) {
	choices, err := NextCommits(ctx, repo, cfg, trunkOverride)
	if err != nil {
		return Commit{}, err
	}
	if len(choices) == 1 {
		return choices[0].Commit, nil
	}
	return Commit{}, fmt.Errorf("multiple next commits available")
}

// PrevCommit returns the previous commit in the current stack.
func PrevCommit(ctx context.Context, repo string, cfg *Config, trunkOverride string) (Commit, error) {
	return previousCommit(ctx, repo, cfg, trunkOverride)
}

// NextCommits returns the next commit choices in the current stack.
func NextCommits(ctx context.Context, repo string, cfg *Config, trunkOverride string) ([]StackItem, error) {
	resolved, err := resolveStack(ctx, repo, cfg, trunkOverride)
	if err != nil {
		return nil, err
	}
	if resolved.changed {
		if err := Save(ctx, repo, *cfg); err != nil {
			return nil, err
		}
	}
	if len(resolved.stack.Commits) == 0 {
		return nil, fmt.Errorf("no commits in stack")
	}
	currentID := strings.TrimSpace(resolved.stack.Current)
	if currentID == "" {
		return nil, fmt.Errorf("current commit not tracked in stack")
	}
	if _, ok := resolved.stack.Commits[currentID]; !ok {
		return nil, fmt.Errorf("current commit not tracked in stack")
	}

	children := stackChildren(resolved.stack)
	childIDs := children[currentID]
	if len(childIDs) == 0 {
		return nil, fmt.Errorf("no further commit in that direction")
	}
	items := make([]StackItem, 0, len(childIDs))
	for _, id := range childIDs {
		meta, ok := resolved.stack.Commits[id]
		if !ok {
			continue
		}
		items = append(items, StackItem{
			ID:       id,
			ParentID: strings.TrimSpace(meta.Parent),
			Commit: Commit{
				SHA:     meta.SHA,
				Short:   shortSHA(meta.SHA),
				Subject: meta.Subject,
				Body:    meta.Body,
			},
		})
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("no further commit in that direction")
	}
	return items, nil
}

func previousCommit(ctx context.Context, repo string, cfg *Config, trunkOverride string) (Commit, error) {
	resolved, err := resolveStack(ctx, repo, cfg, trunkOverride)
	if err != nil {
		return Commit{}, err
	}
	if resolved.changed {
		if err := Save(ctx, repo, *cfg); err != nil {
			return Commit{}, err
		}
	}
	if len(resolved.stack.Commits) == 0 {
		return Commit{}, fmt.Errorf("no commits in stack")
	}
	currentID := strings.TrimSpace(resolved.stack.Current)
	if currentID == "" {
		return Commit{}, fmt.Errorf("current commit not tracked in stack")
	}
	meta, ok := resolved.stack.Commits[currentID]
	if !ok {
		return Commit{}, fmt.Errorf("current commit not tracked in stack")
	}
	parentID := strings.TrimSpace(meta.Parent)
	if parentID == "" {
		return Commit{}, fmt.Errorf("no further commit in that direction")
	}
	parentMeta, ok := resolved.stack.Commits[parentID]
	if !ok {
		return Commit{}, fmt.Errorf("target commit missing from stack metadata")
	}
	return Commit{
		SHA:     parentMeta.SHA,
		Short:   shortSHA(parentMeta.SHA),
		Subject: parentMeta.Subject,
		Body:    parentMeta.Body,
	}, nil
}
