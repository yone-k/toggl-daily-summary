package summary

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

const (
	dateLayout = "2006-01-02"
	noTagKey   = ""
)

type Entry struct {
	Tags     []string
	Project  string
	Task     string
	Start    time.Time
	Duration time.Duration
}

type TaskBucket struct {
	Name  string
	Total time.Duration
}

type TaskSummary struct {
	Name       string
	Total      time.Duration
	FirstStart time.Time
}

type ProjectBucket struct {
	Name  string
	Total time.Duration
	Tasks []TaskBucket
}

type Bucket struct {
	Date          string
	Projects      []ProjectBucket
	ProjectsByTag []ProjectBucket
	Tasks         []TaskSummary
}

type tagProjectAgg struct {
	total    time.Duration
	taskKeys map[string]struct{}
}

type tagAgg struct {
	total    time.Duration
	projects map[string]*tagProjectAgg
}

func (a *tagAgg) add(projectName, taskKey string, duration time.Duration) {
	a.total += duration
	project, ok := a.projects[projectName]
	if !ok {
		project = &tagProjectAgg{taskKeys: map[string]struct{}{}}
		a.projects[projectName] = project
	}
	project.total += duration
	project.taskKeys[taskKey] = struct{}{}
}

type FormatOptions struct {
	Daily        bool
	RangeStart   time.Time
	RangeEnd     time.Time
	Location     *time.Location
	Format       string
	EmptyMessage string
}

type AggregateOptions struct {
	Daily                  bool
	Location               *time.Location
	SeparateTasksByProject bool
}

func Aggregate(entries []Entry, opts AggregateOptions) []Bucket {
	loc := opts.Location
	if loc == nil {
		loc = time.Local
	}

	daily := opts.Daily
	separateTasks := opts.SeparateTasksByProject

	type taskMap map[string]time.Duration
	type projectMap map[string]taskMap
	type taskAgg struct {
		total      time.Duration
		firstStart time.Time
	}

	grouped := map[string]projectMap{}
	taskGroups := map[string]map[string]*taskAgg{}
	tagGroups := map[string]map[string]*tagAgg{}

	for _, entry := range entries {
		dateKey := ""
		if daily {
			dateKey = entry.Start.In(loc).Format(dateLayout)
		}
		projectName := normalizeProject(entry.Project)
		taskName := normalizeTask(entry.Task)
		taskKey := buildTaskKey(projectName, taskName, separateTasks)

		if _, ok := grouped[dateKey]; !ok {
			grouped[dateKey] = projectMap{}
		}
		if _, ok := grouped[dateKey][projectName]; !ok {
			grouped[dateKey][projectName] = taskMap{}
		}
		grouped[dateKey][projectName][taskName] += entry.Duration

		if _, ok := taskGroups[dateKey]; !ok {
			taskGroups[dateKey] = map[string]*taskAgg{}
		}
		if agg, ok := taskGroups[dateKey][taskKey]; ok {
			agg.total += entry.Duration
			if entry.Start.Before(agg.firstStart) {
				agg.firstStart = entry.Start
			}
		} else {
			taskGroups[dateKey][taskKey] = &taskAgg{
				total:      entry.Duration,
				firstStart: entry.Start,
			}
		}

		if _, ok := tagGroups[dateKey]; !ok {
			tagGroups[dateKey] = map[string]*tagAgg{}
		}
		tagKey := buildTagKey(entry.Tags)
		if _, ok := tagGroups[dateKey][tagKey]; !ok {
			tagGroups[dateKey][tagKey] = &tagAgg{projects: map[string]*tagProjectAgg{}}
		}
		tagGroups[dateKey][tagKey].add(projectName, taskKey, entry.Duration)
	}

	dateKeys := make([]string, 0, len(grouped))
	for key := range grouped {
		dateKeys = append(dateKeys, key)
	}
	sort.Strings(dateKeys)

	buckets := make([]Bucket, 0, len(dateKeys))
	for _, dateKey := range dateKeys {
		projects := grouped[dateKey]
		projectBuckets := make([]ProjectBucket, 0, len(projects))
		for projectName := range projects {
			tasks := projects[projectName]
			taskNames := make([]string, 0, len(tasks))
			for name := range tasks {
				taskNames = append(taskNames, name)
			}
			sort.Strings(taskNames)

			taskBuckets := make([]TaskBucket, 0, len(taskNames))
			var projectTotal time.Duration
			for _, taskName := range taskNames {
				total := tasks[taskName]
				projectTotal += total
				taskBuckets = append(taskBuckets, TaskBucket{
					Name:  taskName,
					Total: total,
				})
			}

			projectBuckets = append(projectBuckets, ProjectBucket{
				Name:  projectName,
				Total: projectTotal,
				Tasks: taskBuckets,
			})
		}

		sort.Slice(projectBuckets, func(i, j int) bool {
			if projectBuckets[i].Total == projectBuckets[j].Total {
				return projectBuckets[i].Name < projectBuckets[j].Name
			}
			return projectBuckets[i].Total > projectBuckets[j].Total
		})

		projectRanks, taskRanks := rankByTag(tagGroups[dateKey])

		projectsByTag := append([]ProjectBucket(nil), projectBuckets...)
		sort.Slice(projectsByTag, func(i, j int) bool {
			return projectRanks[projectsByTag[i].Name] < projectRanks[projectsByTag[j].Name]
		})

		tasks := taskGroups[dateKey]
		taskSummaries := make([]TaskSummary, 0, len(tasks))
		for name, agg := range tasks {
			taskSummaries = append(taskSummaries, TaskSummary{
				Name:       name,
				Total:      agg.total,
				FirstStart: agg.firstStart,
			})
		}
		sort.Slice(taskSummaries, func(i, j int) bool {
			rankI := taskRanks[taskSummaries[i].Name]
			rankJ := taskRanks[taskSummaries[j].Name]
			if rankI != rankJ {
				return rankI < rankJ
			}
			if taskSummaries[i].FirstStart.Equal(taskSummaries[j].FirstStart) {
				return taskSummaries[i].Name < taskSummaries[j].Name
			}
			return taskSummaries[i].FirstStart.Before(taskSummaries[j].FirstStart)
		})

		buckets = append(buckets, Bucket{
			Date:          dateKey,
			Projects:      projectBuckets,
			ProjectsByTag: projectsByTag,
			Tasks:         taskSummaries,
		})
	}

	return buckets
}

