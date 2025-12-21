package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Run executes git with optional repo path context and returns stdout.
func Run(ctx context.Context, repo string, args ...string) (string, error) {
	cmdArgs := make([]string, 0, len(args)+2)
	if repo != "" {
		info, err := os.Stat(repo)
		if err != nil {
			return "", fmt.Errorf("repo path %q: %w", repo, err)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("repo path %q is not a directory", repo)
		}
		cmdArgs = append(cmdArgs, "-C", repo)
	}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.CommandContext(ctx, "git", cmdArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}

	return stdout.String(), nil
}

// RunPassthrough executes git with optional repo path context and streams stdio.
func RunPassthrough(ctx context.Context, repo string, args ...string) error {
	cmdArgs := make([]string, 0, len(args)+2)
	if repo != "" {
		info, err := os.Stat(repo)
		if err != nil {
			return fmt.Errorf("repo path %q: %w", repo, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("repo path %q is not a directory", repo)
		}
		cmdArgs = append(cmdArgs, "-C", repo)
	}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.CommandContext(ctx, "git", cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
