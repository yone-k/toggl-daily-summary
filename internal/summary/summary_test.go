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

	buckets := Aggregate(entries, AggregateOptions{
		Daily:    true,
		Location: jst,
	})
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

	buckets := Aggregate(entries, AggregateOptions{
		Daily:    true,
		Location: jst,
	})
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

	buckets := Aggregate(entries, AggregateOptions{
		Daily:    false,
		Location: jst,
	})
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

func TestAggregateSeparateTasksByProject(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	entries := []Entry{
		{
			Project:  "Alpha",
			Task:     "Task1",
			Start:    time.Date(2026, 1, 10, 9, 0, 0, 0, jst),
			Duration: 30 * time.Minute,
		},
		{
			Project:  "Beta",
			Task:     "Task1",
			Start:    time.Date(2026, 1, 10, 10, 0, 0, 0, jst),
			Duration: 60 * time.Minute,
		},
	}

	buckets := Aggregate(entries, AggregateOptions{
		Daily:                  false,
		Location:               jst,
		SeparateTasksByProject: true,
	})
	if len(buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(buckets))
	}
	if len(buckets[0].Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(buckets[0].Tasks))
	}
	if buckets[0].Tasks[0].Name != "Beta / Task1" {
		t.Fatalf("unexpected first task name: %s", buckets[0].Tasks[0].Name)
	}
	if buckets[0].Tasks[1].Name != "Alpha / Task1" {
		t.Fatalf("unexpected second task name: %s", buckets[0].Tasks[1].Name)
	}
}

func TestFormatMarkdownDefaultTasksGroupedByProjectOrder(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	entries := []Entry{
		{
			Project:  "Beta",
			Task:     "B",
			Start:    time.Date(2026, 1, 10, 9, 0, 0, 0, jst),
			Duration: 30 * time.Minute,
		},
		{
			Project:  "Alpha",
			Task:     "A",
			Start:    time.Date(2026, 1, 10, 10, 0, 0, 0, jst),
			Duration: 60 * time.Minute,
		},
		{
			Project:  "Alpha",
			Task:     "C",
			Start:    time.Date(2026, 1, 10, 11, 0, 0, 0, jst),
			Duration: 30 * time.Minute,
		},
		{
			Project:  "Beta",
			Task:     "C",
			Start:    time.Date(2026, 1, 10, 12, 0, 0, 0, jst),
			Duration: 15 * time.Minute,
		},
		{
			Project:  "Alpha",
			Task:     "A",
			Start:    time.Date(2026, 1, 10, 13, 0, 0, 0, jst),
			Duration: 30 * time.Minute,
		},
	}

	buckets := Aggregate(entries, AggregateOptions{
		Daily:    false,
		Location: jst,
	})
	got := FormatMarkdown(buckets, FormatOptions{
		Format:       "default",
		EmptyMessage: "No data",
	})
	want := "" +
		"### タスク\n" +
		"- A 1.50h\n" +
		"- C 0.75h\n" +
		"- B 0.50h\n" +
		"\n" +
		"### プロジェクト\n" +
		"- Alpha 2.00h\n" +
		"- Beta 0.75h\n"

	if got != want {
		t.Fatalf("unexpected markdown:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestFormatMarkdownDefaultOrdersByTag(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	at := func(hour int) time.Time {
		return time.Date(2026, 1, 10, hour, 0, 0, 0, jst)
	}

	tests := []struct {
		name    string
		entries []Entry
		want    string
	}{
		{
			name: "タグの合計時間の降順に並び、タグ内はプロジェクトの合計時間の降順に並ぶ",
			entries: []Entry{
				{Tags: []string{"開発"}, Project: "Alpha", Task: "A", Start: at(9), Duration: 2 * time.Hour},
				{Tags: []string{"開発"}, Project: "Beta", Task: "B", Start: at(10), Duration: 3 * time.Hour},
				{Tags: []string{"会議"}, Project: "Gamma", Task: "C", Start: at(11), Duration: 4 * time.Hour},
			},
			want: "" +
				"### タスク\n" +
				"- B 3.00h\n" +
				"- A 2.00h\n" +
				"- C 4.00h\n" +
				"\n" +
				"### プロジェクト\n" +
				"- Beta 3.00h\n" +
				"- Alpha 2.00h\n" +
				"- Gamma 4.00h\n",
		},
		{
			name: "タグなしは合計時間が最大でも最下部に並ぶ",
			entries: []Entry{
				{Project: "Alpha", Task: "A", Start: at(9), Duration: 5 * time.Hour},
				{Tags: []string{"開発"}, Project: "Beta", Task: "B", Start: at(10), Duration: time.Hour},
			},
			want: "" +
				"### タスク\n" +
				"- B 1.00h\n" +
				"- A 5.00h\n" +
				"\n" +
				"### プロジェクト\n" +
				"- Beta 1.00h\n" +
				"- Alpha 5.00h\n",
		},
		{
			name: "空白のみのタグはタグなしとして扱う",
			entries: []Entry{
				{Tags: []string{" "}, Project: "Alpha", Task: "A", Start: at(9), Duration: 5 * time.Hour},
				{Tags: []string{"開発"}, Project: "Beta", Task: "B", Start: at(10), Duration: time.Hour},
				{Project: "Gamma", Task: "C", Start: at(11), Duration: 2 * time.Hour},
			},
			want: "" +
				"### タスク\n" +
				"- B 1.00h\n" +
				"- A 5.00h\n" +
				"- C 2.00h\n" +
				"\n" +
				"### プロジェクト\n" +
				"- Beta 1.00h\n" +
				"- Alpha 5.00h\n" +
				"- Gamma 2.00h\n",
		},
		{
			name: "同時間のタグは名前順に並ぶ",
			entries: []Entry{
				{Tags: []string{"b"}, Project: "Alpha", Task: "A", Start: at(9), Duration: time.Hour},
				{Tags: []string{"a"}, Project: "Beta", Task: "B", Start: at(10), Duration: time.Hour},
			},
			want: "" +
				"### タスク\n" +
				"- B 1.00h\n" +
				"- A 1.00h\n" +
				"\n" +
				"### プロジェクト\n" +
				"- Beta 1.00h\n" +
				"- Alpha 1.00h\n",
		},
		{
			name: "複数タグは付与順によらず組み合わせごとに1グループになる",
			entries: []Entry{
				{Tags: []string{"a", "b"}, Project: "X", Task: "T1", Start: at(9), Duration: time.Hour},
				{Tags: []string{"b", "a"}, Project: "Y", Task: "T2", Start: at(10), Duration: time.Hour},
				{Tags: []string{"a"}, Project: "Z", Task: "T3", Start: at(11), Duration: 90 * time.Minute},
			},
			want: "" +
				"### タスク\n" +
				"- T1 1.00h\n" +
				"- T2 1.00h\n" +
				"- T3 1.50h\n" +
				"\n" +
				"### プロジェクト\n" +
				"- X 1.00h\n" +
				"- Y 1.00h\n" +
				"- Z 1.50h\n",
		},
		{
			name: "タグをまたぐプロジェクトと同名タスクは先に並ぶタグの位置に置き、全体の合計を表示する",
			entries: []Entry{
				{Tags: []string{"x"}, Project: "Beta", Task: "Shared", Start: at(9), Duration: 3 * time.Hour},
				{Tags: []string{"y"}, Project: "Alpha", Task: "Own", Start: at(10), Duration: 150 * time.Minute},
				{Project: "Alpha", Task: "Shared", Start: at(11), Duration: 5 * time.Hour},
			},
			want: "" +
				"### タスク\n" +
				"- Shared 8.00h\n" +
				"- Own 2.50h\n" +
				"\n" +
				"### プロジェクト\n" +
				"- Beta 3.00h\n" +
				"- Alpha 7.50h\n",
		},
		{
			name: "同じタグとプロジェクト内のタスクは最初の開始時刻順に並ぶ",
			entries: []Entry{
				{Tags: []string{"開発"}, Project: "Alpha", Task: "A", Start: at(10), Duration: time.Hour},
				{Tags: []string{"開発"}, Project: "Alpha", Task: "Z", Start: at(9), Duration: time.Hour},
			},
			want: "" +
				"### タスク\n" +
				"- Z 1.00h\n" +
				"- A 1.00h\n" +
				"\n" +
				"### プロジェクト\n" +
				"- Alpha 2.00h\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buckets := Aggregate(tt.entries, AggregateOptions{Location: jst})
			got := FormatMarkdown(buckets, FormatOptions{
				Format:       "default",
				EmptyMessage: "No data",
			})
			if got != tt.want {
				t.Fatalf("unexpected markdown:\n--- got ---\n%s\n--- want ---\n%s", got, tt.want)
			}
		})
	}
}

