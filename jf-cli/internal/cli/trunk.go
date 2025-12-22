package cli

import (
	"fmt"
	"os"

	"github.com/scottjr632/jf-cli/internal/stack"
	"github.com/spf13/cobra"
)

func newTrunkCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trunk [branch]",
		Short: "Show or set the trunk branch",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := stack.Load(cmd.Context(), opts.repo)
			if err != nil {
				return err
			}
			if len(args) == 0 {
				fmt.Fprintln(os.Stdout, cfg.Trunk)
				return nil
			}
			if err := stack.SetTrunk(&cfg, args[0]); err != nil {
				return err
			}
			return stack.Save(cmd.Context(), opts.repo, cfg)
		},
	}

	return cmd
}
