package summary

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

const dateLayout = "2006-01-02"

type Entry struct {
	Project  string
	Task     string
	Start    time.Time
	Duration time.Duration
}

type TaskBucket struct {
	Name  string
	Total time.Duration
}

type ProjectBucket struct {
	Name  string
	Total time.Duration
	Tasks []TaskBucket
}

type Bucket struct {
	Date     string
	Projects []ProjectBucket
}

func Aggregate(entries []Entry, daily bool, loc *time.Location) []Bucket {
	if loc == nil {
		loc = time.Local
	}

	type taskMap map[string]time.Duration
	type projectMap map[string]taskMap

	grouped := map[string]projectMap{}

	for _, entry := range entries {
		dateKey := ""
		if daily {
			dateKey = entry.Start.In(loc).Format(dateLayout)
		}
		projectName := normalizeProject(entry.Project)
		taskName := normalizeTask(entry.Task)

		if _, ok := grouped[dateKey]; !ok {
			grouped[dateKey] = projectMap{}
		}
		if _, ok := grouped[dateKey][projectName]; !ok {
			grouped[dateKey][projectName] = taskMap{}
		}
		grouped[dateKey][projectName][taskName] += entry.Duration
	}

	dateKeys := make([]string, 0, len(grouped))
	for key := range grouped {
		dateKeys = append(dateKeys, key)
	}
	sort.Strings(dateKeys)

	buckets := make([]Bucket, 0, len(dateKeys))
	for _, dateKey := range dateKeys {
		projects := grouped[dateKey]
		projectNames := make([]string, 0, len(projects))
		for name := range projects {
			projectNames = append(projectNames, name)
		}
		sort.Strings(projectNames)

		projectBuckets := make([]ProjectBucket, 0, len(projectNames))
		for _, projectName := range projectNames {
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

		buckets = append(buckets, Bucket{
			Date:     dateKey,
			Projects: projectBuckets,
		})
	}

	return buckets
}

func FormatMarkdown(buckets []Bucket) string {
	var b strings.Builder
	for _, bucket := range buckets {
		if bucket.Date != "" {
			fmt.Fprintf(&b, "## %s\n", bucket.Date)
		}
		for _, project := range bucket.Projects {
			fmt.Fprintf(&b, "### %s %sh\n", project.Name, formatHours(project.Total))
			for _, task := range project.Tasks {
				fmt.Fprintf(&b, "- %s %sh\n", task.Name, formatHours(task.Total))
			}
		}
	}
	return b.String()
}

func formatHours(d time.Duration) string {
	hours := d.Hours()
	rounded := math.Round(hours*100) / 100
	return fmt.Sprintf("%.2f", rounded)
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
