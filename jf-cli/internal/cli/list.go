package cli

import (
	"fmt"

	"github.com/scottjr632/jf-cli/internal/worktree"
	"github.com/spf13/cobra"
)

func newListCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List worktrees for the repo",
		RunE: func(cmd *cobra.Command, _ []string) error {
			output, err := worktree.List(cmd.Context(), opts.repo)
			if err != nil {
				return err
			}
			fmt.Print(output)
			return nil
		},
	}

	return cmd
}
