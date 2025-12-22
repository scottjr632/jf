# Plan

We will extend `jf` with a commit-stack workflow backed by local metadata files and GitHub PR creation/update via the `gh` CLI, defaulting trunk to `main` while allowing user override.

## Scope
- In: stack metadata format + storage, CLI commands to manage stacks, PR submission/update for a stack via `gh`, sync/rebase helpers, tests/docs updates.
- Out: non-GitHub providers, server-side services, UI features beyond the CLI.

## Action items
 [x] Review current CLI flow and identify hooks for commit-based stacks.
 [x] Define config format for trunk/remote and stack discovery from commits.
 [x] Implement `jf ls` for current stack (commit list) and update output formatting.
 [x] Implement `jf submit` to create/update PRs per commit via `gh` with stacked bases.
 [x] Wire config + stack helpers; update docs/tests; run `go test ./...`.

## Design
- Stacks are persisted in `.jf/stack.json` with UUID-backed commit entries and a current pointer.
- `jf commit` appends a new stack item (UUID + SHA) and updates the pointer.
- `jf amend` updates the current stack item SHA and rebases descendants when needed.
- `jf ls` lists the current stack items from metadata.
- `jf submit` creates a PR per stack item and infers PR titles from commit subjects.
- Commits are treated as atomic units of change.
