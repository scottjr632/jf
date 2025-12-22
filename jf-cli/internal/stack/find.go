package stack

import (
	"sort"
	"strings"
)

// FindStackCommit locates a commit by full SHA in tracked stacks.
func FindStackCommit(cfg *Config, sha string) (string, string, CommitMeta, bool) {
	if cfg == nil {
		return "", "", CommitMeta{}, false
	}
	applyDefaults(cfg)

	sha = strings.TrimSpace(sha)
	if sha == "" {
		return "", "", CommitMeta{}, false
	}

	if cfg.CurrentStack != "" {
		if stack, ok := cfg.Stacks[cfg.CurrentStack]; ok {
			if id, meta, ok := commitMetaForSHA(stack, sha); ok {
				return cfg.CurrentStack, id, meta, true
			}
		}
	}

	names := make([]string, 0, len(cfg.Stacks))
	for name := range cfg.Stacks {
		if name == cfg.CurrentStack {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		stack := cfg.Stacks[name]
		if id, meta, ok := commitMetaForSHA(stack, sha); ok {
			return name, id, meta, true
		}
	}

	return "", "", CommitMeta{}, false
}

func commitMetaForSHA(stack StackMeta, sha string) (string, CommitMeta, bool) {
	id := commitIDForSHA(stack, sha)
	if id == "" {
		return "", CommitMeta{}, false
	}
	meta, ok := stack.Commits[id]
	if !ok {
		return "", CommitMeta{}, false
	}
	return id, meta, true
}
