package summary

import (
	"testing"
	"time"
)

func TestFormatMarkdownSingleBucket(t *testing.T) {
	buckets := []Bucket{
		{
			Date: "",
			Projects: []ProjectBucket{
				{
					Name:  "Alpha",
					Total: 2*time.Hour + 30*time.Minute,
					Tasks: []TaskBucket{
						{Name: "Build", Total: 30 * time.Minute},
						{Name: "Design", Total: 2 * time.Hour},
					},
				},
				{
					Name:  "Beta",
					Total: 1 * time.Hour,
					Tasks: []TaskBucket{
						{Name: "No Description", Total: 1 * time.Hour},
					},
				},
			},
		},
	}

	got := FormatMarkdown(buckets, FormatOptions{Format: "detail"})
	want := "" +
		"### Alpha 2.50h\n" +
		"- Build 0.50h\n" +
		"- Design 2.00h\n" +
		"\n" +
		"### Beta 1.00h\n" +
		"- No Description 1.00h\n"

	if got != want {
		t.Fatalf("unexpected markdown:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestAggregateDailyBuckets(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	entries := []Entry{
		{
			Project:  "Alpha",
			Task:     "Design",
			Start:    time.Date(2026, 1, 10, 9, 0, 0, 0, jst),
			Duration: 90 * time.Minute,
		},
		{
			Project:  "Alpha",
			Task:     "Build",
			Start:    time.Date(2026, 1, 11, 10, 0, 0, 0, jst),
			Duration: 30 * time.Minute,
		},
	}

	buckets := Aggregate(entries, true, jst)
	if len(buckets) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(buckets))
	}
	if buckets[0].Date != "2026-01-10" || buckets[1].Date != "2026-01-11" {
		t.Fatalf("unexpected dates: %+v", []string{buckets[0].Date, buckets[1].Date})
	}
	if len(buckets[0].Projects) != 1 || buckets[0].Projects[0].Name != "Alpha" {
		t.Fatalf("unexpected project grouping: %+v", buckets[0].Projects)
	}
}

func TestFormatMarkdownDailySpacingAndOrder(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	entries := []Entry{
		{
			Project:  "Beta",
			Task:     "B",
			Start:    time.Date(2026, 1, 10, 10, 0, 0, 0, jst),
			Duration: 30 * time.Minute,
		},
		{
			Project:  "Alpha",
			Task:     "A",
			Start:    time.Date(2026, 1, 10, 9, 0, 0, 0, jst),
			Duration: 90 * time.Minute,
		},
		{
			Project:  "Alpha",
			Task:     "B",
			Start:    time.Date(2026, 1, 10, 11, 0, 0, 0, jst),
			Duration: 30 * time.Minute,
		},
	}

	buckets := Aggregate(entries, true, jst)
	got := FormatMarkdown(buckets, FormatOptions{
		Format:     "detail",
		Daily:      true,
		RangeStart: time.Date(2026, 1, 10, 0, 0, 0, 0, jst),
		RangeEnd:   time.Date(2026, 1, 11, 0, 0, 0, 0, jst),
		Location:   jst,
	})
	want := "" +
		"## 2026-01-10\n" +
		"\n" +
		"### Alpha 2.00h\n" +
		"- A 1.50h\n" +
		"- B 0.50h\n" +
		"\n" +
		"### Beta 0.50h\n" +
		"- B 0.50h\n"

	if got != want {
		t.Fatalf("unexpected markdown:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestFormatMarkdownEmptyRange(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	got := FormatMarkdown(nil, FormatOptions{
		Format:     "detail",
		Daily:      false,
		RangeStart: time.Date(2026, 1, 10, 0, 0, 0, 0, jst),
		RangeEnd:   time.Date(2026, 1, 11, 0, 0, 0, 0, jst),
		Location:   jst,
	})
	want := "## 2026-01-10\n"
	if got != want {
		t.Fatalf("unexpected markdown:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestFormatMarkdownEmptyDailyRange(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	got := FormatMarkdown(nil, FormatOptions{
		Format:     "detail",
		Daily:      true,
		RangeStart: time.Date(2026, 1, 10, 0, 0, 0, 0, jst),
		RangeEnd:   time.Date(2026, 1, 12, 0, 0, 0, 0, jst),
		Location:   jst,
	})
	want := "" +
		"## 2026-01-10\n" +
		"\n" +
		"## 2026-01-11\n"
	if got != want {
		t.Fatalf("unexpected markdown:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestFormatMarkdownDefaultTasksThenProjects(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	entries := []Entry{
		{
			Project:  "Alpha",
			Task:     "Task1",
			Start:    time.Date(2026, 1, 10, 9, 0, 0, 0, jst),
			Duration: 30 * time.Minute,
		},
		{
			Project:  "Alpha",
			Task:     "Task2",
			Start:    time.Date(2026, 1, 10, 10, 0, 0, 0, jst),
			Duration: 60 * time.Minute,
		},
		{
			Project:  "Alpha",
			Task:     "Task1",
			Start:    time.Date(2026, 1, 10, 11, 0, 0, 0, jst),
			Duration: 15 * time.Minute,
		},
		{
			Project:  "Beta",
			Task:     "Task3",
			Start:    time.Date(2026, 1, 10, 12, 0, 0, 0, jst),
			Duration: 30 * time.Minute,
		},
	}

	buckets := Aggregate(entries, false, jst)
	got := FormatMarkdown(buckets, FormatOptions{
		Format:       "default",
		EmptyMessage: "No data",
	})
	want := "" +
		"### タスク\n" +
		"- Task1 0.75h\n" +
		"- Task2 1.00h\n" +
		"- Task3 0.50h\n" +
		"\n" +
		"### プロジェクト\n" +
		"- Alpha 1.75h\n" +
		"- Beta 0.50h\n"

	if got != want {
		t.Fatalf("unexpected markdown:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestFormatMarkdownDefaultDaily(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	entries := []Entry{
		{
			Project:  "Alpha",
			Task:     "Task1",
			Start:    time.Date(2026, 1, 10, 9, 0, 0, 0, jst),
			Duration: 30 * time.Minute,
		},
	}

	buckets := Aggregate(entries, true, jst)
	got := FormatMarkdown(buckets, FormatOptions{
		Format:       "default",
		Daily:        true,
		RangeStart:   time.Date(2026, 1, 10, 0, 0, 0, 0, jst),
		RangeEnd:     time.Date(2026, 1, 11, 0, 0, 0, 0, jst),
		Location:     jst,
		EmptyMessage: "No data",
	})
	want := "" +
		"## 2026-01-10\n" +
		"\n" +
		"### タスク\n" +
		"- Task1 0.50h\n" +
		"\n" +
		"### プロジェクト\n" +
		"- Alpha 0.50h\n"

	if got != want {
		t.Fatalf("unexpected markdown:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestFormatMarkdownDefaultEmpty(t *testing.T) {
	got := FormatMarkdown(nil, FormatOptions{
		Format:       "default",
		EmptyMessage: "No data",
	})
	want := "No data\n"
	if got != want {
		t.Fatalf("unexpected markdown:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
