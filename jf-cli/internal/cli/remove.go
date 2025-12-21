package cli

import (
	"errors"

	"github.com/scottjr632/jf-cli/internal/worktree"
	"github.com/spf13/cobra"
)

func newRemoveCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <path|name>",
		Short: "Remove a worktree",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("expected <path|name>")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return worktree.Remove(cmd.Context(), opts.repo, args[0])
		},
	}

	return cmd
}
