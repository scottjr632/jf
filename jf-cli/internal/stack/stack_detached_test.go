package stack

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestCurrentStackDetachedHeadUsesContainingBranch(t *testing.T) {
	originalRun := runGit
	originalWrite := writeFile
	originalMkdir := mkdirAll
	defer func() { runGit = originalRun }()
	defer func() {
		writeFile = originalWrite
		mkdirAll = originalMkdir
	}()

	format := "%H%x1f%h%x1f%s%x1f%b%x1e"

	runGit = func(ctx context.Context, repo string, args ...string) (string, error) {
		joined := strings.Join(args, " ")
		switch joined {
		case "rev-parse --verify refs/heads/main":
			return "", nil
		case "rev-parse --abbrev-ref HEAD":
			return "HEAD\n", nil
		case "rev-parse --show-toplevel":
			return "/repo\n", nil
		case "rev-parse HEAD":
			return "abc123\n", nil
		case "rev-parse --verify ORIG_HEAD":
			return "", errors.New("missing")
		case "branch --contains HEAD --sort=-committerdate --format=%(refname:short)":
			return "feature\nmain\n", nil
		case "merge-base --is-ancestor main feature":
			return "", nil
		case "log --reverse --format=" + format + " main..feature":
			return "abc123\x1fabc123\x1fFirst\x1f\x1e", nil
		case "rev-parse --short HEAD":
			return "abc123\n", nil
		default:
			return "", errors.New("unexpected git call")
		}
	}

	writeFile = func(string, []byte, os.FileMode) error { return nil }
	mkdirAll = func(string, os.FileMode) error { return nil }

	cfg := DefaultConfig()
	stackInfo, err := CurrentStack(context.Background(), "/repo", &cfg, "main")
	if err != nil {
		t.Fatalf("CurrentStack returned error: %v", err)
	}
	if stackInfo.Head != "abc123" {
		t.Fatalf("expected short sha head, got %q", stackInfo.Head)
	}
	if len(stackInfo.Commits) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(stackInfo.Commits))
	}
}
