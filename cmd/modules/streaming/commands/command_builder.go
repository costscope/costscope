package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	streamingTypes "local/costscope/cmd/modules/streaming/types"
	"local/costscope/internal/core/logging"
)

// StreamingCommands manages all streaming-related commands
type StreamingCommands struct {
	flags  *streamingTypes.StreamingFlags
	logger *logging.Logger
}

// NewStreamingCommands creates a new instance of StreamingCommands
func NewStreamingCommands() *StreamingCommands {
	return &StreamingCommands{
		flags:  &streamingTypes.StreamingFlags{},
		logger: logging.NewLogger(logging.LevelInfo),
	}
}

// BuildMainCommand creates the main streaming command with all subcommands
func (sc *StreamingCommands) BuildMainCommand() *cobra.Command {
	streamingCmd := &cobra.Command{
		Use:   "streaming",
		Short: "Manage streaming processing jobs",
		Long: `Manage streaming processing jobs with pause, resume, and status commands.

The streaming command provides advanced control over long-running conversion jobs,
allowing you to pause, resume, and monitor large file processing operations.

Features:
  • Start new streaming jobs with custom configuration
  • Pause/Resume processing with checkpoint support
  • Real-time job status and progress monitoring
  • Job history and metrics tracking
  • Automatic crash recovery and resumption
  • Resource usage monitoring
  • Signal handling for graceful shutdown`,
		Example: `  # Start a new streaming job
  costscope streaming start --provider aws --input large_file.csv --output output.parquet

  # List all streaming jobs
  costscope streaming list

  # Check job status
  costscope streaming status --job-id job123

  # Pause a running job
  costscope streaming pause --job-id job123

  # Resume a paused job
  costscope streaming resume --job-id job123`,
		Run: func(cmd *cobra.Command, args []string) {
			sc.logger.Info("Streaming command called")
			_ = cmd.Help()
		},
	}

	// Add subcommands
	streamingCmd.AddCommand(sc.buildStartCommand())
	streamingCmd.AddCommand(sc.buildListCommand())
	streamingCmd.AddCommand(sc.buildStatusCommand())
	streamingCmd.AddCommand(sc.buildPauseCommand())
	streamingCmd.AddCommand(sc.buildResumeCommand())
	streamingCmd.AddCommand(sc.buildStopCommand())
	streamingCmd.AddCommand(sc.buildHistoryCommand())

	return streamingCmd
}

// buildStartCommand creates the start subcommand
func (sc *StreamingCommands) buildStartCommand() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start a new streaming job",
		Long:  "Start a new streaming processing job with specified configuration.",
		Run: func(cmd *cobra.Command, args []string) {
			// Basic provider validation (moved from removed IsValid helpers)
			validProviders := map[string]struct{}{"aws": {}, "azure": {}, "gcp": {}}
			if _, ok := validProviders[sc.flags.Provider]; !ok {
				sc.logger.Error(fmt.Sprintf("invalid provider specified: %s", sc.flags.Provider))
				fmt.Printf(" Invalid provider: %s (allowed: aws, azure, gcp)\n", sc.flags.Provider)
				return
			}

			if sc.flags.Workers <= 0 {
				sc.logger.Error(fmt.Sprintf("workers must be > 0 (got %d)", sc.flags.Workers))
				fmt.Println(" Workers must be greater than 0")
				return
			}

			if sc.flags.Input == "" || sc.flags.Output == "" {
				fmt.Println(" Both --input and --output are required")
				return
			}

			sc.logger.Info(fmt.Sprintf("Starting streaming job provider=%s workers=%d", sc.flags.Provider, sc.flags.Workers))
			fmt.Println(" Streaming job start functionality (coming soon)")
		},
	}

	startCmd.Flags().StringVarP(&sc.flags.Provider, "provider", "p", "aws", "Cloud provider (aws, azure, gcp)")
	startCmd.Flags().StringVarP(&sc.flags.Input, "input", "i", "", "Input file path")
	startCmd.Flags().StringVarP(&sc.flags.Output, "output", "o", "", "Output file path")
	startCmd.Flags().IntVarP(&sc.flags.Workers, "workers", "w", 4, "Number of worker processes")

	return startCmd
}

