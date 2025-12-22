package cli

import (
	"fmt"
	"os"

	"github.com/scottjr632/jf-cli/internal/stack"
	"github.com/spf13/cobra"
)

func newLsCmd(opts *rootOptions) *cobra.Command {
	var trunk string

	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List the current stack commits",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := stack.Load(cmd.Context(), opts.repo)
			if err != nil {
				return err
			}
			if trunk == "" {
				trunk = cfg.Trunk
			}

			stackInfo, err := stack.CurrentStack(cmd.Context(), opts.repo, trunk)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "trunk: %s\n", stackInfo.Trunk)
			fmt.Fprintf(os.Stdout, "head: %s\n", stackInfo.Head)
			if len(stackInfo.Commits) == 0 {
				fmt.Fprintln(os.Stdout, "No commits in stack.")
				return nil
			}
			for i, commit := range stackInfo.Commits {
				fmt.Fprintf(os.Stdout, "  %d) %s %s\n", i+1, commit.Short, commit.Subject)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&trunk, "trunk", "", "Override the trunk branch")

	return cmd
}
