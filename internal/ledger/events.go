package ledger

import (
	"bufio"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type InstallEvent struct {
	ID         string `json:"id"`
	Timestamp  string `json:"timestamp"`
	CWD        string `json:"cwd"`
	Manager    string `json:"manager"`
	Action     string `json:"action"`
	Category   string `json:"category"`
	Name       string `json:"name"`
	Source     string `json:"source,omitempty"`
	RawCommand string `json:"rawCommand"`
	Status     string `json:"status"`
}

func eventsFile() string {
	return filepath.Join(ledgerDir(), "events.jsonl")
}

func NewInstallEvent(timestamp time.Time, cwd string, command string) InstallEvent {
	manager, action, category, name, source := parseInstallCommand(command)

	event := InstallEvent{
		Timestamp:  timestamp.Format(time.RFC3339),
		CWD:        cwd,
		Manager:    manager,
		Action:     action,
		Category:   category,
		Name:       name,
		Source:     source,
		RawCommand: command,
		Status:     "captured",
	}

	event.ID = eventID(event)
	return event
}

func AppendEvent(event InstallEvent) error {
	if err := ensureDir(); err != nil {
		return err
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(eventsFile(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(string(data) + "\n")
	return err
}

func AppendEventIfNew(event InstallEvent) error {
	events, err := ReadEvents()
	if err != nil {
		return err
	}

	for _, existing := range events {
		if existing.ID == event.ID {
			return nil
		}
	}

	return AppendEvent(event)
}

func ReadEvents() ([]InstallEvent, error) {
	f, err := os.Open(eventsFile())
	if err != nil {
		if os.IsNotExist(err) {
			return []InstallEvent{}, nil
		}
		return nil, err
	}
	defer f.Close()

	events := []InstallEvent{}
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var event InstallEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		events = append(events, event)
	}

	return events, scanner.Err()
}

func ShowEvents(limit int) error {
	events, err := ReadEvents()
	if err != nil {
		return err
	}

	if len(events) == 0 {
		fmt.Println("No structured events found yet.")
		fmt.Println("Run: il migrate")
		return nil
	}

	if limit <= 0 {
		limit = 20
	}

	start := len(events) - limit
	if start < 0 {
		start = 0
	}

	fmt.Println("Install Ledger Events")
	fmt.Println("")

	for i := len(events) - 1; i >= start; i-- {
		printEvent(events[i])
	}

	return nil
}

func ShowEventRecent(limit int) error {
	events, err := ReadEvents()
	if err != nil {
		return err
	}

	if len(events) == 0 {
		return ShowRecentFromLog(limit)
	}

	if limit <= 0 {
		limit = 10
	}

	start := len(events) - limit
	if start < 0 {
		start = 0
	}

	fmt.Println("Recent installs")
	fmt.Println("")

	for i := len(events) - 1; i >= start; i-- {
		printEvent(events[i])
	}

	return nil
}

func MigrateLogToEvents() error {
	data, err := os.ReadFile(logFile())
	if err != nil {
		fmt.Println("No install-log.md found.")
		return nil
	}

	existingEvents, err := ReadEvents()
	if err != nil {
		return err
	}

	existingIDs := map[string]bool{}
	for _, event := range existingEvents {
		existingIDs[event.ID] = true
	}

	lines := cleanLines(string(data))
	migrated := 0
	skipped := 0

	for _, line := range lines {
		parts := strings.SplitN(line, " | ", 3)
		if len(parts) != 3 {
			skipped++
			continue
		}

		timestampRaw := strings.TrimSpace(parts[0])
		cwd := strings.TrimSpace(parts[1])
		command := strings.TrimSpace(parts[2])

		if !looksLikeInstallCommand(command) {
			skipped++
			continue
		}

		timestamp, err := time.ParseInLocation("2006-01-02 15:04:05", timestampRaw, time.Local)
		if err != nil {
			skipped++
			continue
		}

		event := NewInstallEvent(timestamp, cwd, command)
		if existingIDs[event.ID] {
			skipped++
			continue
		}

		if err := AppendEvent(event); err != nil {
			return err
		}

		existingIDs[event.ID] = true
		migrated++
	}

	fmt.Println("Migration complete.")
	fmt.Println("Migrated:", migrated)
	fmt.Println("Skipped:", skipped)
	fmt.Println("Events file:", eventsFile())

	return nil
}

func printEvent(event InstallEvent) {
	timeLabel := shortTimestamp(event.Timestamp)

	name := event.Name
	if name == "" {
		name = "-"
	}

	source := event.Source
	if source != "" {
		source = " from " + source
	}

	fmt.Printf("%s  %-8s  %-18s  %-24s%s\n",
		timeLabel,
		event.Manager,
		event.Action,
		name,
		source,
	)
}

func shortTimestamp(value string) string {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return value
	}

	return t.Format("2006-01-02 15:04")
}

func eventID(event InstallEvent) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(event.Timestamp))
	_, _ = h.Write([]byte(event.CWD))
	_, _ = h.Write([]byte(event.RawCommand))

	return fmt.Sprintf("%x", h.Sum64())
}

