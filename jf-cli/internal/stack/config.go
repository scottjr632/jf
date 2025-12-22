package stack

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/scottjr632/jf-cli/internal/git"
)

const defaultTrunk = "main"
const defaultRemote = "origin"
const defaultBranchPrefix = "jf/stack"

var runGit = git.Run
var readFile = os.ReadFile
var writeFile = os.WriteFile
var mkdirAll = os.MkdirAll

// CommitMeta stores metadata for a stack commit.
type CommitMeta struct {
	SHA     string `json:"sha"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// StackMeta stores commit ordering for a stack.
type StackMeta struct {
	Trunk   string                `json:"trunk"`
	Order   []string              `json:"order"`
	Commits map[string]CommitMeta `json:"commits"`
	Current string                `json:"current"`
}

// Config stores stack settings for a repository.
type Config struct {
	Trunk        string               `json:"trunk"`
	Remote       string               `json:"remote"`
	BranchPrefix string               `json:"branchPrefix"`
	CurrentStack string               `json:"currentStack"`
	Stacks       map[string]StackMeta `json:"stacks"`
}

// DefaultConfig returns the default configuration for a repo.
func DefaultConfig() Config {
	return Config{
		Trunk:        defaultTrunk,
		Remote:       defaultRemote,
		BranchPrefix: defaultBranchPrefix,
		Stacks:       map[string]StackMeta{},
	}
}

// Load reads stack configuration from disk or returns defaults if missing.
func Load(ctx context.Context, repo string) (Config, error) {
	path, err := ConfigPath(ctx, repo)
	if err != nil {
		return Config{}, err
	}

	data, err := readFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse stack config: %w", err)
	}
	applyDefaults(&cfg)

	return cfg, nil
}

// Save writes stack configuration to disk.
func Save(ctx context.Context, repo string, cfg Config) error {
	path, err := ConfigPath(ctx, repo)
	if err != nil {
		return err
	}
	applyDefaults(&cfg)

	dir := filepath.Dir(path)
	if err := mkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal stack config: %w", err)
	}
	data = append(data, '\n')

	if err := writeFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write stack config: %w", err)
	}

	return nil
}

// ConfigPath returns the location for the stack config file.
func ConfigPath(ctx context.Context, repo string) (string, error) {
	root, err := repoRoot(ctx, repo)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, ".jf", "stack.json"), nil
}

// SetTrunk updates the trunk branch name.
func SetTrunk(cfg *Config, trunk string) error {
	name := strings.TrimSpace(trunk)
	if name == "" {
		return fmt.Errorf("trunk branch is required")
	}
	cfg.Trunk = name
	return nil
}

func applyDefaults(cfg *Config) {
	if strings.TrimSpace(cfg.Trunk) == "" {
		cfg.Trunk = defaultTrunk
	}
	if strings.TrimSpace(cfg.Remote) == "" {
		cfg.Remote = defaultRemote
	}
	if strings.TrimSpace(cfg.BranchPrefix) == "" {
		cfg.BranchPrefix = defaultBranchPrefix
	}
	if cfg.Stacks == nil {
		cfg.Stacks = map[string]StackMeta{}
	}
	for name, stack := range cfg.Stacks {
		if strings.TrimSpace(stack.Trunk) == "" {
			stack.Trunk = cfg.Trunk
		}
		if stack.Commits == nil {
			stack.Commits = map[string]CommitMeta{}
		}
		cfg.Stacks[name] = stack
	}
}

func repoRoot(ctx context.Context, repo string) (string, error) {
	out, err := runGit(ctx, repo, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	root := strings.TrimSpace(out)
	if root == "" {
		return "", fmt.Errorf("git rev-parse --show-toplevel returned empty path")
	}
	return root, nil
}
