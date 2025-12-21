package cli

import (
	"github.com/scottjr632/jf-cli/internal/worktree"
	"github.com/spf13/cobra"
)

func newMainCmd(opts *rootOptions) *cobra.Command {
	var shell string

	cmd := &cobra.Command{
		Use:   "main",
		Short: "Open the main worktree in a subshell",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := worktree.PathForBranches(cmd.Context(), opts.repo, []string{"main", "master"})
			if err != nil {
				return err
			}
			if err := ensureDir(path); err != nil {
				return err
			}
			return enterShell(cmd.Context(), path, shell)
		},
	}

	cmd.Flags().StringVar(&shell, "shell", "", "Shell to launch (defaults to $SHELL)")

	return cmd
}
