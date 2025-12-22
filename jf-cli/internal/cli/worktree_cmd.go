package cli

import "github.com/spf13/cobra"

func newWorktreeCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "worktree",
		Short:   "Manage git worktrees",
		Aliases: []string{"w"},
	}

	cmd.AddCommand(
		newListCmd(opts),
		newNewCmd(opts),
		newCheckoutCmd(opts),
		newMainCmd(opts),
		newMergeCmd(opts),
		newCommitCmd(opts),
		newAmendCmd(opts),
		newRemoveCmd(opts),
		newPruneCmd(opts),
	)

	return cmd
}
