package cli

import (
	"reflect"
	"testing"
)

func TestParseAmendArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    amendArgs
		wantErr bool
	}{
		{
			name: "default",
			args: []string{},
			want: amendArgs{},
		},
		{
			name: "edit",
			args: []string{"--edit"},
			want: amendArgs{
				edit: true,
			},
		},
		{
			name: "short-edit",
			args: []string{"-e", "-m", "msg"},
			want: amendArgs{
				edit:    true,
				gitArgs: []string{"-m", "msg"},
			},
		},
		{
			name: "worktree",
			args: []string{"--worktree", "feature", "-m", "msg"},
			want: amendArgs{
				worktree: "feature",
				gitArgs:  []string{"-m", "msg"},
			},
		},
		{
			name: "worktree-equals-terminator",
			args: []string{"--worktree=feature", "--", "--edit"},
			want: amendArgs{
				worktree: "feature",
				gitArgs:  []string{"--edit"},
			},
		},
		{
			name: "terminator-only",
			args: []string{"--"},
			want: amendArgs{},
		},
		{
			name: "missing-worktree-value",
			args: []string{"--worktree"},
			want: amendArgs{
				promptWorktree: true,
			},
		},
	}

	for _, tc := range tests {
		got, err := parseAmendArgs(tc.args)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("case %s: expected error, got nil", tc.name)
			}
			continue
		}
		if err != nil {
			t.Fatalf("case %s: unexpected error: %v", tc.name, err)
		}
		if got.edit != tc.want.edit || got.worktree != tc.want.worktree || got.promptWorktree != tc.want.promptWorktree || !reflect.DeepEqual(got.gitArgs, tc.want.gitArgs) {
			t.Fatalf("case %s: expected %#v, got %#v", tc.name, tc.want, got)
		}
	}
}

func TestShouldAddNoEdit(t *testing.T) {
	tests := []struct {
		name string
		edit bool
		args []string
		want bool
	}{
		{name: "default", edit: false, args: nil, want: true},
		{name: "edit-flag", edit: true, args: nil, want: false},
		{name: "edit-git-arg", edit: false, args: []string{"--edit"}, want: false},
		{name: "short-edit-git-arg", edit: false, args: []string{"-e"}, want: false},
		{name: "no-edit-git-arg", edit: false, args: []string{"--no-edit"}, want: false},
	}

	for _, tc := range tests {
		if got := shouldAddNoEdit(tc.edit, tc.args); got != tc.want {
			t.Fatalf("case %s: expected %v, got %v", tc.name, tc.want, got)
		}
	}
}
