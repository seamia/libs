package post

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	timeoutPostJSON = 30 * time.Second
)

func PostJSON(ctx context.Context, baseURL, path string, data interface{}, headers map[string]string) error {
	// Marshal data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Determine transport type and construct URL
	var client *http.Client
	var reqURL string

	if strings.HasPrefix(baseURL, "unix:") || strings.HasPrefix(baseURL, "uds:") {
		// Extract socket path (remove prefix)
		socketPath := strings.TrimPrefix(baseURL, "unix:")
		socketPath = strings.TrimPrefix(socketPath, "uds:")

		// Use http://localhost as dummy URL for UDS
		reqURL = "http://localhost" + path

		// Create client with Unix socket transport
		client = &http.Client{
			Timeout: timeoutPostJSON,
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					return net.Dial("unix", socketPath)
				},
			},
		}
	} else {
		// Standard HTTP/HTTPS
		reqURL = baseURL + path
		client = &http.Client{
			Timeout: timeoutPostJSON,
		}
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set default Content-Type header
	req.Header.Set("Content-Type", "application/json")

	// Add/replace custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
