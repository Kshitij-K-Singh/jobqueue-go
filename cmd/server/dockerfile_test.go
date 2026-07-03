package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDockerfileBuildsServerOnPort8080(t *testing.T) {
	path := filepath.Join("..", "..", "Dockerfile")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read Dockerfile: %v", err)
	}

	dockerfile := string(content)
	for _, want := range []string{
		"./cmd/server",
		"EXPOSE 8080",
		`ENTRYPOINT ["/server"]`,
	} {
		if !strings.Contains(dockerfile, want) {
			t.Fatalf("Dockerfile missing %q", want)
		}
	}
}