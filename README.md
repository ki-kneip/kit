# kit

Your commands, without the make face.

`kit` is a tiny task runner: name your big commands, run them by name — from
any terminal, on any OS. Task bodies run in an **embedded POSIX shell**
([mvdan/sh](https://github.com/mvdan/sh)), so `&&`, pipes, `$VAR`, `if` and
`for` behave exactly the same on Windows, Linux and macOS. No WSL, no Git
Bash, no YAML gymnastics.

```
$ kit deploy staging
$ kit testes/unit
$ kit @sshinto/vm1
```

## Install

```sh
go install github.com/ki-kneip/kit/cmd/kit@latest
```

Or build the Windows MSI installer (see [installer/](installer/)).

## Quick start

```sh
kit init      # creates a starter .kit file
kit hello     # runs the hello task
kit ls        # shows the task files kit can see
```

## Task files

Tasks live in `.kit` files written in kit's own format:

```
# .kit — header is "name: description", body is plain shell, indented.

build: compile the project
    go build -o bin/app ./cmd/app

deploy: build and ship ($1 = target host)
    kit build
    scp bin/app "$1:/opt/app"

test: run tests (pass "all" to include integration)
    if [ "$1" = "all" ]; then go test -tags=integration ./...
    else go test ./...; fi
```

Three rules, no exceptions:

- After the `:` is **always a description** — commands are always indented.
- Task names: letters, digits, `-` and `_`.
- No conditionals in the format itself: the body is real shell, which
  already has `if`, `case` and `for`.

### Where files live, and scopes

The **file name is the scope** — there is no scope syntax inside files:

| File                    | Invocation             |
| ----------------------- | ---------------------- |
| `.kit`                  | `kit build`            |
| `testes.kit`            | `kit testes/unit`      |
| `.kitfiles/deploy.kit`  | `kit deploy/staging`   |
| `.kitfiles/db.yaml`     | `kit db/migrate`       |
| global `sshinto.kit`    | `kit @sshinto/vm1`     |

- kit looks in the **current directory**: `.kit`, `*.kit`, and everything
  inside a `.kitfiles/` folder (where extensionless files are kit format
  and `.yaml` files use the YAML frontend).
- `@` targets the **global directory** (`%AppData%\Roaming\kit` on Windows,
  `~/.config/kit` on Linux) — available from anywhere: `kit @sshinto/vm1`.
- A bare scope name lists its tasks: `kit testes`, `kit @sshinto`.
- Files with the same scope merge; defining the same task twice is an error.

The YAML frontend, for when you want it:

```yaml
migrate:
  desc: run database migrations
  run: goose -dir ./migrations postgres "$DATABASE_URL" up
```

## CLI

```
kit                      face + the reserved verbs
kit <task> [args...]     run a task; args become $1..$n, $@
kit <scope>/<task>       run a scoped task
kit @<scope>/<task>      run a global task
kit <scope>              list the tasks of a scope
```

Reserved verbs (they win over task names):

```
kit help [verb]          overview, or one verb's help
kit init [@][scope] [-k] create a starter file (-k = inside .kitfiles/)
kit list | kit ls [@]    show the task files kit can see
kit shell <cmd...>       run a raw command in the embedded shell
```

Flags (before the target):

```
kit -l 3 build           run 3 times, stop at first failure
kit -e 5s health         rerun every 5s until Ctrl+C
```

## Batteries included

- **Portable coreutils** — `mkdir`, `rm`, `cp`, `mv`, `cat`, `touch`, `ls`
  are built into the binary and work identically everywhere (yes, in
  `kit shell` on Windows too). Builtins win over PATH on purpose.
- **`hey`** — a small HTTP load tester:
  `kit shell hey -n 500 -c 50 https://staging.example.com/health`
- **`kit` inside tasks** — calling `kit other-task` in a body dispatches
  in-process: no child process, Ctrl+C cancels the whole chain.
- **`.env`** — if the current directory has one, its variables are loaded
  into every task run (they override inherited environment values).

## Development

This repo eats its own dog food — see [.kit](.kit):

```sh
kit build    # compile into bin/
kit test     # go test ./...
```

## License

[MIT](LICENSE)
