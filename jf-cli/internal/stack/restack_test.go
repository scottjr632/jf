package stack

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestRestackNoopWhenAligned(t *testing.T) {
	originalRun := runGit
	originalRunPassthrough := runGitPassthrough
	originalWrite := writeFile
	originalMkdir := mkdirAll
	defer func() {
		runGit = originalRun
		runGitPassthrough = originalRunPassthrough
		writeFile = originalWrite
		mkdirAll = originalMkdir
	}()

	runGit = func(ctx context.Context, repo string, args ...string) (string, error) {
		joined := strings.Join(args, " ")
		switch joined {
		case "rev-parse --abbrev-ref HEAD":
			return "feature\n", nil
		case "rev-parse HEAD":
			return "sha2\n", nil
		case "rev-parse --verify ORIG_HEAD":
			return "", errors.New("missing")
		case "rev-parse --show-toplevel":
			return "/repo\n", nil
		case "rev-parse main":
			return "trunksha\n", nil
		case "rev-parse sha1^":
			return "trunksha\n", nil
		case "rev-parse sha2^":
			return "sha1\n", nil
		default:
			return "", errors.New("unexpected git call")
		}
	}

	runGitPassthrough = func(context.Context, string, ...string) error {
		return errors.New("unexpected rebase")
	}
	writeFile = func(string, []byte, os.FileMode) error { return nil }
	mkdirAll = func(string, os.FileMode) error { return nil }

	cfg := DefaultConfig()
	cfg.CurrentStack = "feature"
	cfg.Stacks["feature"] = StackMeta{
		Trunk: "main",
		Order: []string{"id-1", "id-2"},
		Commits: map[string]CommitMeta{
			"id-1": {SHA: "sha1", Subject: "First", Body: ""},
			"id-2": {SHA: "sha2", Subject: "Second", Body: ""},
		},
		Current: "id-2",
	}

	if err := Restack(context.Background(), "/repo", &cfg, ""); err != nil {
		t.Fatalf("Restack returned error: %v", err)
	}
}

func TestRestackRebasesWhenParentMismatch(t *testing.T) {
	originalRun := runGit
	originalRunPassthrough := runGitPassthrough
	originalWrite := writeFile
	originalMkdir := mkdirAll
	defer func() {
		runGit = originalRun
		runGitPassthrough = originalRunPassthrough
		writeFile = originalWrite
		mkdirAll = originalMkdir
	}()

	format := "%H%x1f%h%x1f%s%x1f%b%x1e"

	runGit = func(ctx context.Context, repo string, args ...string) (string, error) {
		joined := strings.Join(args, " ")
		switch joined {
		case "rev-parse --abbrev-ref HEAD":
			return "feature\n", nil
		case "rev-parse HEAD":
			return "newsha2\n", nil
		case "rev-parse --verify ORIG_HEAD":
			return "", errors.New("missing")
		case "rev-parse --show-toplevel":
			return "/repo\n", nil
		case "rev-parse main":
			return "trunksha\n", nil
		case "rev-parse sha1^":
			return "trunksha\n", nil
		case "rev-parse sha2^":
			return "trunksha\n", nil
		case "log --reverse --format=" + format + " main..feature":
			return "newsha1\x1fnewsha1\x1fFirst\x1f\x1e" +
				"newsha2\x1fnewsha2\x1fSecond\x1f\x1e", nil
		default:
			return "", errors.New("unexpected git call")
		}
	}

	rebaseCalled := false
	runGitPassthrough = func(ctx context.Context, repo string, args ...string) error {
		rebaseCalled = true
		if len(args) != 5 || args[0] != "rebase" || args[1] != "--onto" || args[2] != "sha1" || args[3] != "trunksha" || args[4] != "feature" {
			return errors.New("unexpected rebase args")
		}
		return nil
	}

	writeFile = func(string, []byte, os.FileMode) error { return nil }
	mkdirAll = func(string, os.FileMode) error { return nil }

	cfg := DefaultConfig()
	cfg.CurrentStack = "feature"
	cfg.Stacks["feature"] = StackMeta{
		Trunk: "main",
		Order: []string{"id-1", "id-2"},
		Commits: map[string]CommitMeta{
			"id-1": {SHA: "sha1", Subject: "First", Body: ""},
			"id-2": {SHA: "sha2", Subject: "Second", Body: ""},
		},
		Current: "id-2",
	}

	if err := Restack(context.Background(), "/repo", &cfg, ""); err != nil {
		t.Fatalf("Restack returned error: %v", err)
	}
	if !rebaseCalled {
		t.Fatalf("expected rebase to run")
	}
	if cfg.Stacks["feature"].Commits["id-1"].SHA != "newsha1" {
		t.Fatalf("expected updated first commit")
	}
	if cfg.Stacks["feature"].Commits["id-2"].SHA != "newsha2" {
		t.Fatalf("expected updated second commit")
	}
}
