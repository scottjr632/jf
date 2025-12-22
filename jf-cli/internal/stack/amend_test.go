package stack

import (
	"context"
	"errors"
	"os"
	"testing"
)

func TestRecordAmendUpdatesCommitSHA(t *testing.T) {
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
	runGitPassthrough = func(context.Context, string, ...string) error { return nil }

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

func TestRecordAmendRebasesDescendants(t *testing.T) {
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
		switch {
		case len(args) == 3 && args[0] == "rev-parse" && args[1] == "--abbrev-ref" && args[2] == "HEAD":
			return "HEAD\n", nil
		case len(args) == 2 && args[0] == "rev-parse" && args[1] == "HEAD":
			return "newsha1\n", nil
		case len(args) == 3 && args[0] == "rev-parse" && args[1] == "--verify" && args[2] == "ORIG_HEAD":
			return "oldsha1\n", nil
		case len(args) == 4 && args[0] == "log" && args[1] == "-1":
			return "newsha1\x1fnewsha1\x1fNew\x1fBody\x1e", nil
		case len(args) == 2 && args[0] == "rev-parse" && args[1] == "--show-toplevel":
			return "/repo\n", nil
		case len(args) == 6 && args[0] == "branch" && args[1] == "--contains":
			return "", nil
		case len(args) == 4 && args[0] == "log" && args[1] == "--reverse" && args[2] == "--format="+format:
			return "newsha1\x1fnewsha1\x1fNew\x1fBody\x1e" +
				"childsha\x1fchildsha\x1fChild\x1f\x1e", nil
		default:
			return "", errors.New("unexpected git call")
		}
	}

	rebaseCalled := false
	runGitPassthrough = func(ctx context.Context, repo string, args ...string) error {
		rebaseCalled = true
		if len(args) != 5 || args[0] != "rebase" || args[1] != "--onto" || args[2] != "newsha1" || args[3] != "oldsha1" || args[4] != "HEAD" {
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
			"id-1": {SHA: "oldsha1", Subject: "Old", Body: ""},
			"id-2": {SHA: "childold", Subject: "Child", Body: ""},
		},
		Current: "id-1",
	}

	if err := RecordAmend(context.Background(), "/repo", &cfg, ""); err != nil {
		t.Fatalf("RecordAmend returned error: %v", err)
	}
	if !rebaseCalled {
		t.Fatalf("expected rebase to run for descendants")
	}
	if cfg.Stacks["feature"].Commits["id-1"].SHA != "newsha1" {
		t.Fatalf("expected amended commit to update")
	}
	if cfg.Stacks["feature"].Commits["id-2"].SHA != "childsha" {
		t.Fatalf("expected descendant commit to update")
	}
}
