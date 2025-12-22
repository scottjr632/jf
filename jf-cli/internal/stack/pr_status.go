package stack

import (
	"context"
	"encoding/json"
	"fmt"
)

// PRStatus describes GitHub PR info for a branch.
type PRStatus struct {
	Number int
	State  string
	URL    string
	Title  string
}

type prSummary struct {
	Number int    `json:"number"`
	State  string `json:"state"`
	URL    string `json:"url"`
	Title  string `json:"title"`
}

// PRForBranch returns the PR status for a branch, if any.
func PRForBranch(ctx context.Context, repoRoot, branch string) (*PRStatus, error) {
	out, err := runGh(ctx, repoRoot, "pr", "list", "--head", branch, "--state", "all", "--json", "number,state,url,title")
	if err != nil {
		return nil, err
	}
	var prs []prSummary
	if err := json.Unmarshal([]byte(out), &prs); err != nil {
		return nil, fmt.Errorf("parse gh output: %w", err)
	}
	if len(prs) == 0 {
		return nil, nil
	}
	if len(prs) > 1 {
		return nil, fmt.Errorf("multiple PRs found for %q", branch)
	}
	pr := prs[0]
	return &PRStatus{
		Number: pr.Number,
		State:  pr.State,
		URL:    pr.URL,
		Title:  pr.Title,
	}, nil
}
