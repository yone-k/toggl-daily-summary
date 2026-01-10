package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/yone/toggl-daily-summary/internal/config"
	"github.com/yone/toggl-daily-summary/internal/summary"
	"github.com/yone/toggl-daily-summary/internal/toggl"
)

const dateLayout = "2006-01-02"

type DateRange struct {
	Start   time.Time
	End     time.Time
	IsRange bool
}

func Run(ctx context.Context, opts Options) error {
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

	return run(ctx, opts, cfg, runDeps{
		now:    time.Now,
		stdout: os.Stdout,
	})
}

type TogglClient interface {
	FetchTimeEntries(ctx context.Context, start, end time.Time) ([]toggl.TimeEntry, error)
	FetchProjects(ctx context.Context, workspaceID string) (map[int64]string, error)
}

type runDeps struct {
	client TogglClient
	stdout io.Writer
	now    func() time.Time
}

func run(ctx context.Context, opts Options, cfg config.Config, deps runDeps) error {
	if cfg.APIToken == "" {
		return errors.New("missing API token: set TOGGL_API_TOKEN or config api_token")
	}
	if cfg.WorkspaceID == "" {
		return errors.New("missing workspace ID: set TOGGL_WORKSPACE_ID or config workspace_id")
	}
	if deps.now == nil {
		deps.now = time.Now
	}
	if deps.stdout == nil {
		deps.stdout = os.Stdout
	}
	if deps.client == nil {
		deps.client = toggl.NewClient(cfg.BaseURL, cfg.APIToken, nil)
	}

	dr, err := resolveDateRange(opts, deps.now)
	if err != nil {
		return err
	}

	projects, err := deps.client.FetchProjects(ctx, cfg.WorkspaceID)
	if err != nil {
		return err
	}
	timeEntries, err := deps.client.FetchTimeEntries(ctx, dr.Start, dr.End)
	if err != nil {
		return err
	}

	entries := buildSummaryEntries(timeEntries, projects)
	buckets := summary.Aggregate(entries, opts.Daily, time.Local)
	output := summary.FormatMarkdown(buckets, summary.FormatOptions{
		Daily:      opts.Daily,
		RangeStart: dr.Start,
		RangeEnd:   dr.End,
		Location:   time.Local,
	})

	return writeOutput(opts.Out, output, deps.stdout)
}

func resolveDateRange(opts Options, now func() time.Time) (DateRange, error) {
	if now == nil {
		now = time.Now
	}
	if opts.Date != "" && (opts.From != "" || opts.To != "") {
		return DateRange{}, errors.New("use either --date or --from/--to, not both")
	}

	if opts.Date == "" && opts.From == "" && opts.To == "" {
		current := now().In(time.Local)
		start := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, time.Local)
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
