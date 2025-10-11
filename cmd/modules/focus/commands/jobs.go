package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"local/costscope/internal/core/focus/conversion"
	"local/costscope/internal/core/focus/conversion/store"
	"local/costscope/internal/core/focus/types"

	"github.com/spf13/cobra"
)

// sharedConversionManager is a lightweight in-process singleton used by CLI job subcommands.
// The server (API) creates its own manager instance; this is only for direct CLI usage
// when users submit jobs with --submit-only and want to poll without running the API.
var sharedConversionManager *conversion.ConversionManager

func getSharedConversionManager() *conversion.ConversionManager {
	if sharedConversionManager == nil {
		// default max concurrency 4 mirrors registry wiring
		sharedConversionManager = conversion.NewConfiguredConversionManager(4)
	}
	return sharedConversionManager
}

// NOTE: previously exposed a CloseSharedConversionManager helper; removed as unused to reduce deadcode noise.

func NewJobsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "Manage asynchronous FOCUS conversion jobs (CLI local mode)",
	}

	cmd.AddCommand(newJobsListCommand())
	cmd.AddCommand(newJobsStatusCommand())
	cmd.AddCommand(newJobsCancelCommand())
	cmd.AddCommand(newJobsHistoryCommand())
	cmd.AddCommand(newJobsMaintainCommand())

	return cmd
}

// newJobsMaintainCommand adds maintenance operations for Bolt JobStore: prune old entries and compact the DB.
func newJobsMaintainCommand() *cobra.Command {
	var dbPath string
	var removeOlderStr string
	var compact bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "maintain",
		Short: "Run maintenance on the Bolt job store: prune old results and compact DB",
		Long:  "Run maintenance against a Bolt job store file. If --db-path is omitted, the JOB_STORE_PATH environment variable will be used.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// determine DB path
			if dbPath == "" {
				dbPath = os.Getenv("JOB_STORE_PATH")
			}
			if dbPath == "" {
				return fmt.Errorf("no Bolt job store path provided; set --db-path or JOB_STORE_PATH")
			}

			// open BoltJobStore
			js, err := store.NewBoltJobStore(dbPath)
			if err != nil {
				return fmt.Errorf("failed to open Bolt job store at %s: %w", dbPath, err)
			}
			// ensure close and check error
			defer func() {
				if err := js.Close(); err != nil {
					// best-effort log to stderr; avoid failing CLI after operations
					fmt.Fprintf(os.Stderr, "warning: failed to close job store: %v\n", err)
				}
			}()

			// helper to parse removeOlderStr which may be like '30d' or a Go duration
			var cutoff time.Time
			if removeOlderStr != "" {
				d, perr := parseDurationOrDays(removeOlderStr)
				if perr != nil {
					return perr
				}
				cutoff = time.Now().Add(-d)
			}

			// perform dry-run pruning by scanning ListResults
			if removeOlderStr != "" {
				if dryRun {
					results := js.ListResults(0)
					var count int
					for _, r := range results {
						var t time.Time
						if !r.EndTime.IsZero() {
							t = r.EndTime
						} else {
							t = r.StartTime
						}
						if t.Before(cutoff) {
							count++
						}
					}
					if _, werr := fmt.Fprintf(cmd.OutOrStdout(), "Dry-run: %d entries would be removed older than %s\n", count, cutoff.Format(time.RFC3339)); werr != nil {
						return werr
					}
				} else {
					removed, remErr := js.RemoveOlderThan(cutoff)
					if remErr != nil {
						return fmt.Errorf("failed to remove older entries: %w", remErr)
					}
					if _, werr := fmt.Fprintf(cmd.OutOrStdout(), "Removed %d entries older than %s\n", removed, cutoff.Format(time.RFC3339)); werr != nil {
						return werr
					}
				}
			}

			if compact {
				if dryRun {
					if _, werr := fmt.Fprintf(cmd.OutOrStdout(), "Dry-run: would compact Bolt DB at %s\n", dbPath); werr != nil {
						return werr
					}
				} else {
					if cerr := js.Compact(); cerr != nil {
						return fmt.Errorf("failed to compact Bolt DB: %w", cerr)
					}
					if _, werr := fmt.Fprintln(cmd.OutOrStdout(), "Compact completed successfully"); werr != nil {
						return werr
					}
				}
			}

			if removeOlderStr == "" && !compact {
				if _, werr := fmt.Fprintln(cmd.OutOrStdout(), "no operation specified; use --remove-older and/or --compact"); werr != nil {
					return werr
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db-path", "", "Path to Bolt job store file (or set JOB_STORE_PATH)")
	cmd.Flags().StringVar(&removeOlderStr, "remove-older", "", "Remove entries older than the provided duration. Accepts Go duration (e.g. 72h) or days like '30d'.")
	cmd.Flags().BoolVar(&compact, "compact", false, "Compact the Bolt DB to reclaim space")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")
	return cmd
}

// parseDurationOrDays accepts strings like '30d' or standard Go durations and returns time.Duration
func parseDurationOrDays(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	// support suffix 'd' for days
	if strings.HasSuffix(s, "d") || strings.HasSuffix(s, "D") {
		num := strings.TrimSuffix(strings.TrimSuffix(s, "d"), "D")
		// parse integer days
		var days int
		_, err := fmt.Sscanf(num, "%d", &days)
		if err != nil {
			return 0, fmt.Errorf("invalid days duration: %s", s)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	// fallback to time.ParseDuration
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration: %w", err)
	}
	return d, nil
}

func newJobsListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List active conversion jobs (local CLI manager)",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := getSharedConversionManager()
			jobs := mgr.ListActiveJobs()
			if len(jobs) == 0 {
				if _, werr := fmt.Fprintln(cmd.OutOrStdout(), "No active jobs"); werr != nil {
					return werr
				}
				return nil
			}
			for _, j := range jobs {
				if _, werr := fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%v\n", j.ID, j.Status, j.StartTime.Format(time.RFC3339)); werr != nil {
					return werr
				}
			}
			return nil
		},
	}
}

func newJobsStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status <job_id>",
		Short: "Show job status (local CLI manager)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := getSharedConversionManager()
			job, err := mgr.GetJobStatus(args[0])
			if err != nil {
				return err
			}
			if _, werr := fmt.Fprintf(cmd.OutOrStdout(), "ID: %s\nStatus: %s\nStarted: %s\n", job.ID, job.Status, job.StartTime.Format(time.RFC3339)); werr != nil {
				return werr
			}
			if job.EndTime != nil {
				if _, werr := fmt.Fprintf(cmd.OutOrStdout(), "Ended: %s\n", job.EndTime.Format(time.RFC3339)); werr != nil {
					return werr
				}
			}
			if job.Progress != nil && job.Progress.LastError != "" {
				if _, werr := fmt.Fprintf(cmd.OutOrStdout(), "Error: %s\n", job.Progress.LastError); werr != nil {
					return werr
				}
			}
			if job.Result != nil {
				if _, werr := fmt.Fprintf(cmd.OutOrStdout(), "Records: in=%d out=%d Duration=%s Success=%v\n", job.Result.InputRecords, job.Result.OutputRecords, job.Result.Duration, job.Result.Success); werr != nil {
					return werr
				}
			}
			return nil
		},
	}
}

func newJobsCancelCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "cancel <job_id>",
		Short: "Cancel a running job (local CLI manager)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := getSharedConversionManager()
			if err := mgr.CancelJob(args[0]); err != nil {
				return err
			}
			if _, werr := fmt.Fprintln(cmd.OutOrStdout(), "Cancellation requested"); werr != nil {
				return werr
			}
			return nil
		},
	}
}

func newJobsHistoryCommand() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "history",
		Short: "List recently completed jobs (local CLI manager)",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := getSharedConversionManager()
			history := mgr.GetJobHistory(limit)
			if len(history) == 0 {
				if _, werr := fmt.Fprintln(cmd.OutOrStdout(), "No completed jobs"); werr != nil {
					return werr
				}
				return nil
			}
			for _, r := range history {
				if _, werr := fmt.Fprintf(cmd.OutOrStdout(), "%s\t%d/%d\t%v\n", r.OutputFile, r.InputRecords, r.OutputRecords, r.Duration); werr != nil {
					return werr
				}
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of history entries (0 = all)")
	return cmd
}

// Wire into existing focus command tree if not already added.
// ensure we reference types to avoid unused import warnings until deeper integration
var _ types.ConversionStatus
