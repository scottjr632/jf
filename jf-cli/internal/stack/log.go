package stack

import (
	"context"
	"fmt"
	"strings"
)

func listCommitsRange(ctx context.Context, repo, trunk, head string) ([]Commit, error) {
	format := "%H%x1f%h%x1f%s%x1f%b%x1e"
	rangeSpec := trunk + ".." + head
	out, err := runGit(ctx, repo, "log", "--reverse", "--format="+format, rangeSpec)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(out) == "" {
		return nil, nil
	}

	records := strings.Split(out, "\x1e")
	commits := make([]Commit, 0, len(records))
	for _, record := range records {
		record = strings.TrimSpace(record)
		if record == "" {
			continue
		}
		fields := strings.Split(record, "\x1f")
		if len(fields) < 4 {
			return nil, fmt.Errorf("unexpected git log output")
		}
		body := strings.TrimSpace(fields[3])
		commits = append(commits, Commit{
			SHA:     strings.TrimSpace(fields[0]),
			Short:   strings.TrimSpace(fields[1]),
			Subject: strings.TrimSpace(fields[2]),
			Body:    body,
		})
	}

	return commits, nil
}
