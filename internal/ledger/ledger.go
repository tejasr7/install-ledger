package ledger

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Inventory struct {
	ScannedAt string            `json:"scannedAt"`
	System    map[string]string `json:"system"`
	Tools     map[string]string `json:"tools"`
}

func ledgerDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".install-ledger")
}

func logFile() string {
	return filepath.Join(ledgerDir(), "install-log.md")
}

func inventoryFile() string {
	return filepath.Join(ledgerDir(), "inventory.json")
}

func hookFile() string {
	return filepath.Join(ledgerDir(), "zsh-hook.zsh")
}

func ensureDir() error {
	return os.MkdirAll(ledgerDir(), 0755)
}

func Init() error {
	if err := ensureDir(); err != nil {
		return err
	}

	hook := strings.TrimSpace(`
# Install Ledger shell hook
# This tracks install commands locally.

_il_track_command() {
  local cmd="$1"

  if command -v il >/dev/null 2>&1; then
    il capture "$PWD" "$cmd" >/dev/null 2>&1
  fi
}

autoload -Uz add-zsh-hook
add-zsh-hook preexec _il_track_command
`) + "\n"

	if err := os.WriteFile(hookFile(), []byte(hook), 0644); err != nil {
		return err
	}

	home, _ := os.UserHomeDir()
	zshrcPath := filepath.Join(home, ".zshrc")
	sourceLine := `source "$HOME/.install-ledger/zsh-hook.zsh"`

	existing := ""
	if data, err := os.ReadFile(zshrcPath); err == nil {
		existing = string(data)
	}

	if !strings.Contains(existing, sourceLine) {
		f, err := os.OpenFile(zshrcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = f.WriteString("\n# Install Ledger\n" + sourceLine + "\n")
		if err != nil {
			return err
		}
	}

	fmt.Println("Install Ledger initialized.")
	fmt.Println("Data folder:", ledgerDir())
	fmt.Println("")
	fmt.Println("Now run:")
	fmt.Println("source ~/.zshrc")

	return nil
}

func Capture(cwd string, command string) error {
	if !looksLikeInstallCommand(command) {
		return nil
	}

	if err := ensureDir(); err != nil {
		return err
	}

	cleanCommand := redactSecrets(command)
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	line := fmt.Sprintf("%s | %s | %s\n", timestamp, cwd, cleanCommand)

	f, err := os.OpenFile(logFile(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(line)
	return err
}

func Scan() error {
	if err := ensureDir(); err != nil {
		return err
	}

	inventory := Inventory{
		ScannedAt: time.Now().Format(time.RFC3339),
		System: map[string]string{
			"os":       run("uname -s"),
			"kernel":   run("uname -r"),
			"machine":  run("uname -m"),
			"hostname": run("hostname"),
		},
		Tools: map[string]string{
			"brew":              runIfExists("brew", "brew list --versions"),
			"brew_leaves":       runIfExists("brew", "brew leaves"),
			"npm_global":        runIfExists("npm", "npm list -g --depth=0"),
			"pipx":              runIfExists("pipx", "pipx list"),
			"uv_tools":          runIfExists("uv", "uv tool list"),
			"conda_envs":        runIfExists("conda", "conda env list"),
			"vscode_extensions": runIfExists("code", "code --list-extensions"),
			"codex_plugins":     runIfExists("codex", "codex plugin list"),
			"mac_pkgutil":       run("pkgutil --pkgs"),
		},
	}

	data, err := json.MarshalIndent(inventory, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(inventoryFile(), data, 0644); err != nil {
		return err
	}

	fmt.Println("Inventory saved:", inventoryFile())
	return nil
}

func ShowLog() error {
	data, err := os.ReadFile(logFile())
	if err != nil {
		fmt.Println("No install log found yet.")
		fmt.Println("Run: il init")
		return nil
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		fmt.Println("Install log is empty.")
		return nil
	}

	fmt.Println(content)
	return nil
}

func ShowToday() error {
	data, err := os.ReadFile(logFile())
	if err != nil {
		fmt.Println("No install log found yet.")
		return nil
	}

	today := time.Now().Format("2006-01-02")
	lines := strings.Split(string(data), "\n")

	found := false

	for _, line := range lines {
		if strings.HasPrefix(line, today) {
			fmt.Println(line)
			found = true
		}
	}

	if !found {
		fmt.Println("No installs tracked today.")
	}

	return nil
}

func Find(query string) error {
	data, err := os.ReadFile(logFile())
	if err != nil {
		fmt.Println("No install log found yet.")
		return nil
	}

	lines := strings.Split(string(data), "\n")
	q := strings.ToLower(query)

	found := false

	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), q) {
			fmt.Println(line)
			found = true
		}
	}

	if !found {
		fmt.Println("No matches found for:", query)
	}

	return nil
}

func looksLikeInstallCommand(command string) bool {
	patterns := []string{
		"brew install",
		"brew tap",
		"npm install -g",
		"npm i -g",
		"pnpm add -g",
		"yarn global add",
		"pip install",
		"python -m pip install",
		"python3 -m pip install",
		"pipx install",
		"uv tool install",
		"cargo install",
		"go install",
		"gem install",
		"conda install",
		"code --install-extension",
		"codex plugin add",
		"codex plugin marketplace add",
	}

	lower := strings.ToLower(command)

	for _, pattern := range patterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	return false
}

func redactSecrets(command string) string {
	secretPattern := regexp.MustCompile(`(?i)(password|token|apikey|api_key|secret)=("[^"]*"|'[^']*'|[^[:space:]]+)`)
	return secretPattern.ReplaceAllString(command, `${1}=REDACTED`)
}

func runIfExists(binary string, command string) string {
	if !commandExists(binary) {
		return ""
	}

	return run(command)
}

func commandExists(binary string) bool {
	_, err := exec.LookPath(binary)
	return err == nil
}

func run(command string) string {
	cmd := exec.Command("zsh", "-lc", command)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}
