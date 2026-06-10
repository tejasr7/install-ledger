# Install Ledger

Install Ledger is a local CLI that tracks developer install commands and keeps a simple timeline of changes to your machine.

The command is:

```bash
il
```

It is local-only. Install Ledger writes data under `~/.install-ledger` and does not sync anything to a server.

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

Initialize shell tracking:

```bash
il init
source ~/.zshrc
```

## Daily usage

```bash
il recent
il today
il log
il find codex
il scan
il scan-summary
```

## Commands

### `il init`

Creates `~/.install-ledger`, writes the zsh hook, and adds the hook source line to `~/.zshrc`.

### `il recent`

Shows the latest install events, newest first.

```bash
il recent
il recent -n 5
```

### `il today`

Shows install events captured today.

### `il log`

Shows the full plain-text install log.

### `il find <query>`

Searches install history.

```bash
il find brew
il find codex
il find python
```

### `il scan`

Scans a clean default inventory of common developer tooling and saves it to:

```text
~/.install-ledger/inventory.json
```

The default scan intentionally avoids very noisy system dumps.

### `il scan-summary`

Shows a readable count summary from the latest inventory file.

### `il doctor`

Checks whether Install Ledger is properly installed:

- `il` binary is on PATH
- data directory exists
- zsh hook exists
- `~/.zshrc` sources the hook
- install log exists
- inventory file exists

### `il path`

Prints the Install Ledger data directory and file paths.

## Data files

```text
~/.install-ledger/
  install-log.md
  inventory.json
  zsh-hook.zsh
```

## v0.2 changes

v0.2 made the CLI more useful for daily use:

- Added `il recent`
- Added `il doctor`
- Added `il path`
- Added `il scan-summary`
- Cleaned up `il scan` so the default inventory is smaller and easier to inspect

## Roadmap

v0.3 will focus on structured install events. The current log stays plain text in v0.2.
