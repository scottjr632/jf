package stack

import "testing"

func TestFindStackCommitPrefersCurrentStack(t *testing.T) {
	cfg := Config{
		CurrentStack: "feature-b",
		Stacks: map[string]StackMeta{
			"feature-a": {
				Commits: map[string]CommitMeta{
					"1": {SHA: "sha-a"},
				},
			},
			"feature-b": {
				Commits: map[string]CommitMeta{
					"2": {SHA: "sha-b"},
				},
			},
		},
	}

	name, id, meta, ok := FindStackCommit(&cfg, "sha-b")
	if !ok {
		t.Fatalf("expected commit to be found")
	}
	if name != "feature-b" {
		t.Fatalf("expected feature-b, got %s", name)
	}
	if id != "2" {
		t.Fatalf("expected id 2, got %s", id)
	}
	if meta.SHA != "sha-b" {
		t.Fatalf("expected sha-b, got %s", meta.SHA)
	}
}

func TestFindStackCommitSearchesOtherStacks(t *testing.T) {
	cfg := Config{
		CurrentStack: "feature-b",
		Stacks: map[string]StackMeta{
			"feature-a": {
				Commits: map[string]CommitMeta{
					"1": {SHA: "sha-a"},
				},
			},
			"feature-b": {
				Commits: map[string]CommitMeta{
					"2": {SHA: "sha-b"},
				},
			},
		},
	}

	name, id, meta, ok := FindStackCommit(&cfg, "sha-a")
	if !ok {
		t.Fatalf("expected commit to be found")
	}
	if name != "feature-a" {
		t.Fatalf("expected feature-a, got %s", name)
	}
	if id != "1" {
		t.Fatalf("expected id 1, got %s", id)
	}
	if meta.SHA != "sha-a" {
		t.Fatalf("expected sha-a, got %s", meta.SHA)
	}
}

func TestFindStackCommitMissing(t *testing.T) {
	cfg := Config{
		Stacks: map[string]StackMeta{
			"feature-a": {
				Commits: map[string]CommitMeta{
					"1": {SHA: "sha-a"},
				},
			},
		},
	}

	_, _, _, ok := FindStackCommit(&cfg, "sha-missing")
	if ok {
		t.Fatalf("expected commit to be missing")
	}
}
