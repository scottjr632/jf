package stack

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestNextCommitMovesForward(t *testing.T) {
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
		case "rev-parse main":
			return "trunksha\n", nil
		case "merge-base --is-ancestor main feature":
			return "", nil
		case "log --reverse --topo-order --format=" + format + " main..feature":
			return "sha1\x1ftrunksha\x1fsha1\x1fFirst\x1f\x1e" +
				"sha2\x1fsha1\x1fsha2\x1fSecond\x1f\x1e", nil
		case "rev-parse HEAD":
			return "sha1\n", nil
		default:
			return "", errors.New("unexpected git call")
		}
	}

	writeFile = func(string, []byte, os.FileMode) error { return nil }
	mkdirAll = func(string, os.FileMode) error { return nil }

	cfg := DefaultConfig()
	commit, err := NextCommit(context.Background(), "/repo", &cfg, "main")
	if err != nil {
		t.Fatalf("NextCommit returned error: %v", err)
	}
	if commit.SHA != "sha2" {
		t.Fatalf("expected sha2, got %s", commit.SHA)
	}
}

func TestPrevCommitMovesBackward(t *testing.T) {
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
		case "rev-parse main":
			return "trunksha\n", nil
		case "merge-base --is-ancestor main feature":
			return "", nil
		case "log --reverse --topo-order --format=" + format + " main..feature":
			return "sha1\x1ftrunksha\x1fsha1\x1fFirst\x1f\x1e" +
				"sha2\x1fsha1\x1fsha2\x1fSecond\x1f\x1e", nil
		case "rev-parse HEAD":
			return "sha2\n", nil
		default:
			return "", errors.New("unexpected git call")
		}
	}

	writeFile = func(string, []byte, os.FileMode) error { return nil }
	mkdirAll = func(string, os.FileMode) error { return nil }

	cfg := DefaultConfig()
	commit, err := PrevCommit(context.Background(), "/repo", &cfg, "main")
	if err != nil {
		t.Fatalf("PrevCommit returned error: %v", err)
	}
	if commit.SHA != "sha1" {
		t.Fatalf("expected sha1, got %s", commit.SHA)
	}
}
