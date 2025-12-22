package cli

import (
	"fmt"
	"os"

	"github.com/scottjr632/jf-cli/internal/stack"
	"github.com/spf13/cobra"
)

func newStackCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stack",
		Short: "Stack metadata commands",
	}

	cmd.AddCommand(newStackStatusCmd(opts))

	return cmd
}

func newStackStatusCmd(opts *rootOptions) *cobra.Command {
	var trunk string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current stack status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := stack.Load(cmd.Context(), opts.repo)
			if err != nil {
				return err
			}
			status, err := stack.Status(cmd.Context(), opts.repo, &cfg, trunk)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "stack: %s\n", status.Name)
			fmt.Fprintf(os.Stdout, "trunk: %s\n", status.Trunk)
			fmt.Fprintf(os.Stdout, "head: %s\n", status.Head)
			fmt.Fprintf(os.Stdout, "count: %d\n", status.Count)
			if status.CurrentSHA != "" {
				fmt.Fprintf(os.Stdout, "current: %s (%s)\n", status.CurrentID, status.CurrentShort)
			} else {
				fmt.Fprintln(os.Stdout, "current: none")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&trunk, "trunk", "", "Override the trunk branch")

	return cmd
}
