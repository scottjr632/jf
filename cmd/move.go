package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/scottjr632/sequoia/internal/engine"
	"github.com/scottjr632/sequoia/internal/git"
	"github.com/scottjr632/sequoia/utils/branches"
	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move [destination]",
	Short: "Move the current branch changes onto another stacked branch",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var destination string
		var err error

		if len(args) > 0 {
			destination = args[0]
		} else {
			destination, err = branches.PromptForBranchesAndReturnSelection(false)
			if err != nil {
				return err
			}
		}

		if destination == "" {
			return fmt.Errorf("a destination branch is required")
		}

		return moveChanges(destination)
	},
}

func moveChanges(destination string) error {
	sourceStack, err := engine.GetStackForCurrentBranch()
	if err != nil {
		return err
	}

	if sourceStack.IsTrunk {
		return fmt.Errorf("cannot move changes from trunk")
	}

	if len(sourceStack.Children) > 0 {
		return fmt.Errorf("cannot move changes from a branch with stacked children")
	}

	destStack, err := engine.GetStackForBranch(destination)
	if err != nil {
		return err
	}

	if destStack == nil {
		return fmt.Errorf("destination branch %s not found in stack", destination)
	}

	if destStack.IsTrunk {
		return fmt.Errorf("cannot move changes onto trunk")
	}

	if destStack.ID == sourceStack.ID {
		return fmt.Errorf("destination branch must be different from the current branch")
	}

	currentBranch, err := git.GetCurrentBranchName()
	if err != nil {
		return err
	}

	if currentBranch != sourceStack.Name {
		return fmt.Errorf("must be on the source branch to move changes")
	}

	sourceSha, err := git.GetCurrentBranchCommitSha()
	if err != nil {
		return err
	}

	if _, err = git.CheckoutBranch(destStack.Name); err != nil {
		return err
	}

	destSha, err := git.GetCurrentBranchCommitSha()
	if err != nil {
		return err
	}

	if err = git.CherryPick(sourceSha); err != nil {
		color.Red("failed to cherry-pick changes from %s onto %s", sourceStack.Name, destStack.Name)
		return err
	}

	if err = git.ResetSoft("HEAD~1"); err != nil {
		return err
	}

	if err = git.ResetMixed("HEAD"); err != nil {
		return err
	}

	if err = git.PromptToPatch(); err != nil {
		_ = git.ResetHard("HEAD")
		if _, checkoutErr := git.CheckoutBranch(sourceStack.Name); checkoutErr != nil {
			return checkoutErr
		}
		return err
	}

	if err = git.EnsureStagedFiles(); err != nil {
		if git.IsNoStagedFilesError(err) {
			color.Yellow("no changes to move from %s", sourceStack.Name)
			if resetErr := git.ResetHard("HEAD"); resetErr != nil {
				return resetErr
			}
			if _, checkoutErr := git.CheckoutBranch(sourceStack.Name); checkoutErr != nil {
				return checkoutErr
			}
			return nil
		}
		return err
	}

	destStack.AddRevision(destSha)

	if err = git.AmendCommit(); err != nil {
		destStack.PopRevision()
		return err
	}

	if err = git.ResetHard("HEAD"); err != nil {
		return err
	}

	if err = engine.RestackChildren(destStack); err != nil {
		destStack.PopRevision()
		return err
	}

	if err = engine.Save(); err != nil {
		return err
	}

	if err = engine.RemoveBranchFromStack(sourceStack.Name); err != nil {
		return err
	}

	if err = git.DeleteBranchForce(sourceStack.Name); err != nil {
		return err
	}

	color.Green("Moved changes from %s to %s", sourceStack.Name, destStack.Name)

	return nil
}

func init() {
	RootCmd.AddCommand(moveCmd)
}
