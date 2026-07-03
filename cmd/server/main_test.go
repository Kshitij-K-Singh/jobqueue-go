package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateJobReturnsCreated(t *testing.T) {
	server := NewServer(NewQueue())
	request := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(`{"name":"send email"}`))
	recorder := httptest.NewRecorder()

	server.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusCreated)
	}
}

func TestListJobsReturnsCreatedJobs(t *testing.T) {
	queue := NewQueue()
	queue.CreateJob("first job")
	queue.CreateJob("second job")
	server := NewServer(queue)

	request := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	recorder := httptest.NewRecorder()
	server.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusOK)
	}

	var jobs []Job
	if err := json.NewDecoder(recorder.Body).Decode(&jobs); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("got %d jobs, want 2", len(jobs))
	}
	if jobs[0].Name != "first job" || jobs[1].Name != "second job" {
		t.Fatalf("got job names %q and %q", jobs[0].Name, jobs[1].Name)
	}
}

func TestGetUnknownJobReturnsNotFound(t *testing.T) {
	server := NewServer(NewQueue())
	request := httptest.NewRequest(http.MethodGet, "/jobs/999", nil)
	recorder := httptest.NewRecorder()

	server.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestStatsCountsJobsByStatus(t *testing.T) {
	queue := &Queue{
		nextID: 4,
		jobs: []Job{
			{ID: 1, Name: "first", Status: StatusPending},
			{ID: 2, Name: "second", Status: StatusRunning},
			{ID: 3, Name: "third", Status: StatusDone},
		},
	}
	server := NewServer(queue)

	request := httptest.NewRequest(http.MethodGet, "/stats", nil)
	recorder := httptest.NewRecorder()
	server.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusOK)
	}

	var stats Stats
	if err := json.NewDecoder(recorder.Body).Decode(&stats); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if stats.Pending != 1 || stats.Running != 1 || stats.Done != 1 {
		t.Fatalf("got stats %+v, want one job in each status", stats)
	}
}
