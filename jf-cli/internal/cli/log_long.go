package cli

import (
	"fmt"
	"os"

	"github.com/scottjr632/jf-cli/internal/stack"
	"github.com/spf13/cobra"
)

func newLogLongCmd(opts *rootOptions) *cobra.Command {
	var trunk string

	cmd := &cobra.Command{
		Use:     "log-long",
		Short:   "List stack commits with PR status",
		Aliases: []string{"ll"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := stack.Load(cmd.Context(), opts.repo)
			if err != nil {
				return err
			}
			if trunk == "" {
				trunk = cfg.Trunk
			}

			stackInfo, err := stack.CurrentStack(cmd.Context(), opts.repo, &cfg, trunk)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "trunk: %s\n", stackInfo.Trunk)
			fmt.Fprintf(os.Stdout, "head: %s\n", stackInfo.Head)
			if len(stackInfo.Commits) == 0 {
				fmt.Fprintln(os.Stdout, "No commits in stack.")
				return nil
			}

			root, err := stack.RepoRoot(cmd.Context(), opts.repo)
			if err != nil {
				return err
			}
			for i, commit := range stackInfo.Commits {
				fmt.Fprintf(os.Stdout, "  %d) %s %s\n", i+1, commit.Short, commit.Subject)
				branch := stack.BranchNameForCommit(cfg.BranchPrefix, i+1, commit)
				pr, err := stack.PRForBranch(cmd.Context(), root, branch)
				if err != nil {
					return err
				}
				if pr == nil {
					fmt.Fprintln(os.Stdout, "     PR: none")
					continue
				}
				fmt.Fprintf(os.Stdout, "     PR: %s %s\n", pr.State, pr.URL)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&trunk, "trunk", "", "Override the trunk branch")

	return cmd
}
