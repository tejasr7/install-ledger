package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tejasr7/install-ledger/internal/ledger"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "il",
		Short: "Install Ledger - your developer machine install timeline",
		Long:  "Install Ledger tracks install commands, scans installed tools, and shows a searchable timeline.",
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Initialize Install Ledger tracking",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ledger.Init()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "scan",
		Short: "Scan installed tools and save inventory",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ledger.Scan()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "log",
		Short: "Show install timeline",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ledger.ShowLog()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "today",
		Short: "Show installs tracked today",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ledger.ShowToday()
		},
	})

	timelineCmd := &cobra.Command{
		Use:   "timeline",
		Short: "Show grouped install timeline",
		RunE: func(cmd *cobra.Command, args []string) error {
			days, _ := cmd.Flags().GetInt("days")
			return ledger.ShowTimeline(days)
		},
	}

	timelineCmd.Flags().IntP("days", "d", 7, "Number of days to show")
	rootCmd.AddCommand(timelineCmd)

	diffCmd := &cobra.Command{
		Use:   "diff [since]",
		Short: "Show install events since a time period",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			since := "yesterday"
			if len(args) > 0 {
				since = args[0]
			}

			return ledger.ShowDiff(since)
		},
	}

	rootCmd.AddCommand(diffCmd)

	recentCmd := &cobra.Command{
		Use:   "recent",
		Short: "Show recent install events",
		RunE: func(cmd *cobra.Command, args []string) error {
			limit, _ := cmd.Flags().GetInt("limit")
			return ledger.ShowRecent(limit)
		},
	}

	recentCmd.Flags().IntP("limit", "n", 10, "Number of recent events to show")
	rootCmd.AddCommand(recentCmd)

	eventsCmd := &cobra.Command{
		Use:   "events",
		Short: "Show structured install events",
		RunE: func(cmd *cobra.Command, args []string) error {
			limit, _ := cmd.Flags().GetInt("limit")
			return ledger.ShowEvents(limit)
		},
	}

	eventsCmd.Flags().IntP("limit", "n", 20, "Number of events to show")
	rootCmd.AddCommand(eventsCmd)

	rootCmd.AddCommand(&cobra.Command{
		Use:   "migrate",
		Short: "Migrate old install-log.md entries into structured events",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ledger.MigrateLogToEvents()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "doctor",
		Short: "Check Install Ledger setup health",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ledger.Doctor()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Show Install Ledger data paths",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ledger.ShowPath()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "scan-summary",
		Short: "Show a clean summary of the latest inventory scan",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ledger.ScanSummary()
		},
	})

	exportCmd := &cobra.Command{
		Use:   "export [markdown|json|brewfile]",
		Short: "Export install history and inventory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			outPath, _ := cmd.Flags().GetString("out")
			return ledger.Export(args[0], outPath)
		},
	}

	exportCmd.Flags().StringP("out", "o", "", "Output file path")
	rootCmd.AddCommand(exportCmd)

	rootCmd.AddCommand(&cobra.Command{
		Use:   "find <query>",
		Short: "Search install history",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ledger.Find(args[0])
		},
	})

	captureCmd := &cobra.Command{
		Use:    "capture <cwd> <command>",
		Short:  "Capture an install command from shell hook",
		Hidden: true,
		Args:   cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ledger.Capture(args[0], args[1])
		},
	}

	rootCmd.AddCommand(captureCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
