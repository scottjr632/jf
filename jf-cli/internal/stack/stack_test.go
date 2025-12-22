package stack

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestCurrentStackParsesCommits(t *testing.T) {
	originalRun := runGit
	originalWrite := writeFile
	originalMkdir := mkdirAll
	defer func() { runGit = originalRun }()
	defer func() {
		writeFile = originalWrite
		mkdirAll = originalMkdir
	}()

	format := "%H%x1f%P%x1f%h%x1f%s%x1f%b%x1e"
	runGit = func(ctx context.Context, repo string, args ...string) (string, error) {
		joined := strings.Join(args, " ")
		switch joined {
		case "rev-parse --verify refs/heads/main":
			return "", nil
		case "rev-parse --abbrev-ref HEAD":
			return "feature\n", nil
		case "rev-parse --show-toplevel":
			return "/repo\n", nil
		case "rev-parse HEAD":
			return "def456\n", nil
		case "rev-parse --verify ORIG_HEAD":
			return "", errors.New("missing")
		case "merge-base --is-ancestor main feature":
			return "", nil
		case "log --reverse --topo-order --format=" + format + " main..feature":
			return "abc123\x1ftrunksha\x1fabc123\x1fFirst\x1fBody\x1e" +
				"def456\x1fabc123\x1fdef456\x1fSecond\x1f\x1e", nil
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
	if stackInfo.Trunk != "main" || stackInfo.Head != "feature" {
		t.Fatalf("unexpected stack info: %#v", stackInfo)
	}
	if len(stackInfo.Commits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(stackInfo.Commits))
	}
	if stackInfo.Commits[0].Subject != "First" || stackInfo.Commits[1].Subject != "Second" {
		t.Fatalf("unexpected commits: %#v", stackInfo.Commits)
	}
}
