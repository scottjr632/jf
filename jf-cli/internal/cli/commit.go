package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/scottjr632/jf-cli/internal/git"
	"github.com/scottjr632/jf-cli/internal/stack"
	"github.com/scottjr632/jf-cli/internal/worktree"
	"github.com/spf13/cobra"
)

type commitArgs struct {
	amend          bool
	worktree       string
	promptWorktree bool
	gitArgs        []string
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

			gitArgs := []string{"commit"}
			if parsed.amend {
				gitArgs = append(gitArgs, "--amend")
			}
			gitArgs = append(gitArgs, parsed.gitArgs...)
			if err := git.RunPassthrough(cmd.Context(), repo, gitArgs...); err != nil {
				return err
			}
			cfg, err := stack.Load(cmd.Context(), repo)
			if err != nil {
				return err
			}
			return stack.RecordCommit(cmd.Context(), repo, &cfg, "")
		},
	}

	return cmd
}

func stageForCommit(ctx context.Context, repo string) error {
	if err := git.RunPassthrough(ctx, repo, "add", "-p"); err != nil {
		return err
	}

	untracked, err := listUntrackedFiles(ctx, repo)
	if err != nil {
		return err
	}
	if len(untracked) == 0 {
		return nil
	}

	reader := bufio.NewReader(os.Stdin)
	for _, path := range untracked {
		ok, err := promptYesNo(reader, fmt.Sprintf("Add untracked file %q? [y/N]: ", path))
		if err != nil {
			return err
		}
		if ok {
			if _, err := git.Run(ctx, repo, "add", "--", path); err != nil {
				return err
			}
		}
	}

	return nil
}

func listUntrackedFiles(ctx context.Context, repo string) ([]string, error) {
	out, err := git.Run(ctx, repo, "ls-files", "--others", "--exclude-standard", "-z")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}

	parts := strings.Split(out, "\x00")
	paths := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		paths = append(paths, part)
	}

	return paths, nil
}

func promptYesNo(reader *bufio.Reader, prompt string) (bool, error) {
	for {
		fmt.Fprint(os.Stdout, prompt)
		text, err := reader.ReadString('\n')
		if err != nil && len(text) == 0 {
			return false, err
		}

		trimmed := strings.TrimSpace(strings.ToLower(text))
		if trimmed == "y" || trimmed == "yes" {
			return true, nil
		}
		if trimmed == "" || trimmed == "n" || trimmed == "no" {
			return false, nil
		}
		if err == io.EOF {
			return false, err
		}
	}
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
