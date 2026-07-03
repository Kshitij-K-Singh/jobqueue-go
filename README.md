# JobQueue Go

This is a tiny Go HTTP server that demonstrates a simple in-memory job queue.

It is meant for learning. Jobs are stored in a Go slice protected by a mutex, and a background goroutine processes one pending job at a time. There is no database, frontend, Docker setup, retry system, or third-party router.

Jobs disappear when the server stops.

## Run

```bash
go run ./cmd/server
```

The server listens on `:8080` by default.

Use a different port with:

```bash
go run ./cmd/server -addr :3000
```

## API Examples

Create a job:

```bash
curl -i -X POST http://localhost:8080/jobs \
  -H 'Content-Type: application/json' \
  -d '{"name":"send welcome email"}'
```

List all jobs:

```bash
curl http://localhost:8080/jobs
```

Get one job:

```bash
curl http://localhost:8080/jobs/1
```

Get queue stats:

```bash
curl http://localhost:8080/stats
```

## Job Statuses

Each job starts as `pending`.

The background processor changes one pending job to `running`, waits for two seconds to simulate work, and then changes it to `done`.

## Project Layout

```text
cmd/server/main.go       Server, handlers, queue, and processor
cmd/server/main_test.go  Small HTTP handler tests
go.mod                  Module name and Go version
README.md               This guide
```

## Tests

```bash
GOCACHE=/tmp/jobqueue-go-cache go test ./...
```
