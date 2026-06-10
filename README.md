# Install Ledger

Install Ledger is a small local CLI for tracking developer machine installs.

It records install-like shell commands, keeps a searchable timeline, and can scan a clean summary of common developer tools installed on your Mac.

```bash
il recent
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

Build and install from this repo:

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

2026-06-10 01:32:02  codex     codex plugin add pm-toolkit@pm-skills  (/Users/tejasr/production/install-ledger)
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

Commands are stored as plain text in `install-log.md`.

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

## v0.2

v0.2 makes the CLI useful for daily use:

- `il recent`
- `il doctor`
- `il path`
- `il scan-summary`
- smaller default `il scan` output

## Roadmap

Planned future work:

- structured install events
- better package name parsing
- export commands
- richer summaries and stats
- optional full inventory scan mode
