package stack

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestSubmitCurrentCreatesAndUpdates(t *testing.T) {
	originalRunGit := runGit
	originalRunGh := runGh
	originalWrite := writeFile
	originalMkdir := mkdirAll
	defer func() {
		runGit = originalRunGit
		runGh = originalRunGh
		writeFile = originalWrite
		mkdirAll = originalMkdir
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
		if reflect.DeepEqual(args, []string{"rev-parse", "HEAD"}) {
			return "def456\n", nil
		}
		if reflect.DeepEqual(args, []string{"rev-parse", "--verify", "ORIG_HEAD"}) {
			return "", errors.New("missing")
		}
		if reflect.DeepEqual(args, []string{"merge-base", "--is-ancestor", "main", "feature"}) {
			return "", nil
		}
		if reflect.DeepEqual(args, []string{"log", "--reverse", "--format=" + format, "main..feature"}) {
			return "abc123\x1fabc123\x1fFirst\x1fBody\x1e" +
				"def456\x1fdef456\x1fSecond\x1f\x1e", nil
		}
		if len(args) == 4 && args[0] == "branch" && args[1] == "-f" {
			return "", nil
		}
		if len(args) == 4 && args[0] == "push" && args[1] == "-f" {
			return "", nil
		}
		return "", errors.New("unexpected git call")
	}

	listCounts := map[string]int{}
	runGh = func(ctx context.Context, repo string, args ...string) (string, error) {
		if repo != "/repo" {
			return "", errors.New("unexpected repo")
		}
		if len(args) >= 2 && args[0] == "repo" && args[1] == "view" {
			return "owner/repo\n", nil
		}
		if len(args) >= 2 && args[0] == "pr" && args[1] == "list" {
			branch := args[3]
			listCounts[branch]++
			switch branch {
			case "jf/stack/01-first-abc123":
				if listCounts[branch] == 1 {
					return "[]", nil
				}
				return "[{\"number\":1,\"baseRefName\":\"main\",\"headRefName\":\"jf/stack/01-first-abc123\",\"title\":\"First\"}]", nil
			case "jf/stack/02-second-def456":
				return "[{\"number\":2,\"baseRefName\":\"wrong\",\"headRefName\":\"jf/stack/02-second-def456\",\"title\":\"Second\"}]", nil
			default:
				return "[]", nil
			}
		}
		if len(args) >= 2 && args[0] == "pr" && args[1] == "create" {
			return "", nil
		}
		if len(args) >= 2 && args[0] == "pr" && args[1] == "comment" {
			return "", nil
		}
		if len(args) >= 1 && args[0] == "api" {
			return "[]", nil
		}
		if len(args) >= 2 && args[0] == "pr" && args[1] == "edit" {
			return "", nil
		}
		return "", errors.New("unexpected gh call")
	}

	writeFile = func(string, []byte, os.FileMode) error { return nil }
	mkdirAll = func(string, os.FileMode) error { return nil }

	cfg := DefaultConfig()
	results, err := SubmitCurrent(context.Background(), "/repo", cfg, SubmitOptions{})
	if err != nil {
		t.Fatalf("SubmitCurrent returned error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Action != SubmitCreated || !strings.Contains(results[0].Branch, "01-first") {
		t.Fatalf("unexpected first result: %#v", results[0])
	}
	if results[1].Action != SubmitUpdated || !strings.Contains(results[1].Branch, "02-second") {
		t.Fatalf("unexpected second result: %#v", results[1])
	}
}
