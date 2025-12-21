package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/scottjr632/jf-cli/internal/worktree"
	"github.com/spf13/cobra"
)

func newCheckoutCmd(opts *rootOptions) *cobra.Command {
	var shell string

	cmd := &cobra.Command{
		Use:     "checkout [path|name]",
		Short:   "Open a worktree in a subshell",
		Aliases: []string{"co"},
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("expected [path|name]")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var path string
			if len(args) == 0 {
				selected, err := promptWorktreeSelection(cmd.Context(), opts.repo)
				if err != nil {
					return err
				}
				path = selected
			} else {
				resolved, err := worktree.ResolvePath(cmd.Context(), opts.repo, args[0])
				if err != nil {
					return err
				}
				path = resolved
			}
			if err := ensureDir(path); err != nil {
				return err
			}
			return enterShell(cmd.Context(), path, shell)
		},
	}

	cmd.Flags().StringVar(&shell, "shell", "", "Shell to launch (defaults to $SHELL)")

	return cmd
}

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

func enterShell(ctx context.Context, dir, shell string) error {
	chosenShell := shell
	if chosenShell == "" {
		chosenShell = os.Getenv("SHELL")
	}
	if chosenShell == "" {
		chosenShell = "sh"
	}

	cmd := exec.CommandContext(ctx, chosenShell)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("launch shell %q: %w", chosenShell, err)
	}
	return nil
}

func ensureDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("worktree %q: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("worktree %q is not a directory", path)
	}
	return nil
}
