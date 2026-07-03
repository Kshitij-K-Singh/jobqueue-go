package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	StatusPending = "pending"
	StatusRunning = "running"
	StatusDone    = "done"
)

type Job struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

type Stats struct {
	Pending int `json:"pending"`
	Running int `json:"running"`
	Done    int `json:"done"`
}

type Queue struct {
	mu     sync.Mutex
	nextID int
	jobs   []Job
}

func NewQueue() *Queue {
	return &Queue{nextID: 1}
}

func (q *Queue) CreateJob(name string) Job {
	q.mu.Lock()
	defer q.mu.Unlock()

	job := Job{
		ID:        q.nextID,
		Name:      name,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}

	q.nextID++
	q.jobs = append(q.jobs, job)
	return job
}

func (q *Queue) ListJobs() []Job {
	q.mu.Lock()
	defer q.mu.Unlock()

	jobs := make([]Job, len(q.jobs))
	copy(jobs, q.jobs)
	return jobs
}

func (q *Queue) GetJob(id int) (Job, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, job := range q.jobs {
		if job.ID == id {
			return job, true
		}
	}

	return Job{}, false
}

func (q *Queue) Stats() Stats {
	q.mu.Lock()
	defer q.mu.Unlock()

	var stats Stats
	for _, job := range q.jobs {
		switch job.Status {
		case StatusPending:
			stats.Pending++
		case StatusRunning:
			stats.Running++
		case StatusDone:
			stats.Done++
		}
	}

	return stats
}

func (q *Queue) StartNextJob() (int, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i := range q.jobs {
		if q.jobs[i].Status == StatusPending {
			q.jobs[i].Status = StatusRunning
			return q.jobs[i].ID, true
		}
	}

	return 0, false
}

func (q *Queue) FinishJob(id int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i := range q.jobs {
		if q.jobs[i].ID == id {
			now := time.Now()
			q.jobs[i].Status = StatusDone
			q.jobs[i].FinishedAt = &now
			return
		}
	}
}

type Server struct {
	queue *Queue
}

func NewServer(queue *Queue) *Server {
	return &Server{queue: queue}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/jobs", s.handleJobs)
	mux.HandleFunc("/jobs/", s.handleGetJob)
	mux.HandleFunc("/stats", s.handleStats)
	return mux
}

func (s *Server) handleJobs(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/jobs" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodPost:
		s.handleCreateJob(w, r)
	case http.MethodGet:
		s.handleListJobs(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleCreateJob(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	request.Name = strings.TrimSpace(request.Name)
	if request.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusCreated, s.queue.CreateJob(request.Name))
}

func (s *Server) handleListJobs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.queue.ListJobs())
}

func (s *Server) handleGetJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idText := strings.TrimPrefix(r.URL.Path, "/jobs/")
	if idText == "" || strings.Contains(idText, "/") {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.Atoi(idText)
	if err != nil {
		http.Error(w, "invalid job id", http.StatusBadRequest)
		return
	}

	job, ok := s.queue.GetJob(id)
	if !ok {
		http.NotFound(w, r)
		return
	}

	writeJSON(w, http.StatusOK, job)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, s.queue.Stats())
}

func processJobs(queue *Queue) {
	for {
		jobID, ok := queue.StartNextJob()
		if !ok {
			time.Sleep(250 * time.Millisecond)
			continue
		}

		time.Sleep(2 * time.Second)
		queue.FinishJob(jobID)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()

	queue := NewQueue()
	server := NewServer(queue)

	go processJobs(queue)

	log.Printf("listening on %s", *addr)
	if err := http.ListenAndServe(*addr, server.Routes()); err != nil {
		log.Fatal(err)
	}
}
