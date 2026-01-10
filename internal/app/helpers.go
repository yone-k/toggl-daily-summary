package app

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/yone/toggl-daily-summary/internal/summary"
	"github.com/yone/toggl-daily-summary/internal/toggl"
)

func buildSummaryEntries(entries []toggl.TimeEntry) []summary.Entry {
	out := make([]summary.Entry, 0, len(entries))
	for _, entry := range entries {
		project := strings.TrimSpace(entry.ProjectName)
		if strings.TrimSpace(project) == "" {
			project = "No Project"
		}
		task := entry.Description
		if strings.TrimSpace(task) == "" {
			task = "No Description"
		}
		out = append(out, summary.Entry{
			Project:  project,
			Task:     task,
			Start:    entry.Start,
			Duration: entry.Duration,
		})
	}
	return out
}

func splitEntriesByDay(entries []summary.Entry, loc *time.Location) []summary.Entry {
	if loc == nil {
		loc = time.Local
	}

	out := make([]summary.Entry, 0, len(entries))
	for _, entry := range entries {
		if entry.Duration <= 0 {
			out = append(out, entry)
			continue
		}

		start := entry.Start.In(loc)
		end := start.Add(entry.Duration)
		current := start
		for current.Before(end) {
			dayStart := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, loc)
			nextDay := dayStart.AddDate(0, 0, 1)
			segmentEnd := end
			if nextDay.Before(end) {
				segmentEnd = nextDay
			}
			segmentDuration := segmentEnd.Sub(current)
			if segmentDuration > 0 {
				out = append(out, summary.Entry{
					Project:  entry.Project,
					Task:     entry.Task,
					Start:    current,
					Duration: segmentDuration,
				})
			}
			current = segmentEnd
		}
	}

	return out
}

func writeOutput(outPath, content string, stdout io.Writer) error {
	if outPath == "" {
		_, err := io.WriteString(stdout, content)
		return err
	}
	return os.WriteFile(outPath, []byte(content), 0o644)
}
