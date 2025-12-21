package worktree

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestListUsesGitWorktreeList(t *testing.T) {
	original := runGit
	defer func() { runGit = original }()

	var gotArgs []string
	runGit = func(ctx context.Context, repo string, args ...string) (string, error) {
		gotArgs = append([]string{}, args...)
		return "ok", nil
	}

	out, err := List(context.Background(), "")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if out != "ok" {
		t.Fatalf("expected output to be ok, got %q", out)
	}
	want := []string{"worktree", "list"}
	if !reflect.DeepEqual(gotArgs, want) {
		t.Fatalf("expected args %v, got %v", want, gotArgs)
	}
}

func TestListEntriesParsesPorcelain(t *testing.T) {
	original := runGit
	defer func() { runGit = original }()

	runGit = func(ctx context.Context, repo string, args ...string) (string, error) {
		if !reflect.DeepEqual(args, []string{"worktree", "list", "--porcelain"}) {
			t.Fatalf("expected args %v, got %v", []string{"worktree", "list", "--porcelain"}, args)
		}
		return "worktree /tmp/main\nHEAD 123\nbranch refs/heads/main\n\nworktree /tmp/feature\nHEAD 456\ndetached\n", nil
	}

	entries, err := ListEntries(context.Background(), "/repo")
	if err != nil {
		t.Fatalf("ListEntries returned error: %v", err)
	}
	want := []Entry{
		{Path: "/tmp/main", Branch: "main"},
		{Path: "/tmp/feature", Detached: true},
	}
	if !reflect.DeepEqual(entries, want) {
		t.Fatalf("expected entries %v, got %v", want, entries)
	}
}

func TestAddUsesGitWorktreeAdd(t *testing.T) {
	original := runGit
	defer func() { runGit = original }()

	var gotArgs []string
	runGit = func(ctx context.Context, repo string, args ...string) (string, error) {
		gotArgs = append([]string{}, args...)
		return "", nil
	}

	gotPath, err := Add(context.Background(), "", "/tmp/wt", "main")
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if gotPath != "/tmp/wt" {
		t.Fatalf("expected path %q, got %q", "/tmp/wt", gotPath)
	}

	want := []string{"worktree", "add", "/tmp/wt", "main"}
	if !reflect.DeepEqual(gotArgs, want) {
		t.Fatalf("expected args %v, got %v", want, gotArgs)
	}
}

func TestAddResolvesNamedPath(t *testing.T) {
	originalRun := runGit
	originalHome := userHomeDir
	originalMkdir := mkdirAll
	defer func() {
		runGit = originalRun
		userHomeDir = originalHome
		mkdirAll = originalMkdir
	}()

	runGitCalls := 0
	runGit = func(ctx context.Context, repo string, args ...string) (string, error) {
		runGitCalls++
		if reflect.DeepEqual(args, []string{"rev-parse", "--show-toplevel"}) {
			return "/repos/jf\n", nil
		}
		want := []string{"worktree", "add", "/home/me/.jf/jf/worktrees/feature-x"}
		if !reflect.DeepEqual(args, want) {
			t.Fatalf("expected args %v, got %v", want, args)
		}
		return "", nil
	}
	userHomeDir = func() (string, error) {
		return "/home/me", nil
	}
	var mkdirPath string
	mkdirAll = func(path string, _ os.FileMode) error {
		mkdirPath = path
		return nil
	}

	gotPath, err := Add(context.Background(), "", "feature-x", "")
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if gotPath != "/home/me/.jf/jf/worktrees/feature-x" {
		t.Fatalf("expected path %q, got %q", "/home/me/.jf/jf/worktrees/feature-x", gotPath)
	}
	if runGitCalls != 2 {
		t.Fatalf("expected 2 git calls, got %d", runGitCalls)
	}
	if mkdirPath != "/home/me/.jf/jf/worktrees" {
		t.Fatalf("expected mkdir path %q, got %q", "/home/me/.jf/jf/worktrees", mkdirPath)
	}
}

func TestAddWithoutRef(t *testing.T) {
	original := runGit
	defer func() { runGit = original }()

	var gotArgs []string
	runGit = func(ctx context.Context, repo string, args ...string) (string, error) {
		gotArgs = append([]string{}, args...)
		return "", nil
	}

	gotPath, err := Add(context.Background(), "", "/tmp/wt", "")
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if gotPath != "/tmp/wt" {
		t.Fatalf("expected path %q, got %q", "/tmp/wt", gotPath)
	}

	want := []string{"worktree", "add", "/tmp/wt"}
	if !reflect.DeepEqual(gotArgs, want) {
		t.Fatalf("expected args %v, got %v", want, gotArgs)
	}
}

func TestRemoveUsesGitWorktreeRemove(t *testing.T) {
	original := runGit
	defer func() { runGit = original }()

	var gotArgs [][]string
	runGit = func(ctx context.Context, repo string, args ...string) (string, error) {
		gotArgs = append(gotArgs, append([]string{}, args...))
		if reflect.DeepEqual(args, []string{"worktree", "list", "--porcelain"}) {
			return "worktree /tmp/wt\nHEAD 123\nbranch refs/heads/feature-x\n", nil
		}
		return "", nil
	}

	if err := Remove(context.Background(), "", "/tmp/wt"); err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}

	want := [][]string{
		{"worktree", "list", "--porcelain"},
		{"worktree", "remove", "/tmp/wt"},
		{"branch", "-D", "feature-x"},
	}
	if !reflect.DeepEqual(gotArgs, want) {
		t.Fatalf("expected args %v, got %v", want, gotArgs)
	}
}

