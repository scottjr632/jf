package git

import (
	"strings"

	"github.com/scottjr632/sequoia/internal/cli"
)

type NoStagedFilesError struct{}

func (e NoStagedFilesError) Error() string {
	return "no staged files found"
}

func IsNoStagedFilesError(err error) bool {
	_, ok := err.(NoStagedFilesError)
	return ok
}

func Commit(message string) error {
	if err := EnsureStagedFiles(); err != nil {
		return err
	}
	return cli.ExecuteCommandInTerminal("git", "commit", "-m", message)
}

func CherryPick(sha string) error {
	return cli.ExecuteCommandInTerminal("git", "cherry-pick", sha)
}

func ResetSoft(ref string) error {
	return cli.ExecuteCommandInTerminal("git", "reset", "--soft", ref)
}

func ResetMixed(ref string) error {
	return cli.ExecuteCommandInTerminal("git", "reset", ref)
}

func ResetHard(ref string) error {
	return cli.ExecuteCommandInTerminal("git", "reset", "--hard", ref)
}

func CommitAsync(message string) <-chan cli.CmdResult {
	return runGitAsync("commit", "-m", message)
}

func GetCurrentBranchCommitSha() (string, error) {
	sha, err := runGit("rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(sha, "\n", ""), nil
}

// EnsureStagedFiles returns NoStagedFilesError if there are no staged files
func EnsureStagedFiles() error {
	if _, err := runGit("diff", "--quiet", "--cached"); err != nil {
		return nil
	}
	return NoStagedFilesError{}
}

func AmendCommit() error {
	return cli.ExecuteCommandInTerminal("git", "commit", "--amend", "--no-edit")
}
