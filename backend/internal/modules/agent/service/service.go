package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

type GatewayClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
	logger     *zap.Logger
}

type Envelope struct {
	ProtocolVersion string    `json:"protocol_version"`
	RequestID       string    `json:"request_id"`
	Timestamp       time.Time `json:"timestamp"`
}
type Agent struct {
	AgentID       string           `json:"agent_id"`
	AgentVersion  string           `json:"agent_version"`
	Hostname      string           `json:"hostname,omitempty"`
	Status        string           `json:"status"`
	RegisteredAt  time.Time        `json:"registered_at"`
	LastHeartbeat time.Time        `json:"last_heartbeat"`
	Capabilities  []map[string]any `json:"capabilities,omitempty"`
}
type Task struct {
	TaskID     string         `json:"task_id"`
	CallerID   string         `json:"caller_id,omitempty"`
	AgentID    string         `json:"agent_id,omitempty"`
	Status     string         `json:"status"`
	Action     map[string]any `json:"action"`
	CreatedAt  time.Time      `json:"created_at"`
	FinishedAt *time.Time     `json:"finished_at,omitempty"`
	Result     *TaskResult    `json:"result,omitempty"`
}
type TaskResult struct {
	Status       string `json:"status"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
	Stdout       string `json:"stdout,omitempty"`
	Stderr       string `json:"stderr,omitempty"`
	ExitCode     int    `json:"exit_code,omitempty"`
}
type taskResponse struct {
	Envelope
	Task Task `json:"task"`
}
type agentResponse struct {
	Envelope
	Agent Agent `json:"agent"`
}
type createTaskRequest struct {
	Envelope
	CallerID       string         `json:"caller_id"`
	IdempotencyKey string         `json:"idempotency_key"`
	AgentID        string         `json:"agent_id"`
	Action         map[string]any `json:"action"`
	TimeoutSeconds int64          `json:"timeout_seconds,omitempty"`
	MaxAttempts    int            `json:"max_attempts,omitempty"`
}

func NewFromEnv(logger *zap.Logger) *GatewayClient {
	return &GatewayClient{baseURL: strings.TrimRight(strings.TrimSpace(os.Getenv("AGENT_GATEWAY_URL")), "/"), token: strings.TrimSpace(os.Getenv("AGENT_GATEWAY_TOKEN")), httpClient: &http.Client{Timeout: 15 * time.Second}, logger: logger}
}
func (c *GatewayClient) configured() error {
	if c.baseURL == "" {
		return errors.New("AGENT_GATEWAY_URL is not configured")
	}
	return nil
}
func (c *GatewayClient) envelope() Envelope {
	return Envelope{ProtocolVersion: "1", RequestID: fmt.Sprintf("cms-%d", time.Now().UnixNano()), Timestamp: time.Now().UTC()}
}
func (c *GatewayClient) Agent(ctx context.Context, agentID string) (Agent, error) {
	var out agentResponse
	err := c.do(ctx, http.MethodGet, "/api/v1/agents/"+url.PathEscape(agentID), nil, &out)
	return out.Agent, err
}
func (c *GatewayClient) CreateTask(ctx context.Context, agentID, caller, idem string, action map[string]any, timeout int64, attempts int) (Task, error) {
	var out taskResponse
	req := createTaskRequest{Envelope: c.envelope(), CallerID: caller, IdempotencyKey: idem, AgentID: agentID, Action: action, TimeoutSeconds: timeout, MaxAttempts: attempts}
	err := c.do(ctx, http.MethodPost, "/api/v1/tasks", req, &out)
	return out.Task, err
}
func (c *GatewayClient) Task(ctx context.Context, id string) (Task, error) {
	var out taskResponse
	err := c.do(ctx, http.MethodGet, "/api/v1/tasks/"+url.PathEscape(id), nil, &out)
	return out.Task, err
}
func (c *GatewayClient) Cancel(ctx context.Context, id string) (Task, error) {
	var out taskResponse
	err := c.do(ctx, http.MethodPost, "/api/v1/tasks/"+url.PathEscape(id)+"/cancel", c.envelope(), &out)
	return out.Task, err
}
func (c *GatewayClient) do(ctx context.Context, method, path string, body any, out any) error {
	if err := c.configured(); err != nil {
		return err
	}
	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
		return fmt.Errorf("agent gateway returned %s: %s", res.Status, string(data))
	}
	if out != nil {
		return json.NewDecoder(io.LimitReader(res.Body, 4<<20)).Decode(out)
	}
	return nil
}
