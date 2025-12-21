package cli

import (
	"errors"

	"github.com/scottjr632/jf-cli/internal/worktree"
	"github.com/spf13/cobra"
)

func newMergeCmd(opts *rootOptions) *cobra.Command {
	var targetBranch string

	cmd := &cobra.Command{
		Use:   "merge <path|name>",
		Short: "Merge a worktree branch into another branch",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) > 1 {
				return errors.New("expected <path|name>")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				selection, err := promptWorktreeSelection(cmd.Context(), opts.repo)
				if err != nil {
					return err
				}
				args = []string{selection}
			}
			path, err := worktree.ResolvePath(cmd.Context(), opts.repo, args[0])
			if err != nil {
				return err
			}
			if err := ensureDir(path); err != nil {
				return err
			}
			return worktree.Merge(cmd.Context(), opts.repo, path, targetBranch)
		},
	}

	cmd.Flags().StringVar(&targetBranch, "into", "main", "Target branch to merge into")

	return cmd
}
