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
	capturedAt := time.Now()
	timestamp := capturedAt.Format("2006-01-02 15:04:05")

	event := NewInstallEvent(capturedAt, cwd, cleanCommand)
	if err := AppendEventIfNew(event); err != nil {
		return err
	}

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
			"brew_leaves":       runIfExists("brew", "brew leaves"),
			"npm_global":        runIfExists("npm", "npm list -g --depth=0"),
			"pipx":              runIfExists("pipx", "pipx list"),
			"uv_tools":          runIfExists("uv", "uv tool list"),
			"conda_envs":        runIfExists("conda", "conda env list"),
			"vscode_extensions": runIfExists("code", "code --list-extensions"),
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

func ShowRecent(limit int) error {
	return ShowEventRecent(limit)
}

func ShowRecentFromLog(limit int) error {
	data, err := os.ReadFile(logFile())
	if err != nil {
		fmt.Println("No install log found yet.")
		fmt.Println("Run: il init")
		return nil
	}

	lines := cleanLines(string(data))
	if len(lines) == 0 {
		fmt.Println("Install log is empty.")
		return nil
	}

	if limit <= 0 {
		limit = 10
	}

	start := len(lines) - limit
	if start < 0 {
		start = 0
	}

	fmt.Println("Recent installs")
	fmt.Println("")

	for i := len(lines) - 1; i >= start; i-- {
		fmt.Println(formatLogLine(lines[i]))
	}

	return nil
}

func Doctor() error {
	home, _ := os.UserHomeDir()
	zshrcPath := filepath.Join(home, ".zshrc")
	sourceLine := `source "$HOME/.install-ledger/zsh-hook.zsh"`

	fmt.Println("Install Ledger Doctor")
	fmt.Println("")

	passed := 0
	required := 4

	if commandExists("il") {
		checkOK("il binary found")
		passed++
	} else {
		checkFail("il binary not found in PATH")
	}

	if fileExists(ledgerDir()) {
		checkOK("data directory exists")
		passed++
	} else {
		checkFail("data directory missing")
	}

	if fileExists(hookFile()) {
		checkOK("zsh hook exists")
		passed++
	} else {
		checkFail("zsh hook missing")
	}

	zshrcContent := ""
	if data, err := os.ReadFile(zshrcPath); err == nil {
		zshrcContent = string(data)
	}

	if strings.Contains(zshrcContent, sourceLine) {
		checkOK("~/.zshrc sources Install Ledger hook")
		passed++
	} else {
		checkFail("~/.zshrc does not source Install Ledger hook")
	}

	if fileExists(logFile()) {
		checkOK("install log exists")
	} else {
		checkWarn("install log does not exist yet")
	}

	if fileExists(inventoryFile()) {
		checkOK("inventory file exists")
	} else {
		checkWarn("inventory file does not exist yet")
	}

	fmt.Println("")
	fmt.Printf("Required checks: %d/%d passed\n", passed, required)
	fmt.Println("")
	fmt.Println("Data folder:")
	fmt.Println(ledgerDir())

	return nil
}

func ShowPath() error {
	fmt.Println("Install Ledger paths")
	fmt.Println("")
	fmt.Println("Data folder:")
	fmt.Println(ledgerDir())
	fmt.Println("")
	fmt.Println("Files:")
	fmt.Println("-", logFile())
	fmt.Println("-", inventoryFile())
	fmt.Println("-", eventsFile())
	fmt.Println("-", hookFile())
	fmt.Println("-", exportsDir())

	return nil
}

