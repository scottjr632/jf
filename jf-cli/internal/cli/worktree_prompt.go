package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/scottjr632/jf-cli/internal/worktree"
)

func promptWorktreeSelection(ctx context.Context, repo string) (string, error) {
	entries, err := worktree.ListEntries(ctx, repo)
	if err != nil {
		return "", err
	}
	if len(entries) == 0 {
		return "", fmt.Errorf("no worktrees found")
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprintln(os.Stdout, "Select a worktree:")
		for i, entry := range entries {
			fmt.Fprintf(os.Stdout, "  %d) %s\n", i+1, formatWorktreeEntry(entry))
		}
		fmt.Fprint(os.Stdout, "Enter number: ")

		text, err := reader.ReadString('\n')
		if err != nil && len(text) == 0 {
			return "", err
		}
		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			if err == io.EOF {
				return "", err
			}
			continue
		}
		index, convErr := strconv.Atoi(trimmed)
		if convErr != nil || index < 1 || index > len(entries) {
			fmt.Fprintln(os.Stdout, "Invalid selection.")
			if err == io.EOF {
				return "", err
			}
			continue
		}
		return entries[index-1].Path, nil
	}
}

func formatWorktreeEntry(entry worktree.Entry) string {
	label := "detached"
	if entry.Branch != "" {
		label = entry.Branch
	}
	return fmt.Sprintf("%s (%s)", label, entry.Path)
}
