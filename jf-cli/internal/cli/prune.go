package cli

import (
	"github.com/scottjr632/jf-cli/internal/worktree"
	"github.com/spf13/cobra"
)

func newPruneCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Prune stale worktree metadata",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return worktree.Prune(cmd.Context(), opts.repo)
		},
	}

	return cmd
}
