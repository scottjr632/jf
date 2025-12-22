package cli

import "testing"

func TestShouldPassthrough(t *testing.T) {
	cmd := newRootCommand()

	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "empty", args: []string{}, want: false},
		{name: "list", args: []string{"list"}, want: false},
		{name: "main", args: []string{"main"}, want: false},
		{name: "ls", args: []string{"ls"}, want: false},
		{name: "log-long", args: []string{"log-long"}, want: false},
		{name: "ll", args: []string{"ll"}, want: false},
		{name: "next", args: []string{"next"}, want: false},
		{name: "prev", args: []string{"prev"}, want: false},
		{name: "commit", args: []string{"commit"}, want: false},
		{name: "amend", args: []string{"amend"}, want: false},
		{name: "pr", args: []string{"pr"}, want: false},
		{name: "trunk", args: []string{"trunk"}, want: false},
		{name: "stack", args: []string{"stack"}, want: false},
		{name: "sync", args: []string{"sync"}, want: false},
		{name: "restack", args: []string{"restack"}, want: false},
		{name: "submit", args: []string{"submit"}, want: false},
		{name: "help", args: []string{"help"}, want: false},
		{name: "help-flag", args: []string{"--help"}, want: false},
		{name: "short-help", args: []string{"-h"}, want: false},
		{name: "repo-list", args: []string{"-C", "repo", "list"}, want: false},
		{name: "status", args: []string{"status"}, want: true},
		{name: "git-subcommand", args: []string{"git", "status"}, want: false},
		{name: "repo-status", args: []string{"-C", "repo", "status"}, want: true},
		{name: "repo-equals-status", args: []string{"--repo=repo", "status"}, want: true},
		{name: "missing-repo-value", args: []string{"-C"}, want: false},
	}

	for _, tc := range tests {
		if got := shouldPassthrough(cmd, tc.args); got != tc.want {
			t.Fatalf("case %s: expected %v, got %v", tc.name, tc.want, got)
		}
	}
}
