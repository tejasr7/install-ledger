package ledger

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

func ShowTimeline(days int) error {
	events, err := ReadEvents()
	if err != nil {
		return err
	}

	if len(events) == 0 {
		fmt.Println("No structured events found yet.")
		fmt.Println("Run: il migrate")
		return nil
	}

	if days <= 0 {
		days = 7
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	grouped := groupEventsSince(events, cutoff)

	fmt.Println("Install Timeline")
	fmt.Printf("Last %d days\n", days)
	fmt.Println("")

	if len(grouped) == 0 {
		fmt.Printf("No install events found in the last %d days.\n", days)
		return nil
	}

	dates := sortedDateKeysDesc(grouped)
	for dateIndex, date := range dates {
		if dateIndex > 0 {
			fmt.Println("")
		}

		fmt.Println(date)

		managers := sortedManagerKeys(grouped[date])
		for managerIndex, manager := range managers {
			if managerIndex > 0 {
				fmt.Println("")
			}

			fmt.Printf("  %s\n", manager)
			for _, event := range grouped[date][manager] {
				fmt.Printf("    + %s\n", humanEvent(event))
			}
		}
	}

	return nil
}

func ShowDiff(sinceInput string) error {
	events, err := ReadEvents()
	if err != nil {
		return err
	}

	if len(events) == 0 {
		fmt.Println("No structured events found yet.")
		fmt.Println("Run: il migrate")
		return nil
	}

	cutoff, label, err := parseSinceInput(sinceInput)
	if err != nil {
		return err
	}

	filtered := eventsSince(events, cutoff)
	managerCounts := map[string]int{}
	for _, event := range filtered {
		managerCounts[event.Manager]++
	}

	fmt.Println("Install Ledger Diff")
	fmt.Println("Since:", label)
	fmt.Println("")

	if len(filtered) == 0 {
		fmt.Println("No install events found.")
		return nil
	}

	fmt.Println("Summary")
	for _, manager := range sortedCountKeys(managerCounts) {
		fmt.Printf("- %-8s %d\n", manager+":", managerCounts[manager])
	}

	fmt.Println("")
	fmt.Println("Added")

	for i := len(filtered) - 1; i >= 0; i-- {
		event := filtered[i]
		eventTime, _ := parseEventTime(event)
		fmt.Printf("+ %s  %-8s  %s\n",
			eventTime.Format("2006-01-02 15:04"),
			event.Manager,
			humanEvent(event),
		)
	}

	return nil
}

func parseSinceInput(input string) (time.Time, string, error) {
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		input = "yesterday"
	}

	now := time.Now()

	switch input {
	case "today":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return start, "today", nil
	case "yesterday":
		yesterday := now.AddDate(0, 0, -1)
		start := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, now.Location())
		return start, "yesterday", nil
	case "week":
		return now.AddDate(0, 0, -7), "last 7 days", nil
	case "month":
		return now.AddDate(0, -1, 0), "last month", nil
	}

	if strings.HasSuffix(input, "d") {
		rawDays := strings.TrimSuffix(input, "d")
		days, err := strconv.Atoi(rawDays)
		if err != nil || days <= 0 {
			return time.Time{}, "", fmt.Errorf("invalid duration: %s", input)
		}

		return now.AddDate(0, 0, -days), fmt.Sprintf("last %d days", days), nil
	}

	if parsed, err := time.ParseInLocation("2006-01-02", input, time.Local); err == nil {
		return parsed, input, nil
	}

	return time.Time{}, "", fmt.Errorf("invalid since value: %s\nUse: today, yesterday, week, month, 7d, or YYYY-MM-DD", input)
}

func parseEventTime(event InstallEvent) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339, event.Timestamp)
	if err != nil {
		return time.Time{}, false
	}

	return parsed.Local(), true
}

func humanEvent(event InstallEvent) string {
	name := event.Name
	if name == "" {
		name = "-"
	}

	value := event.Action + " " + name
	if event.Source != "" {
		value += " from " + event.Source
	}

	return value
}

func groupEventsSince(events []InstallEvent, cutoff time.Time) map[string]map[string][]InstallEvent {
	grouped := map[string]map[string][]InstallEvent{}

	for i := len(events) - 1; i >= 0; i-- {
		event := events[i]
		eventTime, ok := parseEventTime(event)
		if !ok || eventTime.Before(cutoff) {
			continue
		}

		date := eventTime.Format("2006-01-02")
		if grouped[date] == nil {
			grouped[date] = map[string][]InstallEvent{}
		}

		grouped[date][event.Manager] = append(grouped[date][event.Manager], event)
	}

	return grouped
}

func eventsSince(events []InstallEvent, cutoff time.Time) []InstallEvent {
	filtered := []InstallEvent{}

	for _, event := range events {
		eventTime, ok := parseEventTime(event)
		if !ok || eventTime.Before(cutoff) {
			continue
		}

		filtered = append(filtered, event)
	}

	return filtered
}

func sortedDateKeysDesc(grouped map[string]map[string][]InstallEvent) []string {
	dates := make([]string, 0, len(grouped))
	for date := range grouped {
		dates = append(dates, date)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(dates)))
	return dates
}

func sortedManagerKeys(grouped map[string][]InstallEvent) []string {
	managers := make([]string, 0, len(grouped))
	for manager := range grouped {
		managers = append(managers, manager)
	}

	sort.Strings(managers)
	return managers
}

func sortedCountKeys(counts map[string]int) []string {
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}
