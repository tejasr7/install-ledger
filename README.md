# Install Ledger

Install Ledger is a small local CLI for tracking developer machine installs.

It records install-like shell commands, keeps a searchable timeline, and can scan a clean summary of common developer tools installed on your Mac.

```bash
il recent
il events
il today
il find codex
il scan-summary
```

Install Ledger is local-only. It writes to `~/.install-ledger` and does not sync data to any server.

## Why

Developer machines change constantly: Homebrew packages, npm globals, VS Code extensions, Python tools, Codex plugins, CLIs, databases, and more.

Install Ledger gives you a simple local history so you can answer questions like:

- What did I install recently?
- Did I install this tool today?
- Where is my install history stored?
- Is the shell hook working?
- What tools are currently installed?

## Install

Install directly with Go:

```bash
go install github.com/tejasr7/install-ledger/cmd/il@latest
```

This installs the `il` binary into your Go bin directory, usually:

```text
~/go/bin
```

For local development, build and install from this repo:

```bash
go build -o il ./cmd/il
go install ./cmd/il
```

Make sure Go binaries are on your PATH:

```bash
export PATH="$HOME/go/bin:$PATH"
```

Initialize tracking:

```bash
il init
source ~/.zshrc
```

## Quick Start

Check that everything is installed correctly:

```bash
il doctor
```

Show recent install events:

```bash
il recent
```

Show structured install events:

```bash
il events
```

Search history:

```bash
il find codex
```

Scan your current developer tool inventory:

```bash
il scan
il scan-summary
```

## Commands

| Command | Description |
| --- | --- |
| `il init` | Create the data folder and install the zsh tracking hook. |
| `il recent` | Show the 10 most recent install events. |
| `il recent -n 5` | Show a custom number of recent install events. |
| `il events` | Show structured install events from `events.jsonl`. |
| `il events -n 5` | Show a custom number of structured events. |
| `il migrate` | Migrate old `install-log.md` entries into structured events. |
| `il today` | Show install events captured today. |
| `il log` | Show the full install timeline. |
| `il find <query>` | Search install history. |
| `il scan` | Save a clean inventory of common developer tools. |
| `il scan-summary` | Show readable counts from the latest inventory scan. |
| `il doctor` | Check whether tracking is installed correctly. |
| `il path` | Show Install Ledger data paths. |

## Examples

```bash
il recent
il recent -n 20
il events
il events -n 5
il migrate
il today
il find brew
il find npm
il find codex
il doctor
il path
```

Example recent output:

```text
Recent installs

2026-06-10 01:32  codex     plugin_add          pm-toolkit               from pm-skills
```

Example doctor output:

```text
Install Ledger Doctor

[OK] il binary found
[OK] data directory exists
[OK] zsh hook exists
[OK] ~/.zshrc sources Install Ledger hook
[OK] install log exists
[OK] inventory file exists

Required checks: 4/4 passed
```

## What Gets Tracked

The shell hook captures install-like commands such as:

- `brew install`
- `brew tap`
- `npm install -g`
- `pnpm add -g`
- `yarn global add`
- `pip install`
- `pipx install`
- `uv tool install`
- `cargo install`
- `go install`
- `gem install`
- `conda install`
- `code --install-extension`
- `codex plugin add`

Commands are stored in two formats:

- `install-log.md` for a plain-text human-readable history
- `events.jsonl` for structured install events

Structured events look like this:

```json
{"id":"...","timestamp":"2026-06-10T12:30:00+05:30","cwd":"/Users/tejasr/project","manager":"codex","action":"plugin_add","category":"plugin","name":"pm-toolkit","source":"pm-skills","rawCommand":"codex plugin add pm-toolkit@pm-skills","status":"captured"}
```

If you have older plain-text logs, migrate them with:

```bash
il migrate
```

## Inventory Scan

`il scan` writes a compact inventory to:

```text
~/.install-ledger/inventory.json
```

The v0.2 default scan intentionally avoids huge system dumps. It focuses on useful daily developer inventory:

- Homebrew manual packages
- npm global packages
- pipx tools
- uv tools
- conda environments
- VS Code extensions
- basic system information

Use:

```bash
il scan-summary
```

to see a readable summary instead of opening the JSON file.

## Data Files

Install Ledger stores data here:

```text
~/.install-ledger/
  install-log.md
  events.jsonl
  inventory.json
  zsh-hook.zsh
```

Show these paths with:

```bash
il path
```

## Privacy

Install Ledger is local-first and local-only.

- No cloud account
- No background server
- No telemetry
- No network sync

Commands are written to a local Markdown log. Basic secret redaction is applied for values like `token=...`, `password=...`, `api_key=...`, and `secret=...`.

## v0.3.0

v0.3.0 adds structured install events:

- `events.jsonl`
- `il events`
- `il events -n 5`
- `il migrate`
- `il recent` now prefers structured events and falls back to the old text log

The old `install-log.md` file is still written for backward compatibility.

## v0.2

v0.2 makes the CLI useful for daily use:

- `il recent`
- `il doctor`
- `il path`
- `il scan-summary`
- smaller default `il scan` output

## Roadmap

Planned future work:

- better package name parsing
- export commands
- richer summaries and stats
- optional full inventory scan mode

## License

MIT License. See [LICENSE](LICENSE).
