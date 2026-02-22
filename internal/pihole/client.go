package pihole

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client interacts with a Pi-hole v6 API.
type Client struct {
	host       string
	httpClient *http.Client
}

// NewClient creates a Client for the given Pi-hole host (e.g. "pihole.ucg.sparky.best").
func NewClient(host string) *Client {
	return &Client{
		host: host,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type authRequest struct {
	Password string `json:"password"`
}

type authResponse struct {
	Session struct {
		SID string `json:"sid"`
	} `json:"session"`
}

type blockingRequest struct {
	Blocking bool `json:"blocking"`
	Timer    int  `json:"timer"`
}

// DisableBlocking authenticates, disables ad blocking for the given duration, and logs out.
func (c *Client) DisableBlocking(ctx context.Context, password string, timerSeconds int) error {
	sid, err := c.authenticate(ctx, password)
	if err != nil {
		return fmt.Errorf("authentication: %w", err)
	}
	defer c.logout(ctx, sid)

	if err := c.setBlocking(ctx, sid, timerSeconds); err != nil {
		return fmt.Errorf("disable blocking: %w", err)
	}
	return nil
}

func (c *Client) authenticate(ctx context.Context, password string) (string, error) {
	body, err := json.Marshal(authRequest{Password: password})
	if err != nil {
		return "", fmt.Errorf("marshal auth request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/api/auth"), bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth returned status %d", resp.StatusCode)
	}

	var authResp authResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", fmt.Errorf("decode auth response: %w", err)
	}

	if authResp.Session.SID == "" {
		return "", fmt.Errorf("empty session ID in auth response")
	}

	return authResp.Session.SID, nil
}

func (c *Client) setBlocking(ctx context.Context, sid string, timerSeconds int) error {
	body, err := json.Marshal(blockingRequest{Blocking: false, Timer: timerSeconds})
	if err != nil {
		return fmt.Errorf("marshal blocking request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/api/dns/blocking"), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create blocking request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-FTL-SID", sid)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("blocking request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("blocking returned status %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) logout(ctx context.Context, sid string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url("/api/auth"), nil)
	if err != nil {
		return
	}
	req.Header.Set("X-FTL-SID", sid)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

func (c *Client) url(path string) string {
	return fmt.Sprintf("https://%s%s", c.host, path)
}
