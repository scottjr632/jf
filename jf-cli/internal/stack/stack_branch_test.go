package stack

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestCurrentStackReusesCommitIDsFromCurrentStack(t *testing.T) {
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
		case "rev-parse --abbrev-ref HEAD":
			return "feature2\n", nil
		case "rev-parse --show-toplevel":
			return "/repo\n", nil
		case "rev-parse HEAD":
			return "shaD\n", nil
		case "rev-parse --verify ORIG_HEAD":
			return "", errors.New("missing")
		case "log --reverse --format=" + format + " main..feature2":
			return "shaA\x1fshaA\x1fFirst\x1f\x1e" +
				"shaD\x1fshaD\x1fSecond\x1f\x1e", nil
		default:
			return "", errors.New("unexpected git call")
		}
	}

	writeFile = func(string, []byte, os.FileMode) error { return nil }
	mkdirAll = func(string, os.FileMode) error { return nil }

	cfg := DefaultConfig()
	cfg.CurrentStack = "feature"
	cfg.Stacks["feature"] = StackMeta{
		Trunk: "main",
		Order: []string{"idA", "idB"},
		Commits: map[string]CommitMeta{
			"idA": {SHA: "shaA", Subject: "First"},
			"idB": {SHA: "shaB", Subject: "Second"},
		},
	}

	if _, err := CurrentStackDetails(context.Background(), "/repo", &cfg, "main"); err != nil {
		t.Fatalf("CurrentStackDetails returned error: %v", err)
	}

	branched, ok := cfg.Stacks["feature2"]
	if !ok {
		t.Fatalf("expected new stack for feature2")
	}
	if len(branched.Order) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(branched.Order))
	}
	if branched.Order[0] != "idA" {
		t.Fatalf("expected shared commit to reuse idA, got %q", branched.Order[0])
	}
	if branched.Order[1] == "" || branched.Order[1] == "idA" {
		t.Fatalf("expected new commit to have distinct id, got %q", branched.Order[1])
	}
}
