package cli

import (
	"fmt"
	"os"

	"github.com/scottjr632/jf-cli/internal/git"
	"github.com/scottjr632/jf-cli/internal/stack"
	"github.com/spf13/cobra"
)

func newPrevCmd(opts *rootOptions) *cobra.Command {
	var trunk string

	cmd := &cobra.Command{
		Use:   "prev",
		Short: "Checkout the previous commit in the current stack",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := stack.Load(cmd.Context(), opts.repo)
			if err != nil {
				return err
			}
			if trunk == "" {
				trunk = cfg.Trunk
			}

			target, err := stack.PrevCommit(cmd.Context(), opts.repo, &cfg, trunk)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "checkout %s %s\n", target.Short, target.Subject)
			if err := git.RunPassthrough(cmd.Context(), opts.repo, "checkout", target.SHA); err != nil {
				return err
			}
			return stack.SyncCurrent(cmd.Context(), opts.repo, &cfg, trunk)
		},
	}

	cmd.Flags().StringVar(&trunk, "trunk", "", "Override the trunk branch")

	return cmd
}
