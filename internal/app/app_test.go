package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yone/toggl-daily-summary/internal/config"
	"github.com/yone/toggl-daily-summary/internal/toggl"
)

func TestBuildSummaryEntries(t *testing.T) {
	entries := []toggl.TimeEntry{
		{
			ID:          1,
			Description: "Design",
			Start:       time.Date(2026, 1, 10, 9, 0, 0, 0, time.UTC),
			Duration:    time.Hour,
			ProjectID:   111,
		},
		{
			ID:          2,
			Description: "Misc",
			Start:       time.Date(2026, 1, 10, 10, 0, 0, 0, time.UTC),
			Duration:    30 * time.Minute,
			ProjectID:   999,
		},
	}
	projects := map[int64]string{
		111: "Alpha",
	}

	got := buildSummaryEntries(entries, projects)
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	if got[0].Project != "Alpha" {
		t.Fatalf("unexpected project: %s", got[0].Project)
	}
	if got[1].Project != "No Project" {
		t.Fatalf("unexpected project for missing id: %s", got[1].Project)
	}
	if got[0].Task != "Design" {
		t.Fatalf("unexpected task: %s", got[0].Task)
	}
	if got[0].Duration != time.Hour {
		t.Fatalf("unexpected duration: %v", got[0].Duration)
	}
}

func TestBuildSummaryEntriesDefaults(t *testing.T) {
	entries := []toggl.TimeEntry{
		{
			ID:          1,
			Description: "",
			Start:       time.Date(2026, 1, 10, 9, 0, 0, 0, time.UTC),
			Duration:    time.Minute,
			ProjectID:   0,
		},
	}

	got := buildSummaryEntries(entries, map[int64]string{})
	if got[0].Project != "No Project" {
		t.Fatalf("unexpected default project: %s", got[0].Project)
	}
	if got[0].Task != "No Description" {
		t.Fatalf("unexpected default task: %s", got[0].Task)
	}
}

func TestWriteOutput(t *testing.T) {
	content := "hello"

	t.Run("stdout", func(t *testing.T) {
		var buf bytes.Buffer
		if err := writeOutput("", content, &buf); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if buf.String() != content {
			t.Fatalf("unexpected stdout: %s", buf.String())
		}
	})

	t.Run("file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "out.md")
		if err := writeOutput(path, content, &bytes.Buffer{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("unexpected read error: %v", err)
		}
		if string(data) != content {
			t.Fatalf("unexpected file content: %s", string(data))
		}
	})
}

type fakeTogglClient struct {
	timeEntries []toggl.TimeEntry
	projects    map[int64]string
}

func (f *fakeTogglClient) FetchTimeEntries(_ context.Context, start, end time.Time) ([]toggl.TimeEntry, error) {
	_ = start
	_ = end
	return f.timeEntries, nil
}

func (f *fakeTogglClient) FetchProjects(_ context.Context, workspaceID string) (map[int64]string, error) {
	_ = workspaceID
	return f.projects, nil
}

func TestRunWritesSummary(t *testing.T) {
	client := &fakeTogglClient{
		timeEntries: []toggl.TimeEntry{
			{
				Description: "Design",
				Start:       time.Date(2026, 1, 10, 9, 0, 0, 0, time.UTC),
				Duration:    90 * time.Minute,
				ProjectID:   111,
			},
			{
				Description: "Build",
				Start:       time.Date(2026, 1, 10, 10, 0, 0, 0, time.UTC),
				Duration:    30 * time.Minute,
				ProjectID:   111,
			},
			{
				Description: "",
				Start:       time.Date(2026, 1, 10, 11, 0, 0, 0, time.UTC),
				Duration:    60 * time.Minute,
				ProjectID:   0,
			},
		},
		projects: map[int64]string{
			111: "Alpha",
		},
	}

	opts := Options{
		Date: "2026-01-10",
	}
	cfg := config.Config{
		APIToken:    "token",
		WorkspaceID: "999",
		BaseURL:     "http://example",
	}

	var buf bytes.Buffer
	err := run(context.Background(), opts, cfg, runDeps{
		client: client,
		stdout: &buf,
		now: func() time.Time {
			return time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "" +
		"- Design 1.50h\n" +
		"- Build 0.50h\n" +
		"- No Description 1.00h\n" +
		"\n" +
		"- Alpha 2.00h\n" +
		"- No Project 1.00h\n"

	if buf.String() != want {
		t.Fatalf("unexpected output:\n--- got ---\n%s\n--- want ---\n%s", buf.String(), want)
	}
}
