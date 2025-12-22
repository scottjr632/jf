package stack

import (
	"context"
	"errors"
	"os"
	"testing"
)

func TestRecordAmendUpdatesCommitSHA(t *testing.T) {
	originalRun := runGit
	originalWrite := writeFile
	originalMkdir := mkdirAll
	defer func() {
		runGit = originalRun
		writeFile = originalWrite
		mkdirAll = originalMkdir
	}()

	runGit = func(ctx context.Context, repo string, args ...string) (string, error) {
		switch {
		case len(args) == 3 && args[0] == "rev-parse" && args[1] == "--abbrev-ref" && args[2] == "HEAD":
			return "feature\n", nil
		case len(args) == 2 && args[0] == "rev-parse" && args[1] == "HEAD":
			return "newsha\n", nil
		case len(args) == 3 && args[0] == "rev-parse" && args[1] == "--verify" && args[2] == "ORIG_HEAD":
			return "oldsha\n", nil
		case len(args) == 4 && args[0] == "log" && args[1] == "-1":
			return "newsha\x1fnewsha\x1fNew\x1fBody\x1e", nil
		case len(args) == 2 && args[0] == "rev-parse" && args[1] == "--show-toplevel":
			return "/repo\n", nil
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

	if err := RecordAmend(context.Background(), "/repo", &cfg, ""); err != nil {
		t.Fatalf("RecordAmend returned error: %v", err)
	}

	updated := cfg.Stacks["feature"].Commits["id-1"]
	if updated.SHA != "newsha" || updated.Subject != "New" {
		t.Fatalf("expected commit updated, got %#v", updated)
	}
}
