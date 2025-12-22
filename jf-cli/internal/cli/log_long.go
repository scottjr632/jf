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

			stackInfo, err := stack.CurrentStackDetails(cmd.Context(), opts.repo, &cfg, trunk)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "trunk: %s\n", stackInfo.Trunk)
			fmt.Fprintf(os.Stdout, "head: %s\n", stackInfo.Head)
			if len(stackInfo.Items) == 0 {
				fmt.Fprintln(os.Stdout, "No commits in stack.")
				return nil
			}

			root, err := stack.RepoRoot(cmd.Context(), opts.repo)
			if err != nil {
				return err
			}
			for _, entry := range buildStackTree(stackInfo.Items) {
				commit := entry.Item.Commit
				marker := " "
				if entry.Item.ID == stackInfo.CurrentID {
					marker = "*"
				}
				fmt.Fprintf(os.Stdout, "  %s%s%d) %s %s\n", entry.Prefix, marker, entry.Position, commit.Short, commit.Subject)
				branch := stack.BranchNameForCommit(cfg.BranchPrefix, entry.Position, entry.Item.ID, commit)
				pr, err := stack.PRForBranch(cmd.Context(), root, branch)
				if err != nil {
					return err
				}
				if pr == nil {
					fmt.Fprintf(os.Stdout, "%sPR: none\n", entry.Indent)
					continue
				}
				fmt.Fprintf(os.Stdout, "%sPR: %s %s\n", entry.Indent, pr.State, pr.URL)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&trunk, "trunk", "", "Override the trunk branch")

	return cmd
}
