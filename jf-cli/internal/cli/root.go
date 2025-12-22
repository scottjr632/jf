package cli

import (
	"context"
	"os"
	"strings"

	"github.com/scottjr632/jf-cli/internal/git"
	"github.com/spf13/cobra"
)

type rootOptions struct {
	repo string
}

func newRootCommand() *cobra.Command {
	opts := &rootOptions{}

	cmd := &cobra.Command{
		Use:   "jf",
		Short: "jf makes Git worktrees easy",
		Long:  "jf is a small helper CLI that streamlines common Git worktree workflows.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.repo, "repo", "C", "", "Path to the target Git repository (defaults to current dir)")

	cmd.AddCommand(
		newWorktreeCmd(opts),
		newStatusCmd(opts),
		newLsCmd(opts),
		newLogLongCmd(opts),
		newNextCmd(opts),
		newPrevCmd(opts),
		newPrOpenCmd(opts),
		newAmendCmd(opts),
		newTrunkCmd(opts),
		newStackCmd(opts),
		newSyncCmd(opts),
		newRestackCmd(opts),
		newSubmitCmd(opts),
		newGitCmd(opts),
	)

	return cmd
}

// Execute builds the root command and exits with the correct status code.
func Execute() {
	cmd := newRootCommand()
	cmd.InitDefaultCompletionCmd()
	args := os.Args[1:]
	if shouldPassthrough(cmd, args) {
		repo, gitArgs := splitRepoFlags(args)
		if err := git.RunPassthrough(context.Background(), repo, gitArgs...); err != nil {
			os.Exit(1)
		}
		return
	}

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func shouldPassthrough(cmd *cobra.Command, args []string) bool {
	_, gitArgs := splitRepoFlags(args)
	if len(gitArgs) == 0 {
		return false
	}
	for _, arg := range gitArgs {
		if arg == "-h" || arg == "--help" {
			return false
		}
	}
	name := firstNonFlagArg(gitArgs)
	if name == "" {
		return false
	}
	if name == "help" {
		return false
	}
	return !isJFCommand(cmd, name)
}

func splitRepoFlags(args []string) (string, []string) {
	repo := ""
	out := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-C" || arg == "--repo" {
			if i+1 < len(args) {
				repo = args[i+1]
				i++
				continue
			}
			out = append(out, arg)
			continue
		}
		if strings.HasPrefix(arg, "--repo=") {
			repo = strings.TrimPrefix(arg, "--repo=")
			continue
		}
		out = append(out, arg)
	}

	return repo, out
}

func firstNonFlagArg(args []string) string {
	if len(args) == 0 {
		return ""
	}
	seenTerminator := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			seenTerminator = true
			continue
		}
		if seenTerminator || !strings.HasPrefix(arg, "-") {
			return arg
		}
	}
	return ""
}

func isJFCommand(cmd *cobra.Command, name string) bool {
	for _, sub := range cmd.Commands() {
		if sub.Name() == name {
			return true
		}
		for _, alias := range sub.Aliases {
			if alias == name {
				return true
			}
		}
		if isJFCommand(sub, name) {
			return true
		}
	}
	return false
}
