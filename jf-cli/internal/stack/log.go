package stack

import (
	"context"
	"fmt"
	"strings"
)

type commitNode struct {
	Commit  Commit
	Parents []string
}

func listCommitsRange(ctx context.Context, repo, trunk, head string) ([]commitNode, error) {
	format := "%H%x1f%P%x1f%h%x1f%s%x1f%b%x1e"
	rangeSpec := trunk + ".." + head
	out, err := runGit(ctx, repo, "log", "--reverse", "--topo-order", "--format="+format, rangeSpec)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(out) == "" {
		return nil, nil
	}

	records := strings.Split(out, "\x1e")
	commits := make([]commitNode, 0, len(records))
	for _, record := range records {
		record = strings.TrimSpace(record)
		if record == "" {
			continue
		}
		fields := strings.Split(record, "\x1f")
		if len(fields) < 5 {
			return nil, fmt.Errorf("unexpected git log output")
		}
		body := strings.TrimSpace(fields[4])
		parents := strings.Fields(strings.TrimSpace(fields[1]))
		commits = append(commits, commitNode{
			Commit: Commit{
				SHA:     strings.TrimSpace(fields[0]),
				Short:   strings.TrimSpace(fields[2]),
				Subject: strings.TrimSpace(fields[3]),
				Body:    body,
			},
			Parents: parents,
		})
	}

	return commits, nil
}
