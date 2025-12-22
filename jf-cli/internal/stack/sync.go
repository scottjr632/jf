package stack

import "context"

// SyncStack refreshes stack metadata from git history.
func SyncStack(ctx context.Context, repo string, cfg *Config, trunkOverride string) error {
	resolved, err := resolveStack(ctx, repo, cfg, trunkOverride)
	if err != nil {
		return err
	}
	if len(resolved.stack.Order) == 0 {
		return nil
	}
	if err := refreshStackFromGit(ctx, repo, resolved.effectiveTrunk, resolved.stackHead, &resolved.stack); err != nil {
		return err
	}
	cfg.Stacks[resolved.name] = resolved.stack
	cfg.CurrentStack = resolved.name
	return Save(ctx, repo, *cfg)
}