// buildListCommand creates the list subcommand
func (sc *StreamingCommands) buildListCommand() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all streaming jobs",
		Long:  "List all streaming jobs with their current status and progress.",
		Run: func(cmd *cobra.Command, args []string) {
			sc.logger.Info("Listing streaming jobs")
			fmt.Println(" No streaming jobs found")
		},
	}

	listCmd.Flags().BoolVarP(&sc.flags.Verbose, "verbose", "v", false, "Show detailed information")
	listCmd.Flags().BoolVarP(&sc.flags.JSON, "json", "j", false, "Output in JSON format")

	return listCmd
}

// buildStatusCommand creates the status subcommand
func (sc *StreamingCommands) buildStatusCommand() *cobra.Command {
	return sc.buildJobCommand("status", "Check streaming job status", "Check the status and progress of a specific streaming job.", func() {
		sc.logger.Info("Checking streaming job status")
		fmt.Printf(" Status for job %s: (functionality coming soon)\n", sc.flags.JobID)
	})
}

// buildPauseCommand creates the pause subcommand
func (sc *StreamingCommands) buildPauseCommand() *cobra.Command {
	return sc.buildJobCommand("pause", "Pause a running streaming job", "Pause a currently running streaming job, preserving its state for later resumption.", func() {
		sc.logger.Info("Pausing streaming job")
		fmt.Printf("⏸️  Pausing job %s: (functionality coming soon)\n", sc.flags.JobID)
	})
}

// buildResumeCommand creates the resume subcommand
func (sc *StreamingCommands) buildResumeCommand() *cobra.Command {
	return sc.buildJobCommand("resume", "Resume a paused streaming job", "Resume a previously paused streaming job from its last checkpoint.", func() {
		sc.logger.Info("Resuming streaming job")
		fmt.Printf("▶️  Resuming job %s: (functionality coming soon)\n", sc.flags.JobID)
	})
}

// buildStopCommand creates the stop subcommand
func (sc *StreamingCommands) buildStopCommand() *cobra.Command {
	return sc.buildJobCommand("stop", "Stop a streaming job", "Stop a streaming job permanently. The job cannot be resumed after stopping.", func() {
		sc.logger.Info("Stopping streaming job")
		fmt.Printf(" Stopping job %s: (functionality coming soon)\n", sc.flags.JobID)
	})
}

// buildJobCommand creates a standardized job command that requires job-id
func (sc *StreamingCommands) buildJobCommand(use, short, long string, runFunc func()) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
		Run: func(cmd *cobra.Command, args []string) {
			if sc.flags.JobID == "" {
				fmt.Println(" Job ID is required")
				return
			}
			runFunc()
		},
	}

	cmd.Flags().StringVar(&sc.flags.JobID, "job-id", "", fmt.Sprintf("Job ID to %s", use))
	_ = cmd.MarkFlagRequired("job-id")

	return cmd
}

// buildHistoryCommand creates the history subcommand
func (sc *StreamingCommands) buildHistoryCommand() *cobra.Command {
	historyCmd := &cobra.Command{
		Use:   "history",
		Short: "Show streaming job history",
		Long:  "Show the history of completed, failed, and cancelled streaming jobs.",
		Run: func(cmd *cobra.Command, args []string) {
			sc.logger.Info("Showing streaming job history")
			fmt.Println(" Job history: (functionality coming soon)")
		},
	}

	historyCmd.Flags().IntVar(&sc.flags.Limit, "limit", 10, "Maximum number of jobs to show")
	historyCmd.Flags().BoolVarP(&sc.flags.JSON, "json", "j", false, "Output in JSON format")

	return historyCmd
}
