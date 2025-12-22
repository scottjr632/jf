package stack

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestSubmitCurrentUpdatesTitle(t *testing.T) {
	originalRunGit := runGit
	originalRunGh := runGh
	defer func() {
		runGit = originalRunGit
		runGh = originalRunGh
	}()

	format := "%H%x1f%h%x1f%s%x1f%b%x1e"

	runGit = func(ctx context.Context, repo string, args ...string) (string, error) {
		if reflect.DeepEqual(args, []string{"rev-parse", "--show-toplevel"}) {
			return "/repo\n", nil
		}
		if reflect.DeepEqual(args, []string{"rev-parse", "--verify", "refs/heads/main"}) {
			return "", nil
		}
		if reflect.DeepEqual(args, []string{"rev-parse", "--abbrev-ref", "HEAD"}) {
			return "feature\n", nil
		}
		if reflect.DeepEqual(args, []string{"merge-base", "--is-ancestor", "main", "HEAD"}) {
			return "", nil
		}
		if reflect.DeepEqual(args, []string{"log", "--reverse", "--format=" + format, "main..HEAD"}) {
			return "abc123\x1fabc123\x1fNew title\x1fBody\x1e", nil
		}
		if len(args) == 4 && args[0] == "branch" && args[1] == "-f" {
			return "", nil
		}
		if len(args) == 4 && args[0] == "push" && args[1] == "-f" {
			return "", nil
		}
		return "", errors.New("unexpected git call")
	}

	editCalled := false
	runGh = func(ctx context.Context, repo string, args ...string) (string, error) {
		if repo != "/repo" {
			return "", errors.New("unexpected repo")
		}
		if len(args) >= 2 && args[0] == "pr" && args[1] == "list" {
			return "[{\"number\":1,\"baseRefName\":\"main\",\"headRefName\":\"jf/stack/01-new-title-abc123\",\"title\":\"Old title\"}]", nil
		}
		if len(args) >= 2 && args[0] == "pr" && args[1] == "edit" {
			editCalled = true
			return "", nil
		}
		return "", errors.New("unexpected gh call")
	}

	cfg := DefaultConfig()
	results, err := SubmitCurrent(context.Background(), "/repo", cfg, SubmitOptions{})
	if err != nil {
		t.Fatalf("SubmitCurrent returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !editCalled {
		t.Fatalf("expected PR title update")
	}
	if results[0].Action != SubmitUpdated {
		t.Fatalf("expected updated action, got %s", results[0].Action)
	}
}
