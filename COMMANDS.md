# Install Ledger Commands

## Daily commands

### Initialize tracking

```bash
il init
```

Creates the Install Ledger data folder and zsh hook.

After running it, reload your shell:

```bash
source ~/.zshrc
```

### Scan installed tools

```bash
il scan
```

Scans installed tools and saves the inventory to:

```text
~/.install-ledger/inventory.json
```

### Show install log

```bash
il log
```

Shows the install commands that Install Ledger has captured.

### Show today's installs

```bash
il today
```

Shows install commands captured today.

### Search install history

```bash
il find codex
```

Searches the install log for `codex`.

You can replace `codex` with any query:

```bash
il find brew
il find npm
il find python
il find plugin
```

## Hidden command

### Capture an install command manually

```bash
il capture "$PWD" "codex plugin add pm-toolkit@pm-skills"
```

This is mainly used by the shell hook, but you can run it manually for testing.

## Help

```bash
il --help
```

Shows all available commands.

For help with a specific command:

```bash
il init --help
il scan --help
il log --help
il today --help
il find --help
```