func parseInstallCommand(command string) (manager string, action string, category string, name string, source string) {
	fields := strings.Fields(command)
	lowerFields := normalizedFields(fields)

	manager = "unknown"
	action = "install"
	category = "unknown"

	switch {
	case hasPrefix(lowerFields, "codex", "plugin", "marketplace", "add"):
		manager = "codex"
		action = "marketplace_add"
		category = "marketplace"
		name = argAfter(fields, "add")
		if strings.Contains(name, "/") {
			source = "https://github.com/" + name + ".git"
		}

	case hasPrefix(lowerFields, "codex", "plugin", "add"):
		manager = "codex"
		action = "plugin_add"
		category = "plugin"
		rawName := argAfter(fields, "add")
		if strings.Contains(rawName, "@") {
			parts := strings.SplitN(rawName, "@", 2)
			name = parts[0]
			source = parts[1]
		} else {
			name = rawName
		}

	case hasPrefix(lowerFields, "brew", "install"):
		manager = "brew"
		action = "install"
		category = "package"
		name = firstNonFlagAfter(fields, "install")

	case hasPrefix(lowerFields, "brew", "tap"):
		manager = "brew"
		action = "tap"
		category = "repository"
		name = firstNonFlagAfter(fields, "tap")

	case hasPrefix(lowerFields, "npm", "install", "-g") || hasPrefix(lowerFields, "npm", "i", "-g") || hasPrefix(lowerFields, "npm", "install", "--global"):
		manager = "npm"
		action = "global_install"
		category = "package"
		name = firstNonFlagAfterAny(fields, []string{"install", "i"})

	case hasPrefix(lowerFields, "pnpm", "add", "-g") || hasPrefix(lowerFields, "pnpm", "add", "--global"):
		manager = "pnpm"
		action = "global_install"
		category = "package"
		name = firstNonFlagAfter(fields, "add")

	case hasPrefix(lowerFields, "yarn", "global", "add"):
		manager = "yarn"
		action = "global_install"
		category = "package"
		name = firstNonFlagAfter(fields, "add")

	case hasPrefix(lowerFields, "pipx", "install"):
		manager = "pipx"
		action = "install"
		category = "tool"
		name = firstNonFlagAfter(fields, "install")

	case hasPrefix(lowerFields, "uv", "tool", "install"):
		manager = "uv"
		action = "tool_install"
		category = "tool"
		name = firstNonFlagAfter(fields, "install")

	case hasPrefix(lowerFields, "python", "-m", "pip", "install") || hasPrefix(lowerFields, "python3", "-m", "pip", "install") || hasPrefix(lowerFields, "pip", "install"):
		manager = "pip"
		action = "install"
		category = "package"
		name = firstNonFlagAfter(fields, "install")

	case hasPrefix(lowerFields, "cargo", "install"):
		manager = "cargo"
		action = "install"
		category = "package"
		name = firstNonFlagAfter(fields, "install")

	case hasPrefix(lowerFields, "go", "install"):
		manager = "go"
		action = "install"
		category = "module"
		name = firstNonFlagAfter(fields, "install")

	case hasPrefix(lowerFields, "gem", "install"):
		manager = "gem"
		action = "install"
		category = "package"
		name = firstNonFlagAfter(fields, "install")

	case hasPrefix(lowerFields, "conda", "install"):
		manager = "conda"
		action = "install"
		category = "package"
		name = firstNonFlagAfter(fields, "install")

	case hasPrefix(lowerFields, "code", "--install-extension"):
		manager = "vscode"
		action = "extension_install"
		category = "extension"
		name = firstNonFlagAfter(fields, "--install-extension")
	}

	name = cleanArg(name)
	source = cleanArg(source)

	return manager, action, category, name, source
}

func normalizedFields(fields []string) []string {
	normalized := make([]string, 0, len(fields))
	for _, field := range fields {
		field = cleanArg(strings.ToLower(field))
		if field == "sudo" || field == "command" {
			continue
		}
		normalized = append(normalized, field)
	}

	return normalized
}

func hasPrefix(fields []string, sequence ...string) bool {
	if len(sequence) == 0 || len(sequence) > len(fields) {
		return false
	}

	for i, expected := range sequence {
		if fields[i] != expected {
			return false
		}
	}

	return true
}

func firstNonFlagAfter(fields []string, marker string) string {
	for i, field := range fields {
		if field == marker {
			return firstNonFlag(fields[i+1:])
		}
	}

	return ""
}

func firstNonFlagAfterAny(fields []string, markers []string) string {
	for _, marker := range markers {
		value := firstNonFlagAfter(fields, marker)
		if value != "" {
			return value
		}
	}

	return ""
}

func firstNonFlag(fields []string) string {
	skipNext := false
	for _, field := range fields {
		if skipNext {
			skipNext = false
			continue
		}

		if field == "install" || field == "add" || field == "global" || field == "tool" {
			continue
		}

		if strings.HasPrefix(field, "-") {
			if flagTakesValue(field) {
				skipNext = true
			}
			continue
		}

		return field
	}

	return ""
}

func flagTakesValue(flag string) bool {
	switch flag {
	case "-c", "--channel", "--index-url", "--extra-index-url", "--registry", "--prefix", "--target":
		return true
	default:
		return false
	}
}

func argAfter(fields []string, marker string) string {
	for i, field := range fields {
		if field == marker && i+1 < len(fields) {
			return fields[i+1]
		}
	}

	return ""
}

func cleanArg(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"`)
	value = strings.Trim(value, `'`)
	return value
}
