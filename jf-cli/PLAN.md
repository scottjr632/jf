# Plan

We will extend `jf` with a stacked-branch workflow backed by local metadata files and GitHub PR creation/update via the `gh` CLI, defaulting trunk to `main` while allowing user override.

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
- Stacks are automatically created when a new commit is created
- `jf commit` automatically creates a new commit that is treated as a new stack item
- `jf amend` amends to the current commit
- `jf ls` lists the current stack and its items which are commits
- `jf submit` creates a PR for the current stack, one PR per commit and the PR name should be automatically in inferred
- we should treat commits as atomic units that hold one logical changes
