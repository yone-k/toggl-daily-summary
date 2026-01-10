package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/yone/toggl-daily-summary/internal/config"
)

const dateLayout = "2006-01-02"

type DateRange struct {
	Start   time.Time
	End     time.Time
	IsRange bool
}

func Run(ctx context.Context, opts Options) error {
	_ = ctx

	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return err
	}
	config.ApplyEnv(&cfg)
	if opts.WorkspaceID != "" {
		cfg.WorkspaceID = opts.WorkspaceID
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.track.toggl.com/api/v9"
	}

	if cfg.APIToken == "" {
		return errors.New("missing API token: set TOGGL_API_TOKEN or config api_token")
	}
	if cfg.WorkspaceID == "" {
		return errors.New("missing workspace ID: set TOGGL_WORKSPACE_ID or config workspace_id")
	}

	dr, err := resolveDateRange(opts)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Resolved range: %s to %s (daily=%v)\n",
		dr.Start.Format(time.RFC3339),
		dr.End.Format(time.RFC3339),
		opts.Daily,
	)
	fmt.Fprintln(os.Stdout, "Environment OK. API integration will be added next.")
	return nil
}

func resolveDateRange(opts Options) (DateRange, error) {
	if opts.Date != "" && (opts.From != "" || opts.To != "") {
		return DateRange{}, errors.New("use either --date or --from/--to, not both")
	}

	if opts.Date == "" && opts.From == "" && opts.To == "" {
		now := time.Now().In(time.Local)
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		end := start.AddDate(0, 0, 1)
		return DateRange{Start: start, End: end, IsRange: false}, nil
	}

	if opts.Date != "" {
		date, err := time.ParseInLocation(dateLayout, opts.Date, time.Local)
		if err != nil {
			return DateRange{}, fmt.Errorf("invalid --date: %w", err)
		}
		start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
		end := start.AddDate(0, 0, 1)
		return DateRange{Start: start, End: end, IsRange: false}, nil
	}

	if opts.From == "" || opts.To == "" {
		return DateRange{}, errors.New("both --from and --to are required for a range")
	}

	from, err := time.ParseInLocation(dateLayout, opts.From, time.Local)
	if err != nil {
		return DateRange{}, fmt.Errorf("invalid --from: %w", err)
	}
	to, err := time.ParseInLocation(dateLayout, opts.To, time.Local)
	if err != nil {
		return DateRange{}, fmt.Errorf("invalid --to: %w", err)
	}
	if from.After(to) {
		return DateRange{}, errors.New("--from must be <= --to")
	}

	start := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.Local)
	end := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, 1)
	return DateRange{Start: start, End: end, IsRange: true}, nil
}
