package cli

import (
	"sort"
	"strings"

	"github.com/scottjr632/jf-cli/internal/stack"
)

type stackTreeEntry struct {
	Item     stack.StackItem
	Prefix   string
	Indent   string
	Position int
}

func buildStackTree(items []stack.StackItem) []stackTreeEntry {
	if len(items) == 0 {
		return nil
	}
	orderIndex := make(map[string]int, len(items))
	for i, item := range items {
		orderIndex[item.ID] = i
	}

	children := make(map[string][]string, len(items))
	roots := make([]string, 0, len(items))
	for _, item := range items {
		parent := strings.TrimSpace(item.ParentID)
		if parent == "" {
			roots = append(roots, item.ID)
			continue
		}
		if _, ok := orderIndex[parent]; !ok {
			roots = append(roots, item.ID)
			continue
		}
		children[parent] = append(children[parent], item.ID)
	}
	sort.Slice(roots, func(i, j int) bool {
		return orderIndex[roots[i]] < orderIndex[roots[j]]
	})
	for parent, ids := range children {
		sort.Slice(ids, func(i, j int) bool {
			return orderIndex[ids[i]] < orderIndex[ids[j]]
		})
		children[parent] = ids
	}

	byID := make(map[string]stack.StackItem, len(items))
	for _, item := range items {
		byID[item.ID] = item
	}

	var entries []stackTreeEntry
	var visit func(id, prefix string, isLast, isRoot bool)
	visit = func(id, prefix string, isLast, isRoot bool) {
		item := byID[id]
		linePrefix := ""
		nextPrefix := prefix
		if !isRoot {
			if isLast {
				linePrefix = prefix + "`-- "
				nextPrefix = prefix + "    "
			} else {
				linePrefix = prefix + "|-- "
				nextPrefix = prefix + "|   "
			}
		}
		position := orderIndex[id] + 1
		indent := strings.Repeat(" ", len(linePrefix)) + "   "
		entries = append(entries, stackTreeEntry{
			Item:     item,
			Prefix:   linePrefix,
			Indent:   indent,
			Position: position,
		})
		childIDs := children[id]
		for i, childID := range childIDs {
			visit(childID, nextPrefix, i == len(childIDs)-1, false)
		}
	}
	for i, root := range roots {
		visit(root, "", i == len(roots)-1, true)
	}
	return entries
}
