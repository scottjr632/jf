package cli

import (
	"reflect"
	"testing"
)

func TestParseCommitArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    commitArgs
		wantErr bool
	}{
		{
			name: "message",
			args: []string{"-m", "msg"},
			want: commitArgs{
				gitArgs: []string{"-m", "msg"},
			},
		},
		{
			name: "amend",
			args: []string{"--amend", "-m", "msg"},
			want: commitArgs{
				amend:   true,
				gitArgs: []string{"-m", "msg"},
			},
		},
		{
			name: "worktree",
			args: []string{"--worktree", "feature", "-m", "msg"},
			want: commitArgs{
				worktree: "feature",
				gitArgs:  []string{"-m", "msg"},
			},
		},
		{
			name: "worktree-equals-terminator",
			args: []string{"--worktree=feature", "--", "--amend"},
			want: commitArgs{
				worktree: "feature",
				gitArgs:  []string{"--amend"},
			},
		},
		{
			name: "terminator-only",
			args: []string{"--"},
			want: commitArgs{},
		},
		{
			name: "missing-worktree-value",
			args: []string{"--worktree"},
			want: commitArgs{
				promptWorktree: true,
			},
		},
	}

	for _, tc := range tests {
		got, err := parseCommitArgs(tc.args)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("case %s: expected error, got nil", tc.name)
			}
			continue
		}
		if err != nil {
			t.Fatalf("case %s: unexpected error: %v", tc.name, err)
		}
		if got.amend != tc.want.amend || got.worktree != tc.want.worktree || got.promptWorktree != tc.want.promptWorktree || !reflect.DeepEqual(got.gitArgs, tc.want.gitArgs) {
			t.Fatalf("case %s: expected %#v, got %#v", tc.name, tc.want, got)
		}
	}
}
