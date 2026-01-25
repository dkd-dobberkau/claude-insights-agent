package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dkd/claude-insights-agent/internal/parser"
)

// Client handles communication with the insights server
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// New creates a new API client
func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SessionResponse is the server response for session upload
type SessionResponse struct {
	Status    string   `json:"status"`
	SessionID string   `json:"session_id"`
	Warnings  []string `json:"warnings"`
}

// PlanResponse is the server response for plan upload
type PlanResponse struct {
	Status   string   `json:"status"`
	Name     string   `json:"name"`
	Warnings []string `json:"warnings"`
}

// Upload sends a session to the server
func (c *Client) Upload(s *parser.Session) (*SessionResponse, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("marshal session: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/sessions", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server error %d: %s", resp.StatusCode, string(body))
	}

	var result SessionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &result, nil
}

// UploadBatch sends multiple sessions to the server
func (c *Client) UploadBatch(sessions []*parser.Session) ([]*SessionResponse, error) {
	data, err := json.Marshal(sessions)
	if err != nil {
		return nil, fmt.Errorf("marshal sessions: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/sessions/batch", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server error %d: %s", resp.StatusCode, string(body))
	}

	var results []*SessionResponse
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return results, nil
}

// UploadPlan sends a plan to the server
func (c *Client) UploadPlan(p *parser.Plan) (*PlanResponse, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("marshal plan: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/plans", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server error %d: %s", resp.StatusCode, string(body))
	}

	var result PlanResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &result, nil
}

// UploadPlanBatch sends multiple plans to the server
func (c *Client) UploadPlanBatch(plans []*parser.Plan) ([]*PlanResponse, error) {
	data, err := json.Marshal(plans)
	if err != nil {
		return nil, fmt.Errorf("marshal plans: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/plans/batch", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server error %d: %s", resp.StatusCode, string(body))
	}

	var results []*PlanResponse
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return results, nil
}

// Health checks if the server is reachable
func (c *Client) Health() error {
	req, err := http.NewRequest("GET", c.baseURL+"/health", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("server unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server unhealthy: status %d", resp.StatusCode)
	}

	return nil
}

// Errors
var (
	ErrUnauthorized = fmt.Errorf("unauthorized: invalid API key")
)
