package toggl

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestClientFetchTimeEntries(t *testing.T) {
	var gotAuth string
	var gotQuery url.Values

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotQuery = r.URL.Query()

		if r.URL.Path != "/api/v9/me/time_entries" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
		  {"id":1,"description":"Design","start":"2026-01-10T09:00:00Z","duration":3600,"pid":111},
		  {"id":2,"description":"Running","start":"2026-01-10T10:00:00Z","duration":-1,"pid":111}
		]`))
	}))
	defer server.Close()

	client := NewClient(server.URL+"/api/v9", "token-123", server.Client())

	start := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 10, 23, 59, 59, 0, time.UTC)
	entries, err := client.FetchTimeEntries(context.Background(), start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry (running excluded), got %d", len(entries))
	}
	if entries[0].Description != "Design" {
		t.Fatalf("unexpected description: %s", entries[0].Description)
	}

	if gotQuery.Get("start_date") == "" || gotQuery.Get("end_date") == "" {
		t.Fatalf("missing date query params: %v", gotQuery.Encode())
	}

	if !strings.HasPrefix(gotAuth, "Basic ") {
		t.Fatalf("missing basic auth header: %s", gotAuth)
	}
	payload, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(gotAuth, "Basic "))
	if err != nil {
		t.Fatalf("invalid auth header: %v", err)
	}
	if string(payload) != "token-123:api_token" {
		t.Fatalf("unexpected auth payload: %s", string(payload))
	}
}

func TestClientFetchProjects(t *testing.T) {
	var gotAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/v9/workspaces/999/projects" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
		  {"id":111,"name":"Alpha"},
		  {"id":222,"name":"Beta"}
		]`))
	}))
	defer server.Close()

	client := NewClient(server.URL+"/api/v9", "token-123", server.Client())
	projects, err := client.FetchProjects(context.Background(), "999")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if projects[111] != "Alpha" || projects[222] != "Beta" {
		t.Fatalf("unexpected projects: %+v", projects)
	}
	if !strings.HasPrefix(gotAuth, "Basic ") {
		t.Fatalf("missing basic auth header: %s", gotAuth)
	}
}

func TestClientFetchTimeEntriesErrorIncludesCause(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("start_date must not be earlier than 2025-10-10"))
	}))
	defer server.Close()

	client := NewClient(server.URL+"/api/v9", "token-123", server.Client())
	start := time.Date(2025, 1, 9, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	_, err := client.FetchTimeEntries(context.Background(), start, end)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "400") {
		t.Fatalf("expected status in error, got: %s", msg)
	}
	if !strings.Contains(msg, "start_date must not be earlier") {
		t.Fatalf("expected cause in error, got: %s", msg)
	}
	if !strings.Contains(msg, "GET") || !strings.Contains(msg, "/api/v9/me/time_entries") {
		t.Fatalf("expected request info in error, got: %s", msg)
	}
}
