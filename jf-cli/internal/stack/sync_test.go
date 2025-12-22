package stack

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestSyncStackRefreshesCommits(t *testing.T) {
	originalRun := runGit
	originalWrite := writeFile
	originalMkdir := mkdirAll
	defer func() {
		runGit = originalRun
		writeFile = originalWrite
		mkdirAll = originalMkdir
	}()

	format := "%H%x1f%h%x1f%s%x1f%b%x1e"

	runGit = func(ctx context.Context, repo string, args ...string) (string, error) {
		joined := strings.Join(args, " ")
		switch joined {
		case "rev-parse --abbrev-ref HEAD":
			return "feature\n", nil
		case "rev-parse --show-toplevel":
			return "/repo\n", nil
		case "rev-parse --verify refs/heads/main":
			return "", nil
		case "merge-base --is-ancestor main feature":
			return "", nil
		case "log --reverse --format=" + format + " main..feature":
			return "newsha\x1fnewsha\x1fNew\x1f\x1e", nil
		case "rev-parse HEAD":
			return "newsha\n", nil
		case "rev-parse --verify ORIG_HEAD":
			return "", errors.New("missing")
		default:
			return "", errors.New("unexpected git call")
		}
	}

	writeFile = func(string, []byte, os.FileMode) error { return nil }
	mkdirAll = func(string, os.FileMode) error { return nil }

	cfg := DefaultConfig()
	cfg.CurrentStack = "feature"
	cfg.Stacks["feature"] = StackMeta{
		Trunk:   "main",
		Order:   []string{"id-1"},
		Commits: map[string]CommitMeta{"id-1": {SHA: "oldsha", Subject: "Old", Body: ""}},
		Current: "id-1",
	}

	if err := SyncStack(context.Background(), "/repo", &cfg, ""); err != nil {
		t.Fatalf("SyncStack returned error: %v", err)
	}

	updated := cfg.Stacks["feature"].Commits["id-1"]
	if updated.SHA != "newsha" || updated.Subject != "New" {
		t.Fatalf("expected commit updated, got %#v", updated)
	}
}
