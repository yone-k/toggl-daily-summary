package app

import (
	"io"
	"os"
	"strings"

	"github.com/yone/toggl-daily-summary/internal/summary"
	"github.com/yone/toggl-daily-summary/internal/toggl"
)

func buildSummaryEntries(entries []toggl.TimeEntry, projects map[int64]string) []summary.Entry {
	out := make([]summary.Entry, 0, len(entries))
	for _, entry := range entries {
		project := projects[entry.ProjectID]
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

func writeOutput(outPath, content string, stdout io.Writer) error {
	if outPath == "" {
		_, err := io.WriteString(stdout, content)
		return err
	}
	return os.WriteFile(outPath, []byte(content), 0o644)
}
