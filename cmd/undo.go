package cmd

import (
	"fmt"

	"github.com/scottjr632/sequoia/internal/cli"
	"github.com/scottjr632/sequoia/internal/engine"
	"github.com/scottjr632/sequoia/internal/git"
	"github.com/spf13/cobra"
)

var undoCommitCmd = &cobra.Command{
	Use:     "undo",
	Aliases: []string{"uncommit"},
	Short:   "Undo the most recent commit but keep your changes",
	RunE: func(cmd *cobra.Command, args []string) error {
		stack, err := engine.GetStackForCurrentBranch()
		if err != nil {
			return err
		}

		if stack.IsTrunk {
			return fmt.Errorf("cannot undo commit on trunk")
		}

		if err = cli.ExecuteCommandInTerminal("git", "reset", "--soft", "HEAD^"); err != nil {
			return err
		}

		sha, err := git.GetCurrentBranchCommitSha()
		if err != nil {
			return err
		}
		stack.Sha = sha

		if len(stack.Children) > 0 {
			if err = cli.ExecuteCommandInTerminal("git", "stash"); err != nil {
				return err
			}
			defer func(stack *engine.Stack) {
				git.CheckoutBranch(stack.Name)
				cli.ExecuteCommandInTerminal("git", "stash", "pop")
			}(stack)
		}

		if err = engine.RestackChildren(stack); err != nil {
			return err
		}

		return engine.Save()
	},
}
