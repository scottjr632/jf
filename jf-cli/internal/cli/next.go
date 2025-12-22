package cli

import (
	"fmt"
	"os"

	"github.com/scottjr632/jf-cli/internal/git"
	"github.com/scottjr632/jf-cli/internal/stack"
	"github.com/spf13/cobra"
)

func newNextCmd(opts *rootOptions) *cobra.Command {
	var trunk string

	cmd := &cobra.Command{
		Use:   "next",
		Short: "Checkout the next commit in the current stack",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := stack.Load(cmd.Context(), opts.repo)
			if err != nil {
				return err
			}
			if trunk == "" {
				trunk = cfg.Trunk
			}

			choices, err := stack.NextCommits(cmd.Context(), opts.repo, &cfg, trunk)
			if err != nil {
				return err
			}
			target := choices[0]
			if len(choices) > 1 {
				selection, err := promptNextCommitSelection(choices)
				if err != nil {
					return err
				}
				target = selection
			}
			fmt.Fprintf(os.Stdout, "checkout %s %s\n", target.Commit.Short, target.Commit.Subject)
			if err := git.RunPassthrough(cmd.Context(), opts.repo, "checkout", target.Commit.SHA); err != nil {
				return err
			}
			return stack.SyncCurrent(cmd.Context(), opts.repo, &cfg, trunk)
		},
	}

	cmd.Flags().StringVar(&trunk, "trunk", "", "Override the trunk branch")

	return cmd
}
