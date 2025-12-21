package cli

import (
	"errors"
	"strings"

	"github.com/scottjr632/jf-cli/internal/git"
	"github.com/scottjr632/jf-cli/internal/worktree"
	"github.com/spf13/cobra"
)

type commitArgs struct {
	amend    bool
	worktree string
	gitArgs  []string
}

func newCommitCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit [--amend] [--worktree <path|name>] [-- <git commit args...>]",
		Short: "Commit changes",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			parsed, err := parseCommitArgs(args)
			if err != nil {
				return err
			}

			repo := opts.repo
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

			gitArgs := []string{"commit"}
			if parsed.amend {
				gitArgs = append(gitArgs, "--amend")
			}
			gitArgs = append(gitArgs, parsed.gitArgs...)
			return git.RunPassthrough(cmd.Context(), repo, gitArgs...)
		},
	}

	return cmd
}

func parseCommitArgs(args []string) (commitArgs, error) {
	var parsed commitArgs

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			parsed.gitArgs = append(parsed.gitArgs, args[i+1:]...)
			break
		}
		if arg == "--amend" {
			parsed.amend = true
			continue
		}
		if arg == "--worktree" {
			if i+1 >= len(args) {
				return parsed, errors.New("expected value after --worktree")
			}
			parsed.worktree = args[i+1]
			i++
			continue
		}
		if strings.HasPrefix(arg, "--worktree=") {
			value := strings.TrimPrefix(arg, "--worktree=")
			if value == "" {
				return parsed, errors.New("expected value after --worktree")
			}
			parsed.worktree = value
			continue
		}
		parsed.gitArgs = append(parsed.gitArgs, arg)
	}

	return parsed, nil
}