func TestRemoveSkipsBranchDeletionWhenDetached(t *testing.T) {
	original := runGit
	defer func() { runGit = original }()

	var gotArgs [][]string
	runGit = func(ctx context.Context, repo string, args ...string) (string, error) {
		gotArgs = append(gotArgs, append([]string{}, args...))
		if reflect.DeepEqual(args, []string{"worktree", "list", "--porcelain"}) {
			return "worktree /tmp/wt\nHEAD 123\ndetached\n", nil
		}
		return "", nil
	}

	if err := Remove(context.Background(), "", "/tmp/wt"); err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}

	want := [][]string{
		{"worktree", "list", "--porcelain"},
		{"worktree", "remove", "/tmp/wt"},
	}
	if !reflect.DeepEqual(gotArgs, want) {
		t.Fatalf("expected args %v, got %v", want, gotArgs)
	}
}

func TestPruneUsesGitWorktreePrune(t *testing.T) {
	original := runGit
	defer func() { runGit = original }()

	var gotArgs []string
	runGit = func(ctx context.Context, repo string, args ...string) (string, error) {
		gotArgs = append([]string{}, args...)
		return "", nil
	}

	if err := Prune(context.Background(), ""); err != nil {
		t.Fatalf("Prune returned error: %v", err)
	}

	want := []string{"worktree", "prune"}
	if !reflect.DeepEqual(gotArgs, want) {
		t.Fatalf("expected args %v, got %v", want, gotArgs)
	}
}

func TestMergeUsesTargetWorktree(t *testing.T) {
	original := runGit
	defer func() { runGit = original }()

	var gotCalls []struct {
		repo string
		args []string
	}
	runGit = func(ctx context.Context, repo string, args ...string) (string, error) {
		gotCalls = append(gotCalls, struct {
			repo string
			args []string
		}{
			repo: repo,
			args: append([]string{}, args...),
		})
		if reflect.DeepEqual(args, []string{"worktree", "list", "--porcelain"}) {
			return "worktree /tmp/main\nHEAD 123\nbranch refs/heads/main\n\nworktree /tmp/feature\nHEAD 456\nbranch refs/heads/feature\n", nil
		}
		return "", nil
	}

	if err := Merge(context.Background(), "/repo", "/tmp/feature", "main"); err != nil {
		t.Fatalf("Merge returned error: %v", err)
	}

	if len(gotCalls) != 3 {
		t.Fatalf("expected 3 git calls, got %d", len(gotCalls))
	}
	if !reflect.DeepEqual(gotCalls[0].args, []string{"worktree", "list", "--porcelain"}) {
		t.Fatalf("expected first call to list worktrees, got %v", gotCalls[0].args)
	}
	if !reflect.DeepEqual(gotCalls[1].args, []string{"worktree", "list", "--porcelain"}) {
		t.Fatalf("expected second call to list worktrees, got %v", gotCalls[1].args)
	}
	if gotCalls[2].repo != "/tmp/main" {
		t.Fatalf("expected merge to run in %q, got %q", "/tmp/main", gotCalls[2].repo)
	}
	if !reflect.DeepEqual(gotCalls[2].args, []string{"merge", "feature"}) {
		t.Fatalf("expected merge args %v, got %v", []string{"merge", "feature"}, gotCalls[2].args)
	}
}

func TestMergeFailsWhenDetached(t *testing.T) {
	original := runGit
	defer func() { runGit = original }()

	runGit = func(ctx context.Context, repo string, args ...string) (string, error) {
		if reflect.DeepEqual(args, []string{"worktree", "list", "--porcelain"}) {
			return "worktree /tmp/feature\nHEAD 456\ndetached\n", nil
		}
		return "", nil
	}

	if err := Merge(context.Background(), "/repo", "/tmp/feature", "main"); err == nil {
		t.Fatalf("expected merge to fail for detached worktree")
	}
}

func TestPathForBranchFindsWorktree(t *testing.T) {
	original := runGit
	defer func() { runGit = original }()

	var gotArgs []string
	runGit = func(ctx context.Context, repo string, args ...string) (string, error) {
		gotArgs = append([]string{}, args...)
		return "worktree /tmp/main\nHEAD 123\nbranch refs/heads/main\n", nil
	}

	gotPath, err := PathForBranch(context.Background(), "/repo", "main")
	if err != nil {
		t.Fatalf("PathForBranch returned error: %v", err)
	}
	if gotPath != "/tmp/main" {
		t.Fatalf("expected path %q, got %q", "/tmp/main", gotPath)
	}
	if !reflect.DeepEqual(gotArgs, []string{"worktree", "list", "--porcelain"}) {
		t.Fatalf("expected args %v, got %v", []string{"worktree", "list", "--porcelain"}, gotArgs)
	}
}

func TestPathForBranchesFallsBack(t *testing.T) {
	original := runGit
	defer func() { runGit = original }()

	runGit = func(ctx context.Context, repo string, args ...string) (string, error) {
		return "worktree /tmp/master\nHEAD 123\nbranch refs/heads/master\n", nil
	}

	gotPath, err := PathForBranches(context.Background(), "/repo", []string{"main", "master"})
	if err != nil {
		t.Fatalf("PathForBranches returned error: %v", err)
	}
	if gotPath != "/tmp/master" {
		t.Fatalf("expected path %q, got %q", "/tmp/master", gotPath)
	}
}

func TestAddPropagatesError(t *testing.T) {
	original := runGit
	defer func() { runGit = original }()

	expected := errors.New("boom")
	runGit = func(ctx context.Context, repo string, args ...string) (string, error) {
		return "", expected
	}

	gotPath, err := Add(context.Background(), "", "/tmp/wt", "")
	if !errors.Is(err, expected) {
		t.Fatalf("expected error to propagate, got %v", err)
	}
	if gotPath != "" {
		t.Fatalf("expected path to be empty on error, got %q", gotPath)
	}
}
