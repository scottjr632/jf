package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/scottjr632/jf-cli/internal/git"
	"github.com/scottjr632/jf-cli/internal/stack"
	"github.com/spf13/cobra"
)

func newGotoCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "goto [commit]",
		Short: "Checkout a tracked stack commit by hash",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := stack.Load(cmd.Context(), opts.repo)
			if err != nil {
				return err
			}

			var (
				ref       string
				selection *stackCommitOption
			)
			if len(args) == 0 {
				chosen, err := promptStackCommitSelection(cfg)
				if err != nil {
					return err
				}
				selection = &chosen
				ref = chosen.meta.SHA
			} else {
				ref = strings.TrimSpace(args[0])
				if ref == "" {
					return fmt.Errorf("commit hash is required")
				}
			}

			sha, err := resolveCommitSHA(cmd.Context(), opts.repo, ref)
			if err != nil {
				return err
			}

			var (
				stackName string
				commitID  string
				meta      stack.CommitMeta
			)
			if selection != nil {
				stackName = selection.stackName
				commitID = selection.commitID
				meta = selection.meta
				meta.SHA = sha
			} else {
				var ok bool
				stackName, commitID, meta, ok = stack.FindStackCommit(&cfg, sha)
				if !ok {
					return fmt.Errorf("commit %s not found in jf stacks", shortSHA(sha))
				}
			}

			cfg.CurrentStack = stackName
			if stackMeta, ok := cfg.Stacks[stackName]; ok {
				stackMeta.Current = commitID
				cfg.Stacks[stackName] = stackMeta
			}

			fmt.Fprintf(os.Stdout, "checkout %s %s\n", shortSHA(meta.SHA), meta.Subject)
			if err := git.RunPassthrough(cmd.Context(), opts.repo, "checkout", sha); err != nil {
				return err
			}
			return stack.SyncCurrent(cmd.Context(), opts.repo, &cfg, "")
		},
	}

	return cmd
}

func resolveCommitSHA(ctx context.Context, repo, ref string) (string, error) {
	out, err := git.Run(ctx, repo, "rev-parse", "--verify", ref+"^{commit}")
	if err != nil {
		return "", fmt.Errorf("resolve commit %q: %w", ref, err)
	}
	sha := strings.TrimSpace(out)
	if sha == "" {
		return "", fmt.Errorf("resolve commit %q: empty sha", ref)
	}
	return sha, nil
}

func shortSHA(sha string) string {
	sha = strings.TrimSpace(sha)
	if len(sha) <= 7 {
		return sha
	}
	return sha[:7]
}

type stackCommitOption struct {
	stackName string
	commitID  string
	meta      stack.CommitMeta
}

func promptStackCommitSelection(cfg stack.Config) (stackCommitOption, error) {
	options := collectStackCommitOptions(cfg)
	if len(options) == 0 {
		return stackCommitOption{}, fmt.Errorf("no tracked stack commits found")
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprintln(os.Stdout, "Select a commit:")
		for i, option := range options {
			fmt.Fprintf(os.Stdout, "  %d) %s\n", i+1, formatStackCommitOption(option))
		}
		fmt.Fprint(os.Stdout, "Enter number: ")

		text, err := reader.ReadString('\n')
		if err != nil && len(text) == 0 {
			return stackCommitOption{}, err
		}
		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			if err == io.EOF {
				return stackCommitOption{}, err
			}
			continue
		}
		index, convErr := strconv.Atoi(trimmed)
		if convErr != nil || index < 1 || index > len(options) {
			fmt.Fprintln(os.Stdout, "Invalid selection.")
			if err == io.EOF {
				return stackCommitOption{}, err
			}
			continue
		}
		return options[index-1], nil
	}
}

func collectStackCommitOptions(cfg stack.Config) []stackCommitOption {
	options := make([]stackCommitOption, 0, len(cfg.Stacks))

	if cfg.CurrentStack != "" {
		if stackMeta, ok := cfg.Stacks[cfg.CurrentStack]; ok {
			options = append(options, stackOptionsForName(cfg.CurrentStack, stackMeta)...)
		}
	}

	names := make([]string, 0, len(cfg.Stacks))
	for name := range cfg.Stacks {
		if name == cfg.CurrentStack {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		options = append(options, stackOptionsForName(name, cfg.Stacks[name])...)
	}

	return options
}

func stackOptionsForName(name string, stackMeta stack.StackMeta) []stackCommitOption {
	options := make([]stackCommitOption, 0, len(stackMeta.Order))
	for _, id := range stackMeta.Order {
		meta, ok := stackMeta.Commits[id]
		if !ok {
			continue
		}
		options = append(options, stackCommitOption{
			stackName: name,
			commitID:  id,
			meta:      meta,
		})
	}
	return options
}

func formatStackCommitOption(option stackCommitOption) string {
	subject := strings.TrimSpace(option.meta.Subject)
	if subject == "" {
		subject = "(no subject)"
	}
	return fmt.Sprintf("%s: %s %s", option.stackName, shortSHA(option.meta.SHA), subject)
}
