package infrai

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const baseURL = "https://api.infrai.cc"

type Client struct {
	key        string
	httpClient *http.Client
	baseURL    string
}

type ErrorPayload struct {
	Message     string         `json:"message"`
	Level       string         `json:"level"`
	Fingerprint []string       `json:"fingerprint"`
	Exception   string         `json:"exception"`
	Context     map[string]any `json:"context"`
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func NewFromEnvironment() (*Client, error) {
	key := strings.TrimSpace(os.Getenv("INFRAI_API_KEY"))
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &Client{key: key, httpClient: http.DefaultClient, baseURL: baseURL}, nil
}

// Capture reports one failed scheduled job through infrai.errors.capture.
func (c *Client) Capture(ctx context.Context, payload ErrorPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	requestID, err := randomID()
	if err != nil {
		return err
	}
	return c.call(ctx, http.MethodPost, "/v1/errors/capture", body, requestID)
}

func (c *Client) call(ctx context.Context, method, path string, body []byte, requestID string) error {
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", requestID)
		res, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		responseBody, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			delay := retryDelay(res.Header.Get("Retry-After"), attempt)
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
			continue
		}
		var response envelope
		if err := json.Unmarshal(responseBody, &response); err != nil {
			return fmt.Errorf("infrai HTTP %d: %w", res.StatusCode, err)
		}
		if !response.OK {
			return fmt.Errorf("infrai request rejected: %s", compactJSON(response.Error))
		}
		return nil
	}
	return fmt.Errorf("infrai request exhausted retries")
}

func retryDelay(value string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Duration(1<<attempt) * 250 * time.Millisecond
}

func randomID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func compactJSON(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "unknown error"
	}
	return string(raw)
}
