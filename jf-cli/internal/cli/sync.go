package cli

import (
	"github.com/scottjr632/jf-cli/internal/stack"
	"github.com/spf13/cobra"
)

func newSyncCmd(opts *rootOptions) *cobra.Command {
	var trunk string

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync stack metadata with git history",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := stack.Load(cmd.Context(), opts.repo)
			if err != nil {
				return err
			}
			return stack.SyncStack(cmd.Context(), opts.repo, &cfg, trunk)
		},
	}

	cmd.Flags().StringVar(&trunk, "trunk", "", "Override the trunk branch")

	return cmd
}
