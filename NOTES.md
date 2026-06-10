# Install Ledger Notes

## What was set up

- Created the Go CLI project for `github.com/tejasr7/install-ledger`.
- Added the Cobra-based CLI entrypoint at `cmd/il/main.go`.
- Added the core ledger logic at `internal/ledger/ledger.go`.
- Added `.gitignore` so the local build binary `il` is not committed.

## Commands implemented

- `il init` initializes local tracking and installs the zsh hook.
- `il scan` scans installed tools and writes inventory data.
- `il log` shows the install timeline.
- `il today` shows installs tracked today.
- `il find <query>` searches install history.
- `il capture <cwd> <command>` is a hidden command used by the shell hook.

## Local install details

- Built the local binary with:

```bash
go build -o il ./cmd/il
```

- Installed the CLI globally with:

```bash
go install ./cmd/il
```

- The installed binary is at:

```text
/Users/tejasr/go/bin/il
```

- Added this PATH line to `~/.zshrc`:

```bash
export PATH="$HOME/go/bin:$PATH"
```

## Install Ledger data files

- `il init` created the data folder:

```text
/Users/tejasr/.install-ledger
```

- It also created the zsh hook:

```text
/Users/tejasr/.install-ledger/zsh-hook.zsh
```

- Added this source line to `~/.zshrc`:

```bash
source "$HOME/.install-ledger/zsh-hook.zsh"
```

## Verification completed

- `./il --help` works.
- `il --help` works after sourcing `~/.zshrc`.
- `il init` works.
- Manual capture test worked:

```bash
il capture "$PWD" "codex plugin add pm-toolkit@pm-skills"
```

- `il log` shows the captured Codex plugin entry.
- `il find codex` finds the captured entry.
- `il today` shows today’s captured entry.
- `go test ./...` passes.

## Small improvement made

The secret redaction logic was improved so values like:

```text
token=abc123
```

are logged as:

```text
token=REDACTED
```
