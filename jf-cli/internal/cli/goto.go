package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/scottjr632/jf-cli/internal/git"
	"github.com/scottjr632/jf-cli/internal/stack"
	"github.com/spf13/cobra"
)

func newGotoCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "goto <commit>",
		Short: "Checkout a tracked stack commit by hash",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := strings.TrimSpace(args[0])
			if ref == "" {
				return fmt.Errorf("commit hash is required")
			}

			sha, err := resolveCommitSHA(cmd.Context(), opts.repo, ref)
			if err != nil {
				return err
			}

			cfg, err := stack.Load(cmd.Context(), opts.repo)
			if err != nil {
				return err
			}

			stackName, commitID, meta, ok := stack.FindStackCommit(&cfg, sha)
			if !ok {
				return fmt.Errorf("commit %s not found in jf stacks", shortSHA(sha))
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
