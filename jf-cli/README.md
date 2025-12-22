# jf-cli

`jf` is a small helper CLI that streamlines common Git worktree workflows and can passthrough git commands.

## Build

```sh
go build ./...
```

## Usage

Worktree commands:

```sh
jf worktree list
jf worktree new <path|name> [ref]
jf worktree checkout <path|name>
jf worktree main
jf worktree merge <path|name> [--into <branch>]
jf worktree commit [--amend] [--worktree <path|name>] [-- <git commit args...>]
jf worktree amend [--edit] [--worktree <path|name>] [-- <git commit args...>]
jf worktree remove <path|name>
jf worktree prune
```

`jf w` is an alias for `jf worktree`.

`jf worktree commit` runs `git add -p` and then prompts to add untracked files before committing.
`jf amend` and `jf worktree amend` use the same staging flow and default to `--no-edit` unless you pass `--edit`.

When you pass a relative path that does not start with `./` or `../`, `jf`
stores worktrees under `~/.jf/<repo>/worktrees/<name>`. Use `./path` or an
absolute path to opt out.

`jf worktree new` opens a subshell in the new worktree by default. Use
`--no-checkout` to skip that behavior.

Stacked commits:

```sh
jf ls
jf log-long
jf pr
jf trunk [branch]
jf sync
jf restack
jf submit
jf next
jf prev
jf stack status
```

`jf ls` shows commits between trunk and HEAD as the current stack.
`jf log-long` shows stack commits with PR status details.
`jf submit` creates or updates one PR per commit using the `gh` CLI.
`jf restack` ensures stack commits are rebased in order on top of trunk.
Stack metadata (trunk, stacks, and current commit pointer) lives in `.jf/stack.json`.

Examples:

```sh
# Set trunk to main (default)
jf trunk main

# View the current stack of commits (oldest -> newest)
jf ls

# View the current stack with PR status details
jf ll

# Create/update PRs for each commit in the stack
jf submit
 or 
jf ls

# Open the current PR in your browser
jf pr

# Override trunk for a one-off stack listing
jf ls --trunk release

# Refresh stack metadata after manual git history edits
jf sync

# Restack commits onto trunk if needed
jf restack

# Show current stack metadata
jf stack status

# Walk commit stack
jf next
jf prev
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
