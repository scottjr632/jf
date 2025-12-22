# jf-cli

`jf` is a small helper CLI that streamlines common Git worktree workflows and can passthrough git commands.

## Build

```sh
go build ./...
```

## Usage

Worktree commands:

```sh
jf list
jf new <path|name> [ref]
jf checkout <path|name>
jf main
jf merge <path|name> [--into <branch>]
jf commit [--amend] [--worktree <path|name>] [-- <git commit args...>]
jf amend [--edit] [--worktree <path|name>] [-- <git commit args...>]
jf remove <path|name>
jf prune
```

`jf commit` runs `git add -p` and then prompts to add untracked files before committing.
`jf amend` uses the same staging flow and defaults to `--no-edit` unless you pass `--edit`.

When you pass a relative path that does not start with `./` or `../`, `jf`
stores worktrees under `~/.jf/<repo>/worktrees/<name>`. Use `./path` or an
absolute path to opt out.

`jf new` opens a subshell in the new worktree by default. Use
`--no-checkout` to skip that behavior.

Stacked commits:

```sh
jf ls
jf trunk [branch]
jf submit
```

`jf ls` shows commits between trunk and HEAD as the current stack.
`jf submit` creates or updates one PR per commit using the `gh` CLI.
Trunk settings are stored in `.jf/stack.json`.

Examples:

```sh
# Set trunk to main (default)
jf trunk main

# View the current stack of commits (oldest -> newest)
jf ls

# Create/update PRs for each commit in the stack
jf submit

# Override trunk for a one-off stack listing
jf ls --trunk release
```

Git passthrough:

```sh
jf status
jf log --oneline
jf -C /path/to/repo status
```

Dedicated git subcommand:

```sh
jf git status
jf -C /path/to/repo git log --oneline
```