func ScanSummary() error {
	data, err := os.ReadFile(inventoryFile())
	if err != nil {
		fmt.Println("No inventory found yet.")
		fmt.Println("Run: il scan")
		return nil
	}

	var inventory Inventory
	if err := json.Unmarshal(data, &inventory); err != nil {
		return err
	}

	fmt.Println("Install Ledger Scan Summary")
	fmt.Println("")
	fmt.Println("Scanned at:", inventory.ScannedAt)
	fmt.Println("")

	if len(inventory.System) > 0 {
		fmt.Println("System")
		printSystemValue("OS", inventory.System["os"])
		printSystemValue("Kernel", inventory.System["kernel"])
		printSystemValue("Machine", inventory.System["machine"])
		printSystemValue("Host", inventory.System["hostname"])
		fmt.Println("")
	}

	fmt.Println("Tools")
	printToolCount("Homebrew manual packages", inventory.Tools["brew_leaves"], countCleanLines)
	printToolCount("npm global packages", inventory.Tools["npm_global"], countNPMGlobals)
	printToolCount("pipx tools", inventory.Tools["pipx"], countCleanLines)
	printToolCount("uv tools", inventory.Tools["uv_tools"], countUVTools)
	printToolCount("conda environments", inventory.Tools["conda_envs"], countCondaEnvs)
	printToolCount("VS Code extensions", inventory.Tools["vscode_extensions"], countCleanLines)

	return nil
}

func cleanLines(content string) []string {
	rawLines := strings.Split(content, "\n")
	lines := make([]string, 0, len(rawLines))

	for _, line := range rawLines {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}

	return lines
}

func formatLogLine(line string) string {
	parts := strings.SplitN(line, " | ", 3)
	if len(parts) != 3 {
		return line
	}

	timestamp := parts[0]
	cwd := parts[1]
	command := parts[2]
	manager := detectManager(command)

	return fmt.Sprintf("%s  %-8s  %s  (%s)", timestamp, manager, command, cwd)
}

func detectManager(command string) string {
	lower := strings.ToLower(command)

	switch {
	case strings.Contains(lower, "brew "):
		return "brew"
	case strings.Contains(lower, "npm "):
		return "npm"
	case strings.Contains(lower, "pnpm "):
		return "pnpm"
	case strings.Contains(lower, "yarn "):
		return "yarn"
	case strings.Contains(lower, "pipx "):
		return "pipx"
	case strings.Contains(lower, "pip "):
		return "pip"
	case strings.Contains(lower, "uv "):
		return "uv"
	case strings.Contains(lower, "cargo "):
		return "cargo"
	case strings.Contains(lower, "go install"):
		return "go"
	case strings.Contains(lower, "gem "):
		return "gem"
	case strings.Contains(lower, "conda "):
		return "conda"
	case strings.Contains(lower, "code --install-extension"):
		return "vscode"
	case strings.Contains(lower, "codex "):
		return "codex"
	default:
		return "unknown"
	}
}

func printSystemValue(label string, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}

	fmt.Printf("- %s: %s\n", label, value)
}

func printToolCount(label string, raw string, count func(string) int) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		fmt.Printf("- %-28s %s\n", label+":", "not found / empty")
		return
	}

	total := count(raw)
	if total == 0 {
		fmt.Printf("- %-28s %s\n", label+":", "not found / empty")
		return
	}

	fmt.Printf("- %-28s %d\n", label+":", total)
}

func countCleanLines(raw string) int {
	return len(cleanLines(raw))
}

func countNPMGlobals(raw string) int {
	count := 0
	for _, line := range cleanLines(raw) {
		if strings.Contains(line, "── ") {
			count++
		}
	}

	return count
}

func countUVTools(raw string) int {
	count := 0
	for _, line := range cleanLines(raw) {
		if !strings.HasPrefix(line, "- ") {
			count++
		}
	}

	return count
}

func countCondaEnvs(raw string) int {
	count := 0
	for _, line := range cleanLines(raw) {
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "/") {
			count++
		}
	}

	return count
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func checkOK(message string) {
	fmt.Println("[OK]", message)
}

func checkFail(message string) {
	fmt.Println("[FAIL]", message)
}

func checkWarn(message string) {
	fmt.Println("[WARN]", message)
}

func looksLikeInstallCommand(command string) bool {
	manager, _, _, name, source := parseInstallCommand(command)
	return manager != "unknown" && (name != "" || source != "")
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
