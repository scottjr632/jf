# Repository Guidelines

This monorepo contains two independent projects:

| Project | Path | Stack |
|---------|------|--------|
| **PR Atlas** (web app) | `www/apps/dev/` | React, TanStack Start + Router, Vite, Tailwind |
| **jf CLI** | `jf-cli/` | Go (Cobra) |

There is no single root `package.json`; the JavaScript workspace lives under `www/` (see below).

## Repository layout

```
/
├── AGENTS.md          ← This file (monorepo overview)
├── www/               ← pnpm workspace (see www/pnpm-workspace.yaml)
│   ├── package.json   ← Workspace root; `pnpm.onlyBuiltDependencies` for esbuild / unrs-resolver
│   └── apps/dev/      ← PR Atlas app — run app scripts from here or via `pnpm --filter dev`
└── jf-cli/            ← Go module; see jf-cli/AGENTS.md for CLI-specific conventions
```

## Prerequisites

- **Web**: [pnpm](https://pnpm.io/) 10.x (see `packageManager` in `www/package.json`).
- **CLI**: Go toolchain for building and testing `jf-cli/`.

## PR Atlas web app (`www/apps/dev`)

### Install

From the repo root:

```bash
cd www && pnpm install
```

### Commands

Run from `www/apps/dev`, or from `www` using `pnpm --filter dev <script>` (e.g. `pnpm --filter dev dev`).

| Task | Command |
|------|---------|
| Dev server (Vite, port 3000) | `pnpm dev` |
| Production build | `pnpm build` |
| Preview production build | `pnpm preview` |
| Lint | `pnpm lint` (ESLint). Pre-existing lint warnings/errors may exist. |
| Tests | `pnpm test` (Vitest). There are currently no `*.test` / `*.spec` files under `www/`; Vitest exits with code 1 when no tests are found. |
| Format / check | `pnpm format` (Prettier); `pnpm check` runs Prettier write + ESLint fix |
| Storybook | `pnpm storybook` (port 6006) |

### Notes

- The app uses local-only TanStack DB collections with hardcoded seed data — no database or API keys are required for core features.
- **Build scripts**: `esbuild` and `unrs-resolver` require approved postinstall scripts. The `pnpm.onlyBuiltDependencies` field in `www/package.json` must include them; without it, `pnpm install` can skip postinstall and Vite may fail.
- PostgreSQL / Drizzle is only used by the `/demo/drizzle` route and is not required for core PR Atlas behavior.

## jf CLI (`jf-cli`)

See **[jf-cli/AGENTS.md](jf-cli/AGENTS.md)** for package layout, style, and agent-oriented CLI notes.

| Task | Command |
|------|---------|
| Build | `go build ./...` or `go build -o jf ./cmd/jf` |
| Test | `go test ./...` |
| Format | `go fmt ./...` |

## Cursor Cloud — services

| Service | Directory | Dev command | Port | Notes |
|---------|-----------|-------------|------|-------|
| PR Atlas web app | `www/apps/dev` | `pnpm dev` | 3000 | Mock/seed data; no DB or API keys required |
| jf CLI | `jf-cli` | `go build ./cmd/jf` | N/A | Standalone binary, no server |
