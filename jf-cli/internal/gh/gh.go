package gh

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Run executes gh with optional repo directory context and returns stdout.
func Run(ctx context.Context, repo string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "gh", args...)
	if repo != "" {
		cmd.Dir = repo
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("gh %s: %s", strings.Join(args, " "), msg)
	}

	return stdout.String(), nil
}
