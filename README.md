# tod

A fast, beautiful to-do list for your terminal.

```
  OVERDUE · 1
   6  ○  Pay rent  !!! yesterday

  TODAY · 2
   4  ○  Water plants  today ↻ daily
   1  ○  Buy milk  !!! @groceries

  ██████████░░░░░░░░░░  3/6 done · 50%  1 overdue
```

`tod` is a single static binary with zero config. Commands are instant and
scriptable; running `tod` with no arguments opens a full interactive UI.
Every change is undoable.

## Install

```bash
go install github.com/PeteXC/tod@latest
```

Or build from source:

```bash
git clone https://github.com/PeteXC/tod && cd tod
go build -o tod .
```

Optional shell completions:

```bash
source <(tod completion bash)                                  # bash
tod completion zsh > "${fpath[1]}/_tod"                        # zsh
tod completion fish > ~/.config/fish/completions/tod.fish      # fish
```

## Quick start

```bash
tod add "Buy milk" !high @groceries due:tomorrow
tod add "Submit report" '#work' '!!!' due:fri
tod add "Water the plants" every:day
tod ls
tod done 1
tod          # interactive mode
```

> Note: quote `#project` and `'!!!'` so your shell doesn't treat them as a
> comment or history expansion.

## Metadata, inline

Metadata goes anywhere in the task text when adding or editing:

| Syntax        | Meaning                                              |
| ------------- | ---------------------------------------------------- |
| `!high` `!med` `!low` | Priority (or `!!!` `!!` `!`; `!none` clears) |
| `@tag`        | Tag, repeatable                                      |
| `#project`    | Project                                              |
| `due:<when>`  | `today`, `tomorrow`, `fri`, `+3d`, `+2w`, `2026-09-01` (`due:none` clears) |
| `every:<span>`| Recurrence: `day`, `weekday`, `week`, `month`, `mon`, `3d`, `2w` |

## Commands

| Command                     | What it does                                        |
| --------------------------- | --------------------------------------------------- |
| `tod`                       | Interactive UI (plain list when piped)              |
| `tod add <text> [meta]`     | Add a task                                          |
| `tod ls [filters]`          | List tasks, grouped by urgency                      |
| `tod done <id>...`          | Complete tasks (ranges ok: `tod done 1-3`)          |
| `tod undone <id>...`        | Reopen tasks                                        |
| `tod rm <id>...`            | Delete tasks                                        |
| `tod edit <id> <changes>`   | Edit text and/or metadata                           |
| `tod pri <id> <level>`      | Set priority: `high`, `medium`, `low`, `none`       |
| `tod due <id> <when>`       | Set due date (`none` clears)                        |
| `tod search <query>`        | Find tasks (same as `tod ls <query>`)               |
| `tod stats`                 | Productivity dashboard: sparkline, streaks, projects|
| `tod undo` / `tod redo`     | Every change is undoable                            |
| `tod clear [--force]`       | Remove completed tasks                              |
| `tod export`                | Print all tasks as JSON                             |
| `tod path`                  | Show where data lives                               |
| `tod completion <shell>`    | Shell completion script                             |
| `tod help`                  | Full help                                           |

### Listing filters

```bash
tod ls --all            # include completed (last 10)
tod ls --done           # completed only
tod ls @groceries       # by tag
tod ls '#work'          # by project
tod ls --pri high       # by priority
tod ls milk             # free-text search
tod ls --plain          # ASCII output for scripts
```

## Interactive mode

Run `tod` with no arguments:

```
space done · a add · e edit · d delete · u undo · / filter
tab show all · 1/2/3 priority · t due today · T tomorrow · q quit
```

The add/edit inputs understand the same inline metadata as the CLI.

## Recurring tasks

```bash
tod add "Water the plants" every:day
tod add "Standup notes" every:weekday
tod add "Call grandma" every:sun
```

Completing a recurring task spawns its next occurrence, always strictly in
the future — completing an overdue daily task schedules tomorrow, not a
backlog of catch-up copies.

## Your data

Everything lives in plain JSON in `~/.tod` (override with `$TOD_HOME`):

- `tasks.json` — your tasks (atomic writes; a crash can't corrupt it)
- `undo.json` — the last 50 changes, so `tod undo` works across sessions

`tod export` prints the whole store as JSON for backups or scripting.

## Development

```bash
go build -o tod .   # build
go test ./...       # test
go vet ./...        # lint
```

Requires Go 1.20+.

## License

MIT — see [LICENSE](LICENSE).
