package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/yone/toggl-daily-summary/internal/app"
)

func Execute() error {
	return NewRootCmd().Execute()
}

func NewRootCmd() *cobra.Command {
	opts := &app.Options{}

	cmd := &cobra.Command{
		Use:   "toggl-daily-summary",
		Short: "Summarize Toggl Track time entries for a date or range",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			return app.Run(ctx, *opts)
		},
	}

	cmd.Flags().StringVar(&opts.Date, "date", "", "Target date in YYYY-M-D (default: today, local)")
	cmd.Flags().StringVar(&opts.From, "from", "", "Start date in YYYY-M-D")
	cmd.Flags().StringVar(&opts.To, "to", "", "End date in YYYY-M-D")
	cmd.Flags().BoolVar(&opts.Daily, "daily", false, "Split output by day when using a date range")
	cmd.Flags().StringVar(&opts.Out, "out", "", "Write output to file (default: stdout)")
	cmd.Flags().StringVar(&opts.ConfigPath, "config", "", "Config file path (default: ~/.config/toggl-daily-summary/config.json)")
	cmd.Flags().StringVar(&opts.WorkspaceID, "workspace", "", "Workspace ID (overrides config/env)")
	cmd.Flags().StringVar(&opts.Format, "format", "default", "Output format: default or detail")

	return cmd
}
