package stack

import "context"

// StatusReport describes the current stack state.
type StatusReport struct {
	Name         string
	Trunk        string
	Head         string
	CurrentID    string
	CurrentSHA   string
	CurrentShort string
	Count        int
}

// Status returns stack metadata for the current stack.
func Status(ctx context.Context, repo string, cfg *Config, trunkOverride string) (StatusReport, error) {
	resolved, err := resolveStack(ctx, repo, cfg, trunkOverride)
	if err != nil {
		return StatusReport{}, err
	}
	if resolved.changed {
		if err := Save(ctx, repo, *cfg); err != nil {
			return StatusReport{}, err
		}
	}

	headLabel := resolved.headRef
	if resolved.detached {
		label, err := currentShortSHA(ctx, repo)
		if err != nil {
			return StatusReport{}, err
		}
		headLabel = label
	}

	currentSHA := ""
	currentShort := ""
	if resolved.stack.Current != "" {
		if meta, ok := resolved.stack.Commits[resolved.stack.Current]; ok {
			currentSHA = meta.SHA
			currentShort = shortSHA(meta.SHA)
		}
	}

	return StatusReport{
		Name:         resolved.name,
		Trunk:        resolved.effectiveTrunk,
		Head:         headLabel,
		CurrentID:    resolved.stack.Current,
		CurrentSHA:   currentSHA,
		CurrentShort: currentShort,
		Count:        len(resolved.stack.Commits),
	}, nil
}