func FormatMarkdown(buckets []Bucket, opts FormatOptions) string {
	var b strings.Builder
	if len(buckets) == 0 {
		return formatEmpty(&b, opts)
	}
	format := normalizeFormat(opts.Format)
	switch format {
	case "detail":
		return formatDetail(&b, buckets)
	default:
		return formatDefault(&b, buckets, opts)
	}
}

func formatEmpty(b *strings.Builder, opts FormatOptions) string {
	format := normalizeFormat(opts.Format)
	if format == "default" {
		msg := strings.TrimSpace(opts.EmptyMessage)
		if msg == "" {
			msg = "No data"
		}
		if opts.Daily && !opts.RangeStart.IsZero() && !opts.RangeEnd.IsZero() {
			return formatDefaultEmptyDaily(b, opts, msg)
		}
		b.WriteString(msg)
		b.WriteString("\n")
		return b.String()
	}

	loc := opts.Location
	if loc == nil {
		loc = time.Local
	}
	if opts.RangeStart.IsZero() || opts.RangeEnd.IsZero() {
		return ""
	}

	start := opts.RangeStart.In(loc)
	endExclusive := opts.RangeEnd.In(loc)
	endInclusive := endExclusive.AddDate(0, 0, -1)
	if endInclusive.Before(start) {
		endInclusive = start
	}

	if opts.Daily {
		for date := start; !date.After(endInclusive); date = date.AddDate(0, 0, 1) {
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			fmt.Fprintf(b, "## %s\n", date.Format(dateLayout))
		}
		return b.String()
	}

	if sameDay(start, endInclusive) {
		fmt.Fprintf(b, "## %s\n", start.Format(dateLayout))
		return b.String()
	}

	fmt.Fprintf(b, "## %s..%s\n", start.Format(dateLayout), endInclusive.Format(dateLayout))
	return b.String()
}

