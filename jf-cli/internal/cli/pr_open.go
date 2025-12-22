package cli

import (
	"fmt"
	"os"

	"github.com/scottjr632/jf-cli/internal/stack"
	"github.com/spf13/cobra"
)

func newPrOpenCmd(opts *rootOptions) *cobra.Command {
	var trunk string

	cmd := &cobra.Command{
		Use:     "pr",
		Short:   "Open the current PR in a browser",
		Aliases: []string{"pr-open", "open-pr"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := stack.Load(cmd.Context(), opts.repo)
			if err != nil {
				return err
			}
			if trunk == "" {
				trunk = cfg.Trunk
			}

			status, err := stack.Status(cmd.Context(), opts.repo, &cfg, trunk)
			if err != nil {
				return err
			}
			if status.CurrentSHA == "" {
				fmt.Fprintln(os.Stdout, "No current commit tracked in stack.")
				return nil
			}

			stackInfo, err := stack.CurrentStackDetails(cmd.Context(), opts.repo, &cfg, trunk)
			if err != nil {
				return err
			}
			if len(stackInfo.Items) == 0 {
				fmt.Fprintln(os.Stdout, "No commits in stack.")
				return nil
			}

			currentIndex := -1
			var currentItem stack.StackItem
			for i, item := range stackInfo.Items {
				if item.ID == status.CurrentID {
					currentIndex = i
					currentItem = item
					break
				}
			}
			if currentIndex == -1 {
				fmt.Fprintln(os.Stdout, "Current commit not found in stack.")
				return nil
			}

			root, err := stack.RepoRoot(cmd.Context(), opts.repo)
			if err != nil {
				return err
			}
			branch := stack.BranchNameForCommit(cfg.BranchPrefix, currentIndex+1, currentItem.ID, currentItem.Commit)
			pr, err := stack.PRForBranch(cmd.Context(), root, branch)
			if err != nil {
				return err
			}
			if pr == nil {
				fmt.Fprintln(os.Stdout, "No PR found for current commit.")
				return nil
			}
			return stack.OpenPR(cmd.Context(), root, pr.Number)
		},
	}

	cmd.Flags().StringVar(&trunk, "trunk", "", "Override the trunk branch")

	return cmd
}
