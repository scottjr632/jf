package stack

import (
	"context"
	"fmt"
	"sort"
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
		preferredIDs := map[string]string{}
		if cfg.CurrentStack != "" {
			if existing, ok := cfg.Stacks[cfg.CurrentStack]; ok {
				preferredIDs = commitIDsBySHA(existing)
			}
		}
		allIDs := commitIDsBySHAFromConfig(cfg)
		order := make([]string, 0, len(commits))
		commitMap := make(map[string]CommitMeta, len(commits))
		idBySHA := make(map[string]string, len(commits))
		for _, node := range commits {
			commit := node.Commit
			id := preferredIDs[commit.SHA]
			if id == "" {
				id = allIDs[commit.SHA]
			}
			if id == "" {
				id, err = newUUID()
				if err != nil {
					return resolvedStack{}, err
				}
			}
			order = append(order, id)
			commitMap[id] = CommitMeta{SHA: commit.SHA, Subject: commit.Subject, Body: commit.Body}
			idBySHA[commit.SHA] = id
		}
		for _, node := range commits {
			commit := node.Commit
			id := idBySHA[commit.SHA]
			meta := commitMap[id]
			meta.Parent = pickParentID(node.Parents, idBySHA)
			commitMap[id] = meta
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

func commitIDsBySHA(stack StackMeta) map[string]string {
	order := stackOrder(stack)
	ids := make(map[string]string, len(order))
	for _, id := range order {
		meta, ok := stack.Commits[id]
		if !ok {
			continue
		}
		sha := strings.TrimSpace(meta.SHA)
		if sha == "" {
			continue
		}
		if _, exists := ids[sha]; !exists {
			ids[sha] = id
		}
	}
	return ids
}

func commitIDsBySHAFromConfig(cfg *Config) map[string]string {
	if cfg == nil {
		return map[string]string{}
	}
	names := make([]string, 0, len(cfg.Stacks))
	for name := range cfg.Stacks {
		names = append(names, name)
	}
	sort.Strings(names)

	ids := make(map[string]string)
	for _, name := range names {
		stack := cfg.Stacks[name]
		for sha, id := range commitIDsBySHA(stack) {
			if _, exists := ids[sha]; !exists {
				ids[sha] = id
			}
		}
	}
	return ids
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
	order := stackOrder(stack)
	commits := make([]Commit, 0, len(order))
	for _, id := range order {
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
	order := stackOrder(*stack)
	if len(commits) != len(order) {
		return fmt.Errorf("stack size mismatch after rebase")
	}
	newCommits := make(map[string]CommitMeta, len(commits))
	idBySHA := make(map[string]string, len(commits))
	for i, node := range commits {
		commit := node.Commit
		id := order[i]
		newCommits[id] = CommitMeta{SHA: commit.SHA, Subject: commit.Subject, Body: commit.Body}
		idBySHA[commit.SHA] = id
	}
	for _, node := range commits {
		commit := node.Commit
		id := idBySHA[commit.SHA]
		meta := newCommits[id]
		meta.Parent = pickParentID(node.Parents, idBySHA)
		newCommits[id] = meta
	}
	stack.Order = order
	stack.Commits = newCommits
	if headSHA, err := currentSHA(ctx, repo); err == nil {
		if id := commitIDForSHA(*stack, headSHA); id != "" {
			stack.Current = id
		}
	}
	return nil
}

func stackOrder(stack StackMeta) []string {
	if len(stack.Order) != 0 {
		return append([]string{}, stack.Order...)
	}
	if len(stack.Commits) == 0 {
		return nil
	}
	children := make(map[string][]string, len(stack.Commits))
	roots := make([]string, 0, len(stack.Commits))
	for id, meta := range stack.Commits {
		parent := strings.TrimSpace(meta.Parent)
		if parent == "" || stack.Commits[parent].SHA == "" {
			roots = append(roots, id)
			continue
		}
		children[parent] = append(children[parent], id)
	}
	sort.Strings(roots)
	for parent, ids := range children {
		sort.Strings(ids)
		children[parent] = ids
	}
	order := make([]string, 0, len(stack.Commits))
	var visit func(id string)
	visit = func(id string) {
		order = append(order, id)
		for _, child := range children[id] {
			visit(child)
		}
	}
	for _, root := range roots {
		visit(root)
	}
	return order
}

func stackChildren(stack StackMeta) map[string][]string {
	children := make(map[string][]string, len(stack.Commits))
	for id, meta := range stack.Commits {
		parent := strings.TrimSpace(meta.Parent)
		if parent == "" {
			continue
		}
		children[parent] = append(children[parent], id)
	}
	if len(stack.Order) == 0 {
		for parent, ids := range children {
			sort.Strings(ids)
			children[parent] = ids
		}
		return children
	}
	index := make(map[string]int, len(stack.Order))
	for i, id := range stack.Order {
		index[id] = i
	}
	for parent, ids := range children {
		sort.Slice(ids, func(i, j int) bool {
			return index[ids[i]] < index[ids[j]]
		})
		children[parent] = ids
	}
	return children
}

func stackRoots(stack StackMeta) []string {
	roots := make([]string, 0, len(stack.Commits))
	for id, meta := range stack.Commits {
		parent := strings.TrimSpace(meta.Parent)
		if parent == "" || stack.Commits[parent].SHA == "" {
			roots = append(roots, id)
		}
	}
	if len(stack.Order) == 0 {
		sort.Strings(roots)
		return roots
	}
	index := make(map[string]int, len(stack.Order))
	for i, id := range stack.Order {
		index[id] = i
	}
	sort.Slice(roots, func(i, j int) bool {
		return index[roots[i]] < index[roots[j]]
	})
	return roots
}

func pickParentID(parents []string, idsBySHA map[string]string) string {
	for _, parentSHA := range parents {
		if id := idsBySHA[parentSHA]; id != "" {
			return id
		}
	}
	return ""
}
