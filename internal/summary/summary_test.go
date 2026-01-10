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

	got := FormatMarkdown(buckets)
	want := "" +
		"### Alpha 2.50h\n" +
		"- Build 0.50h\n" +
		"- Design 2.00h\n" +
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
