package toggl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

type TimeEntry struct {
	ID          int64
	Description string
	Start       time.Time
	Duration    time.Duration
	ProjectID   int64
}

type timeEntryResponse struct {
	ID          int64   `json:"id"`
	Description string  `json:"description"`
	Start       string  `json:"start"`
	Duration    int64   `json:"duration"`
	PID         *int64  `json:"pid"`
	ProjectID   *int64  `json:"project_id"`
	TaskID      *int64  `json:"task_id"`
	TID         *int64  `json:"tid"`
	ProjectName *string `json:"project_name"`
}

type projectResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func NewClient(baseURL, token string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		baseURL:    baseURL,
		token:      token,
		httpClient: httpClient,
	}
}

func (c *Client) FetchTimeEntries(ctx context.Context, start, end time.Time) ([]TimeEntry, error) {
	endpoint, err := url.JoinPath(c.baseURL, "me/time_entries")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.token, "api_token")

	q := req.URL.Query()
	q.Set("start_date", start.UTC().Format(time.RFC3339))
	q.Set("end_date", end.UTC().Format(time.RFC3339))
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, buildAPIError(req, resp)
	}

	var raw []timeEntryResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	entries := make([]TimeEntry, 0, len(raw))
	for _, item := range raw {
		if item.Duration < 0 {
			continue
		}
		startTime, err := time.Parse(time.RFC3339, item.Start)
		if err != nil {
			return nil, fmt.Errorf("invalid start time: %w", err)
		}

		projectID := int64(0)
		if item.ProjectID != nil {
			projectID = *item.ProjectID
		} else if item.PID != nil {
			projectID = *item.PID
		}

		entries = append(entries, TimeEntry{
			ID:          item.ID,
			Description: item.Description,
			Start:       startTime,
			Duration:    time.Duration(item.Duration) * time.Second,
			ProjectID:   projectID,
		})
	}

	return entries, nil
}

func (c *Client) FetchProjects(ctx context.Context, workspaceID string) (map[int64]string, error) {
	endpoint, err := url.JoinPath(c.baseURL, "workspaces", workspaceID, "projects")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.token, "api_token")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, buildAPIError(req, resp)
	}

	var raw []projectResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	projects := make(map[int64]string, len(raw))
	for _, item := range raw {
		projects[item.ID] = item.Name
	}

	return projects, nil
}

func buildAPIError(req *http.Request, resp *http.Response) error {
	method := "UNKNOWN"
	uri := ""
	if req != nil {
		method = req.Method
		if req.URL != nil {
			uri = req.URL.RequestURI()
		}
	}
	status := "unknown status"
	if resp != nil {
		status = resp.Status
	}

	body := readErrorBody(resp)
	if body != "" {
		return fmt.Errorf("toggl API error: %s (%s %s): %s", status, method, uri, body)
	}
	return fmt.Errorf("toggl API error: %s (%s %s)", status, method, uri)
}

func readErrorBody(resp *http.Response) string {
	if resp == nil || resp.Body == nil {
		return ""
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 2048))
	if err != nil {
		return ""
	}
	text := strings.TrimSpace(string(data))
	return text
}
