# Repository Guidelines

## Project Structure & Module Organization
- `cmd/jf` holds the CLI entrypoint.
- `internal/cli` contains Cobra commands for the jf CLI.
- `internal/git` wraps git execution helpers.
- `internal/worktree` contains git worktree-specific operations.
- Keep Go packages under clear top-level folders (`cmd/` for CLI entrypoints and `internal/` for private packages).

## Build, Test, and Development Commands
- `go build ./...` builds all packages in the module.
- `go test ./...` runs all tests.
- `go fmt ./...` formats all Go files; run before committing changes.

## Coding Style & Naming Conventions
- Use standard Go formatting (`gofmt`); prefer tabs for indentation as per Go conventions.
- Package names should be short, lowercase, and descriptive (`cli`, `config`, `auth`).
- File names should be lowercase with underscores when needed (`config_loader.go`).
- Don't add unnecessary comments to the code.

## Testing Guidelines
- Use Go’s built-in `testing` package with files named `*_test.go`.
- Keep tests close to their packages; name test functions `TestXxx` and table tests `TestXxx_Table`.
- Add tests for new behavior; there is no stated coverage target yet.

## Commit & Pull Request Guidelines
- No Git history is available in this workspace, so commit conventions are unknown.
- Until conventions are defined, write clear, imperative commit messages (for example, `add cli scaffold`).
- Pull requests should include a concise summary, rationale, and any relevant run commands or test results.

## Agent Notes
- The CLI supports passthrough git commands and a dedicated `jf git` subcommand; keep docs in sync when CLI behavior changes.
