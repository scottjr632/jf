package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/scottjr632/jf-cli/internal/worktree"
	"github.com/spf13/cobra"
)

func newCheckoutCmd(opts *rootOptions) *cobra.Command {
	var shell string

	cmd := &cobra.Command{
		Use:     "checkout <path|name>",
		Short:   "Open a worktree in a subshell",
		Aliases: []string{"co"},
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("expected <path|name>")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := worktree.ResolvePath(cmd.Context(), opts.repo, args[0])
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

func enterShell(ctx context.Context, dir, shell string) error {
	chosenShell := shell
	if chosenShell == "" {
		chosenShell = os.Getenv("SHELL")
	}
	if chosenShell == "" {
		chosenShell = "sh"
	}

	cmd := exec.CommandContext(ctx, chosenShell)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("launch shell %q: %w", chosenShell, err)
	}
	return nil
}

func ensureDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("worktree %q: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("worktree %q is not a directory", path)
	}
	return nil
}
