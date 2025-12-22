package cli

import (
	"github.com/scottjr632/jf-cli/internal/stack"
	"github.com/spf13/cobra"
)

func newRestackCmd(opts *rootOptions) *cobra.Command {
	var trunk string

	cmd := &cobra.Command{
		Use:   "restack",
		Short: "Restack commits on top of trunk",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := stack.Load(cmd.Context(), opts.repo)
			if err != nil {
				return err
			}
			return stack.Restack(cmd.Context(), opts.repo, &cfg, trunk)
		},
	}

	cmd.Flags().StringVar(&trunk, "trunk", "", "Override the trunk branch")

	return cmd
}
