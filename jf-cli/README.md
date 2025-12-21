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
jf add <path|name> [ref]
jf checkout <path|name>
jf merge <path|name> [--into <branch>]
jf commit [--amend] [--worktree <path|name>] [-- <git commit args...>]
jf remove <path|name>
jf prune
```

When you pass a relative path that does not start with `./` or `../`, `jf`
stores worktrees under `~/.jf/<repo>/worktrees/<name>`. Use `./path` or an
absolute path to opt out.

`jf add` opens a subshell in the new worktree by default. Use
`--no-checkout` to skip that behavior.

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
