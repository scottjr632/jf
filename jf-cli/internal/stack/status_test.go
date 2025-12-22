package stack

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestStatusReturnsCurrent(t *testing.T) {
	originalRun := runGit
	originalWrite := writeFile
	originalMkdir := mkdirAll
	defer func() {
		runGit = originalRun
		writeFile = originalWrite
		mkdirAll = originalMkdir
	}()

	runGit = func(ctx context.Context, repo string, args ...string) (string, error) {
		joined := strings.Join(args, " ")
		switch joined {
		case "rev-parse --abbrev-ref HEAD":
			return "feature\n", nil
		case "rev-parse --show-toplevel":
			return "/repo\n", nil
		case "rev-parse HEAD":
			return "sha1\n", nil
		case "rev-parse --verify ORIG_HEAD":
			return "", errors.New("missing")
		case "rev-parse --verify refs/heads/main":
			return "", nil
		case "merge-base --is-ancestor main feature":
			return "", nil
		case "log --reverse --topo-order --format=%H%x1f%P%x1f%h%x1f%s%x1f%b%x1e main..feature":
			return "sha1\x1ftrunksha\x1fsha1\x1fFirst\x1f\x1e", nil
		default:
			return "", errors.New("unexpected git call")
		}
	}

	writeFile = func(string, []byte, os.FileMode) error { return nil }
	mkdirAll = func(string, os.FileMode) error { return nil }

	cfg := DefaultConfig()
	status, err := Status(context.Background(), "/repo", &cfg, "main")
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if status.Name != "feature" || status.Count != 1 {
		t.Fatalf("unexpected status: %#v", status)
	}
	if status.CurrentShort != "sha1" {
		t.Fatalf("expected current short sha, got %q", status.CurrentShort)
	}
}
