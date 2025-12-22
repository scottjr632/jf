package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/scottjr632/jf-cli/internal/git"
	"github.com/scottjr632/jf-cli/internal/stack"
	"github.com/scottjr632/jf-cli/internal/worktree"
	"github.com/spf13/cobra"
)

func newStatusCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status",
		Aliases: []string{"st"},
		Short:   "Show git status with worktree and stack info",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := git.RunPassthrough(cmd.Context(), opts.repo, "status"); err != nil {
				return err
			}

			info, err := currentWorktreeInfo(cmd.Context(), opts.repo)
			if err != nil {
				return err
			}

			fmt.Fprintln(os.Stdout)
			if info.path != "" {
				label := info.path
				if info.branch != "" {
					label = fmt.Sprintf("%s (%s)", info.path, info.branch)
				} else if info.detached {
					label = fmt.Sprintf("%s (detached)", info.path)
				}
				fmt.Fprintf(os.Stdout, "worktree: %s\n", label)
			}

			cfg, err := stack.Load(cmd.Context(), opts.repo)
			if err != nil {
				return err
			}
			status, err := stack.Status(cmd.Context(), opts.repo, &cfg, "")
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "stack: %s\n", status.Name)
			if status.CurrentSHA != "" {
				fmt.Fprintf(os.Stdout, "current: %s (%s)\n", status.CurrentID, status.CurrentShort)
			} else {
				fmt.Fprintln(os.Stdout, "current: none")
			}

			return nil
		},
	}

	return cmd
}

type worktreeInfo struct {
	path     string
	branch   string
	detached bool
}

func currentWorktreeInfo(ctx context.Context, repo string) (worktreeInfo, error) {
	root, err := git.Run(ctx, repo, "rev-parse", "--show-toplevel")
	if err != nil {
		return worktreeInfo{}, err
	}
	root = strings.TrimSpace(root)
	if root == "" {
		return worktreeInfo{}, nil
	}

	entries, err := worktree.ListEntries(ctx, repo)
	if err != nil {
		return worktreeInfo{path: root}, nil
	}
	for _, entry := range entries {
		if entry.Path == root {
			return worktreeInfo{path: entry.Path, branch: entry.Branch, detached: entry.Detached}, nil
		}
	}

	return worktreeInfo{path: root}, nil
}
