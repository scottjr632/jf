package cli

import (
	"strings"

	"github.com/scottjr632/jf-cli/internal/git"
	"github.com/scottjr632/jf-cli/internal/stack"
	"github.com/scottjr632/jf-cli/internal/worktree"
	"github.com/spf13/cobra"
)

type amendArgs struct {
	edit           bool
	worktree       string
	promptWorktree bool
	gitArgs        []string
}

func newAmendCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "amend [--edit] [--worktree <path|name>] [-- <git commit args...>]",
		Short:   "Amend the latest commit",
		Aliases: []string{"am"},
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			parsed, err := parseAmendArgs(args)
			if err != nil {
				return err
			}

			repo := opts.repo
			if parsed.worktree == "" && parsed.promptWorktree {
				selection, err := promptWorktreeSelection(cmd.Context(), opts.repo)
				if err != nil {
					return err
				}
				parsed.worktree = selection
			}
			if parsed.worktree != "" {
				path, err := worktree.ResolvePath(cmd.Context(), opts.repo, parsed.worktree)
				if err != nil {
					return err
				}
				if err := ensureDir(path); err != nil {
					return err
				}
				repo = path
			}

			if err := stageForCommit(cmd.Context(), repo); err != nil {
				return err
			}

			gitArgs := []string{"commit", "--amend"}
			if shouldAddNoEdit(parsed.edit, parsed.gitArgs) {
				gitArgs = append(gitArgs, "--no-edit")
			}
			gitArgs = append(gitArgs, parsed.gitArgs...)
			if err := git.RunPassthrough(cmd.Context(), repo, gitArgs...); err != nil {
				return err
			}
			cfg, err := stack.Load(cmd.Context(), repo)
			if err != nil {
				return err
			}
			return stack.RecordAmend(cmd.Context(), repo, &cfg, "")
		},
	}

	return cmd
}

func parseAmendArgs(args []string) (amendArgs, error) {
	var parsed amendArgs

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			parsed.gitArgs = append(parsed.gitArgs, args[i+1:]...)
			break
		}
		if arg == "--edit" || arg == "-e" {
			parsed.edit = true
			continue
		}
		if arg == "--worktree" {
			if i+1 >= len(args) || args[i+1] == "--" || strings.HasPrefix(args[i+1], "-") {
				parsed.promptWorktree = true
				continue
			}
			parsed.worktree = args[i+1]
			i++
			continue
		}
		if after, ok := strings.CutPrefix(arg, "--worktree="); ok {
			value := after
			if value == "" {
				parsed.promptWorktree = true
				continue
			}
			parsed.worktree = value
			continue
		}
		parsed.gitArgs = append(parsed.gitArgs, arg)
	}

	return parsed, nil
}

func shouldAddNoEdit(edit bool, gitArgs []string) bool {
	if edit {
		return false
	}
	for _, arg := range gitArgs {
		if arg == "--edit" || arg == "-e" || arg == "--no-edit" {
			return false
		}
	}
	return true
}
