# Repository Guidelines

This is a monorepo with two independent projects:

1. **PR Atlas** (`www/apps/dev/`) — A React web app (TanStack Start + Vite) for GitHub PR review management
2. **jf CLI** (`jf-cli/`) — A Go CLI tool for Git worktree and stacked-commit workflows

## Cursor Cloud specific instructions

### Services

| Service | Directory | Dev command | Port | Notes |
|---------|-----------|-------------|------|-------|
| PR Atlas web app | `www/apps/dev` | `pnpm dev` | 3000 | Uses mock/seed data; no DB or API keys required |
| jf CLI | `jf-cli` | `go build ./cmd/jf` | N/A | Standalone binary, no server |

### Web app (`www/apps/dev`)

- **Lint**: `pnpm lint` (ESLint). Pre-existing lint warnings/errors exist in the codebase.
- **Tests**: `pnpm test` (Vitest). Currently no test files exist; `vitest run` exits with code 1 when there are no test files.
- **Dev server**: `pnpm dev` starts Vite on port 3000. The app uses local-only TanStack DB collections with hardcoded seed data — no database or external services needed.
- **Build scripts**: `esbuild` and `unrs-resolver` require approved build scripts. The `pnpm.onlyBuiltDependencies` field in `www/package.json` handles this. Without it, `pnpm install` skips their postinstall and Vite will fail.
- **Storybook**: `pnpm storybook` on port 6006 (optional).
- The PostgreSQL/Drizzle integration is only used by the `/demo/drizzle` route and is not required for the core PR Atlas features.

### Go CLI (`jf-cli`)

- See `jf-cli/AGENTS.md` for coding conventions and commands.
- **Build**: `go build ./...` or `go build -o jf ./cmd/jf`
- **Test**: `go test ./...`
- **Format**: `go fmt ./...`
