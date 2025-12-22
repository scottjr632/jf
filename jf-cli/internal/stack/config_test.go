package stack

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestLoadDefaultsWhenMissing(t *testing.T) {
	originalRun := runGit
	originalRead := readFile
	defer func() {
		runGit = originalRun
		readFile = originalRead
	}()

	runGit = func(ctx context.Context, repo string, args ...string) (string, error) {
		return "/repo\n", nil
	}
	readFile = func(path string) ([]byte, error) {
		if path != "/repo/.jf/stack.json" {
			t.Fatalf("unexpected config path %q", path)
		}
		return nil, os.ErrNotExist
	}

	cfg, err := Load(context.Background(), "/repo")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.Trunk != defaultTrunk {
		t.Fatalf("expected trunk %q, got %q", defaultTrunk, cfg.Trunk)
	}
	if cfg.Remote != defaultRemote {
		t.Fatalf("expected remote %q, got %q", defaultRemote, cfg.Remote)
	}
	if cfg.BranchPrefix != defaultBranchPrefix {
		t.Fatalf("expected prefix %q, got %q", defaultBranchPrefix, cfg.BranchPrefix)
	}
	if cfg.Stacks == nil {
		t.Fatalf("expected stacks map to be initialized")
	}
}

func TestSaveWritesConfig(t *testing.T) {
	originalRun := runGit
	originalWrite := writeFile
	originalMkdir := mkdirAll
	defer func() {
		runGit = originalRun
		writeFile = originalWrite
		mkdirAll = originalMkdir
	}()

	runGit = func(ctx context.Context, repo string, args ...string) (string, error) {
		return "/repo\n", nil
	}

	var gotPath string
	var gotData string
	writeFile = func(path string, data []byte, _ os.FileMode) error {
		gotPath = path
		gotData = string(data)
		return nil
	}

	var mkdirPath string
	mkdirAll = func(path string, _ os.FileMode) error {
		mkdirPath = path
		return nil
	}

	cfg := Config{Trunk: "main", Remote: "origin", BranchPrefix: "jf/stack", Stacks: map[string]StackMeta{}}
	if err := Save(context.Background(), "/repo", cfg); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if gotPath != "/repo/.jf/stack.json" {
		t.Fatalf("expected config path %q, got %q", "/repo/.jf/stack.json", gotPath)
	}
	if mkdirPath != "/repo/.jf" {
		t.Fatalf("expected mkdir path %q, got %q", "/repo/.jf", mkdirPath)
	}
	if !strings.Contains(gotData, "\"trunk\": \"main\"") {
		t.Fatalf("expected trunk in config, got %s", gotData)
	}
}
