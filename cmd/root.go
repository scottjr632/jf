package cmd

import (
	"fmt"

	"github.com/scottjr632/sequoia/internal/engine"
	"github.com/scottjr632/sequoia/internal/renderdag"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "jf",
	Short: "Make stacking easy with jiffy",
	RunE: func(cmd *cobra.Command, args []string) error {
		trunk, err := engine.GetTrunk()
		if err != nil {
			fmt.Println("failed to get trunk.", err)
			fmt.Println("jiffy is not initialized. Run `so init` to initialize jiffy.")
			return err
		}
		renderdag.RenderDag(trunk)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
	RootCmd.AddCommand(internalCmd)
	RootCmd.AddCommand(ghCmd)
	RootCmd.AddCommand(gitCmd)
	RootCmd.AddCommand(commitCmd)
	RootCmd.AddCommand(prevCmd)
	RootCmd.AddCommand(nextCmd)
	RootCmd.AddCommand(logCmd)
	RootCmd.AddCommand(logShortCmd)
	RootCmd.AddCommand(amendCmd)
	RootCmd.AddCommand(baseSubmitCmd)
	RootCmd.AddCommand(statusCmd)
	RootCmd.AddCommand(prCommand)
	RootCmd.AddCommand(trunkCmd)
	RootCmd.AddCommand(undoCommitCmd)
}