func formatDetail(b *strings.Builder, buckets []Bucket) string {
	emitGap := false
	for _, bucket := range buckets {
		if bucket.Date != "" {
			if emitGap {
				b.WriteString("\n")
			}
			fmt.Fprintf(b, "## %s\n", bucket.Date)
			emitGap = true
		}
		for i, project := range bucket.Projects {
			if i > 0 || bucket.Date != "" {
				b.WriteString("\n")
			}
			fmt.Fprintf(b, "### %s %sh\n", project.Name, formatHours(project.Total))
			for _, task := range project.Tasks {
				fmt.Fprintf(b, "- %s %sh\n", task.Name, formatHours(task.Total))
			}
		}
	}
	return b.String()
}

func formatDefault(b *strings.Builder, buckets []Bucket, opts FormatOptions) string {
	for i, bucket := range buckets {
		if bucket.Date != "" {
			if i > 0 {
				b.WriteString("\n")
			}
			fmt.Fprintf(b, "## %s\n", bucket.Date)
			b.WriteString("\n")
		}

		b.WriteString("### タスク\n")
		for _, task := range bucket.Tasks {
			fmt.Fprintf(b, "- %s %sh\n", task.Name, formatHours(task.Total))
		}

		b.WriteString("\n")

		b.WriteString("### プロジェクト\n")
		for _, project := range bucket.ProjectsByTag {
			fmt.Fprintf(b, "- %s %sh\n", project.Name, formatHours(project.Total))
		}
	}
	return b.String()
}

func formatDefaultEmptyDaily(b *strings.Builder, opts FormatOptions, msg string) string {
	loc := opts.Location
	if loc == nil {
		loc = time.Local
	}
	start := opts.RangeStart.In(loc)
	endExclusive := opts.RangeEnd.In(loc)
	endInclusive := endExclusive.AddDate(0, 0, -1)
	if endInclusive.Before(start) {
		endInclusive = start
	}
	for date := start; !date.After(endInclusive); date = date.AddDate(0, 0, 1) {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(b, "## %s\n", date.Format(dateLayout))
		fmt.Fprintf(b, "%s\n", msg)
	}
	return b.String()
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func formatHours(d time.Duration) string {
	hours := d.Hours()
	rounded := math.Round(hours*100) / 100
	return fmt.Sprintf("%.2f", rounded)
}

func buildTaskKey(projectName, taskName string, separate bool) string {
	if separate {
		return fmt.Sprintf("%s / %s", projectName, taskName)
	}
	return taskName
}

func buildTagKey(tags []string) string {
	names := make([]string, 0, len(tags))
	for _, tag := range tags {
		if name := strings.TrimSpace(tag); name != "" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	// ", " で連結すると、それを名前に含むタグが別の組み合わせと衝突するため使わない。
	return strings.Join(names, "\x00")
}

func rankByTag(tags map[string]*tagAgg) (projectRanks, taskRanks map[string]int) {
	tagKeys := make([]string, 0, len(tags))
	for key := range tags {
		tagKeys = append(tagKeys, key)
	}
	sort.Slice(tagKeys, func(i, j int) bool {
		a, b := tagKeys[i], tagKeys[j]
		if (a == noTagKey) != (b == noTagKey) {
			return b == noTagKey
		}
		if tags[a].total != tags[b].total {
			return tags[a].total > tags[b].total
		}
		return a < b
	})

	projectRanks = map[string]int{}
	taskRanks = map[string]int{}
	rank := 0
	for _, tagKey := range tagKeys {
		projects := tags[tagKey].projects
		names := make([]string, 0, len(projects))
		for name := range projects {
			names = append(names, name)
		}
		sort.Slice(names, func(i, j int) bool {
			if projects[names[i]].total != projects[names[j]].total {
				return projects[names[i]].total > projects[names[j]].total
			}
			return names[i] < names[j]
		})

		for _, name := range names {
			if _, ok := projectRanks[name]; !ok {
				projectRanks[name] = rank
			}
			for taskKey := range projects[name].taskKeys {
				if _, ok := taskRanks[taskKey]; !ok {
					taskRanks[taskKey] = rank
				}
			}
			rank++
		}
	}
	return projectRanks, taskRanks
}

func normalizeProject(name string) string {
	if strings.TrimSpace(name) == "" {
		return "No Project"
	}
	return name
}

func normalizeTask(name string) string {
	if strings.TrimSpace(name) == "" {
		return "No Description"
	}
	return name
}

func normalizeFormat(format string) string {
	switch strings.TrimSpace(strings.ToLower(format)) {
	case "detail":
		return "detail"
	default:
		return "default"
	}
}
