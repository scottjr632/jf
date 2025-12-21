package cli

import (
	"errors"

	"github.com/scottjr632/jf-cli/internal/worktree"
	"github.com/spf13/cobra"
)

func newAddCmd(opts *rootOptions) *cobra.Command {
	var noCheckout bool

	cmd := &cobra.Command{
		Use:   "add <path|name> [ref]",
		Short: "Add a new worktree",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) < 1 || len(args) > 2 {
				return errors.New("expected <path|name> and optional [ref]")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			ref := ""
			if len(args) == 2 {
				ref = args[1]
			}
			resolvedPath, err := worktree.Add(cmd.Context(), opts.repo, path, ref)
			if err != nil {
				return err
			}
			if noCheckout {
				return nil
			}
			return enterShell(cmd.Context(), resolvedPath, "")
		},
	}

	cmd.Flags().BoolVar(&noCheckout, "no-checkout", false, "Do not open a shell in the new worktree")

	return cmd
}
