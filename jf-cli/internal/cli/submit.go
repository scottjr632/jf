package cli

import (
	"fmt"
	"os"

	"github.com/scottjr632/jf-cli/internal/stack"
	"github.com/spf13/cobra"
)

func newSubmitCmd(opts *rootOptions) *cobra.Command {
	var trunk string

	cmd := &cobra.Command{
		Use:   "submit",
		Short: "Create or update PRs for the current stack",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := stack.Load(cmd.Context(), opts.repo)
			if err != nil {
				return err
			}

			results, err := stack.SubmitCurrent(cmd.Context(), opts.repo, cfg, stack.SubmitOptions{Trunk: trunk})
			if err != nil {
				return err
			}
			for _, result := range results {
				suffix := ""
				if result.Number > 0 {
					suffix = fmt.Sprintf(" (#%d)", result.Number)
				}
				fmt.Fprintf(os.Stdout, "%s %s -> %s%s\n", result.Action, result.Commit.Short, result.Base, suffix)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&trunk, "trunk", "", "Override the trunk branch")

	return cmd
}
