// Command healthcheck verifies that the application and its PostgreSQL
// dependency are ready to serve traffic.
package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	defaultPort  = "8000"
	probePath    = "/api/v1/health/ready"
	probeTimeout = 3 * time.Second
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	endpoint := (&url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort("127.0.0.1", port),
		Path:   probePath,
	}).String()

	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		fail(err)
	}
	client := &http.Client{
		Timeout: probeTimeout,
		Transport: &http.Transport{
			Proxy: nil,
		},
	}
	response, err := client.Do(request)
	if err != nil {
		fail(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		fail(fmt.Errorf("readiness endpoint returned %s", response.Status))
	}
}

func fail(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "readiness probe failed: %v\n", err)
	os.Exit(1)
}
