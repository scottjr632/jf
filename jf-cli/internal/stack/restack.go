package stack

import (
	"context"
	"fmt"
	"strings"
)

// Restack ensures commits are stacked in order on top of trunk.
func Restack(ctx context.Context, repo string, cfg *Config, trunkOverride string) error {
	resolved, err := resolveStack(ctx, repo, cfg, trunkOverride)
	if err != nil {
		return err
	}
	if len(resolved.stack.Commits) == 0 {
		return nil
	}

	trunkSHA, err := resolveRefSHA(ctx, repo, resolved.effectiveTrunk)
	if err != nil {
		return err
	}
	order := stackOrder(resolved.stack)
	for _, id := range order {
		meta, ok := resolved.stack.Commits[id]
		if !ok {
			return fmt.Errorf("missing metadata for stack commit")
		}
		expectedParent := trunkSHA
		parentID := strings.TrimSpace(meta.Parent)
		if parentID != "" {
			parentMeta, ok := resolved.stack.Commits[parentID]
			if !ok {
				return fmt.Errorf("missing parent metadata for stack commit")
			}
			expectedParent = parentMeta.SHA
		}
		parentSHA, err := commitParentSHA(ctx, repo, meta.SHA)
		if err != nil {
			return err
		}
		if parentSHA != expectedParent {
			if err := rebaseDescendants(ctx, repo, parentSHA, expectedParent, resolved.stackHead); err != nil {
				return err
			}
			if err := refreshStackFromGit(ctx, repo, resolved.effectiveTrunk, resolved.stackHead, &resolved.stack); err != nil {
				return err
			}
			cfg.Stacks[resolved.name] = resolved.stack
			cfg.CurrentStack = resolved.name
			return Save(ctx, repo, *cfg)
		}
	}

	if resolved.changed {
		cfg.Stacks[resolved.name] = resolved.stack
		cfg.CurrentStack = resolved.name
		return Save(ctx, repo, *cfg)
	}
	return nil
}

func resolveRefSHA(ctx context.Context, repo, ref string) (string, error) {
	if strings.TrimSpace(ref) == "" {
		return "", fmt.Errorf("ref is required")
	}
	out, err := runGit(ctx, repo, "rev-parse", ref)
	if err != nil {
		return "", err
	}
	sha := strings.TrimSpace(out)
	if sha == "" {
		return "", fmt.Errorf("git rev-parse %s returned empty output", ref)
	}
	return sha, nil
}

func commitParentSHA(ctx context.Context, repo, sha string) (string, error) {
	if strings.TrimSpace(sha) == "" {
		return "", fmt.Errorf("commit sha is required")
	}
	out, err := runGit(ctx, repo, "rev-parse", sha+"^")
	if err != nil {
		return "", err
	}
	parent := strings.TrimSpace(out)
	if parent == "" {
		return "", fmt.Errorf("git rev-parse %s^ returned empty output", sha)
	}
	return parent, nil
}