func TestAggregateSeparateTasksByProjectOrdersByTag(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	start := time.Date(2026, 1, 10, 9, 0, 0, 0, jst)
	entries := []Entry{
		{Tags: []string{"開発"}, Project: "Alpha", Task: "T", Start: start, Duration: time.Hour},
		{Tags: []string{"会議"}, Project: "Beta", Task: "T", Start: start, Duration: 2 * time.Hour},
		{Project: "Gamma", Task: "T", Start: start, Duration: 5 * time.Hour},
	}

	buckets := Aggregate(entries, AggregateOptions{
		Location:               jst,
		SeparateTasksByProject: true,
	})

	want := []string{"Beta / T", "Alpha / T", "Gamma / T"}
	if len(buckets) != 1 || len(buckets[0].Tasks) != len(want) {
		t.Fatalf("unexpected buckets: %+v", buckets)
	}
	for i, name := range want {
		if buckets[0].Tasks[i].Name != name {
			t.Fatalf("task %d: got %s, want %s", i, buckets[0].Tasks[i].Name, name)
		}
	}
}

func TestFormatMarkdownDetailKeepsProjectTotalOrderWithTags(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	start := time.Date(2026, 1, 10, 9, 0, 0, 0, jst)
	entries := []Entry{
		{Tags: []string{"開発"}, Project: "Beta", Task: "B", Start: start, Duration: time.Hour},
		{Project: "Alpha", Task: "A", Start: start, Duration: 5 * time.Hour},
	}

	buckets := Aggregate(entries, AggregateOptions{Location: jst})
	got := FormatMarkdown(buckets, FormatOptions{Format: "detail"})
	want := "" +
		"### Alpha 5.00h\n" +
		"- A 5.00h\n" +
		"\n" +
		"### Beta 1.00h\n" +
		"- B 1.00h\n"

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

	buckets := Aggregate(entries, AggregateOptions{
		Daily:    true,
		Location: jst,
	})
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
