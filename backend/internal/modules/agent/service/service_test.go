package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestGatewayClientTaskLifecycleAndAuth(t *testing.T) {
	var seenAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/agents/vm-a":
			_ = json.NewEncoder(w).Encode(agentResponse{Agent: Agent{AgentID: "vm-a", Status: "online"}})
		case "/api/v1/tasks":
			var request createTaskRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("decode create request: %v", err)
			}
			if request.AgentID != "vm-a" || request.Action["type"] != "command.exec" {
				t.Fatalf("unexpected create request: %+v", request)
			}
			_ = json.NewEncoder(w).Encode(taskResponse{Task: Task{TaskID: "task-1", Status: "queued"}})
		case "/api/v1/tasks/task-1":
			_ = json.NewEncoder(w).Encode(taskResponse{Task: Task{TaskID: "task-1", Status: "succeeded"}})
		case "/api/v1/tasks/task-1/cancel":
			_ = json.NewEncoder(w).Encode(taskResponse{Task: Task{TaskID: "task-1", Status: "cancelled"}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	t.Setenv("AGENT_GATEWAY_URL", server.URL)
	t.Setenv("AGENT_GATEWAY_TOKEN", "secret")
	client := NewFromEnv(nil)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	agent, err := client.Agent(ctx, "vm-a")
	if err != nil || agent.Status != "online" {
		t.Fatalf("agent = %+v, err = %v", agent, err)
	}
	task, err := client.CreateTask(ctx, "vm-a", "cms", "idem-1", map[string]any{"type": "command.exec"}, 30, 3)
	if err != nil || task.TaskID != "task-1" {
		t.Fatalf("create task = %+v, err = %v", task, err)
	}
	if task, err = client.Task(ctx, task.TaskID); err != nil || task.Status != "succeeded" {
		t.Fatalf("get task = %+v, err = %v", task, err)
	}
	if task, err = client.Cancel(ctx, task.TaskID); err != nil || task.Status != "cancelled" {
		t.Fatalf("cancel task = %+v, err = %v", task, err)
	}
	if seenAuth != "Bearer secret" {
		t.Fatalf("authorization header = %q", seenAuth)
	}
}

func TestGatewayClientRequiresConfiguration(t *testing.T) {
	os.Unsetenv("AGENT_GATEWAY_URL")
	client := NewFromEnv(nil)
	if _, err := client.Agent(context.Background(), "vm-a"); err == nil {
		t.Fatal("expected missing gateway configuration error")
	}
}

func TestGatewayClientRotatingTokenFile(t *testing.T) {
	var authorization string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(agentResponse{Agent: Agent{AgentID: "vm-a", Status: "online"}})
	}))
	defer server.Close()
	tokenFile := t.TempDir() + "/gateway.token"
	t.Setenv("AGENT_GATEWAY_URL", server.URL)
	t.Setenv("AGENT_GATEWAY_TOKEN", "fallback-must-not-be-used")
	t.Setenv("AGENT_GATEWAY_TOKEN_FILE", tokenFile)
	client := NewFromEnv(nil)
	if _, err := client.Agent(context.Background(), "vm-a"); err == nil {
		t.Fatal("missing token file must fail closed")
	}
	if err := os.WriteFile(tokenFile, []byte("first\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Agent(context.Background(), "vm-a"); err != nil || authorization != "Bearer first" {
		t.Fatalf("first token authorization=%q err=%v", authorization, err)
	}
	if err := os.WriteFile(tokenFile+".next", []byte("second\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(tokenFile+".next", tokenFile); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Agent(context.Background(), "vm-a"); err != nil || authorization != "Bearer second" {
		t.Fatalf("rotated token authorization=%q err=%v", authorization, err)
	}
}
