package cli

import "testing"

func TestNewArgsValidation(t *testing.T) {
	cmd := newNewCmd(&rootOptions{})

	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "missing", args: []string{}, want: false},
		{name: "one", args: []string{"path"}, want: true},
		{name: "two", args: []string{"path", "ref"}, want: true},
		{name: "three", args: []string{"a", "b", "c"}, want: false},
	}

	for _, tc := range tests {
		if err := cmd.Args(cmd, tc.args); (err == nil) != tc.want {
			t.Fatalf("case %s: expected ok=%v, got err=%v", tc.name, tc.want, err)
		}
	}
}

func TestRemoveArgsValidation(t *testing.T) {
	cmd := newRemoveCmd(&rootOptions{})

	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "missing", args: []string{}, want: false},
		{name: "one", args: []string{"path"}, want: true},
		{name: "two", args: []string{"a", "b"}, want: false},
	}

	for _, tc := range tests {
		if err := cmd.Args(cmd, tc.args); (err == nil) != tc.want {
			t.Fatalf("case %s: expected ok=%v, got err=%v", tc.name, tc.want, err)
		}
	}
}
