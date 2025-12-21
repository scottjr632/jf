package cli

import (
	"github.com/scottjr632/jf-cli/internal/git"
	"github.com/spf13/cobra"
)

func newGitCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git <args...>",
		Short: "Run git commands using jf repo settings",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return git.RunPassthrough(cmd.Context(), opts.repo, args...)
		},
	}

	return cmd
}
