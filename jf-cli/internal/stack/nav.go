package stack

import (
	"context"
	"fmt"
	"strings"
)

// NextCommit returns the next commit in the current stack.
func NextCommit(ctx context.Context, repo string, cfg *Config, trunkOverride string) (Commit, error) {
	return navigateStack(ctx, repo, cfg, trunkOverride, 1)
}

// PrevCommit returns the previous commit in the current stack.
func PrevCommit(ctx context.Context, repo string, cfg *Config, trunkOverride string) (Commit, error) {
	return navigateStack(ctx, repo, cfg, trunkOverride, -1)
}

func navigateStack(ctx context.Context, repo string, cfg *Config, trunkOverride string, delta int) (Commit, error) {
	resolved, err := resolveStack(ctx, repo, cfg, trunkOverride)
	if err != nil {
		return Commit{}, err
	}
	if resolved.changed {
		if err := Save(ctx, repo, *cfg); err != nil {
			return Commit{}, err
		}
	}
	if len(resolved.stack.Order) == 0 {
		return Commit{}, fmt.Errorf("no commits in stack")
	}
	currentID := strings.TrimSpace(resolved.stack.Current)
	if currentID == "" {
		return Commit{}, fmt.Errorf("current commit not tracked in stack")
	}
	index := indexOfStackID(resolved.stack.Order, currentID)
	if index == -1 {
		return Commit{}, fmt.Errorf("current commit not tracked in stack")
	}
	targetIndex := index + delta
	if targetIndex < 0 || targetIndex >= len(resolved.stack.Order) {
		return Commit{}, fmt.Errorf("no further commit in that direction")
	}
	meta, ok := resolved.stack.Commits[resolved.stack.Order[targetIndex]]
	if !ok {
		return Commit{}, fmt.Errorf("target commit missing from stack metadata")
	}
	return Commit{
		SHA:     meta.SHA,
		Short:   shortSHA(meta.SHA),
		Subject: meta.Subject,
		Body:    meta.Body,
	}, nil
}
