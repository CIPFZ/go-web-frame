package browserproxy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/CIPFZ/gowebframe/internal/core/config"
)

const (
	ModeRule   = "rule"
	ModeGlobal = "global"
	ModeCN     = "cn"
	ModeDirect = "direct"
)

type Node struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Scheme    string `json:"scheme"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	Enabled   bool   `json:"enabled"`
	Status    string `json:"status,omitempty"`
	LatencyMs int    `json:"latencyMs,omitempty"`
}

type RuleSet struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Version string   `json:"version"`
	Domains []string `json:"domains"`
}

type Bootstrap struct {
	Enabled       bool      `json:"enabled"`
	Revision      string    `json:"revision"`
	DefaultMode   string    `json:"defaultMode"`
	DefaultNodeID string    `json:"defaultNodeId"`
	Nodes         []Node    `json:"nodes"`
	RuleSets      []RuleSet `json:"ruleSets"`
}

type NodeHealth struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	LatencyMs int    `json:"latencyMs,omitempty"`
}

type Service struct {
	bootstrap Bootstrap
}

func NewService(cfg config.BrowserProxyConfig) (*Service, error) {
	mode := strings.ToLower(strings.TrimSpace(cfg.DefaultMode))
	if mode == "" {
		mode = ModeRule
	}
	if !validMode(mode) {
		return nil, fmt.Errorf("browser proxy default mode %q is invalid", mode)
	}

	seen := make(map[string]struct{}, len(cfg.Nodes))
	nodes := make([]Node, 0, len(cfg.Nodes))
	for _, item := range cfg.Nodes {
		node, err := normalizeNode(item)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[node.ID]; ok {
			return nil, fmt.Errorf("browser proxy node %q is duplicated", node.ID)
		}
		seen[node.ID] = struct{}{}
		nodes = append(nodes, node)
	}
	defaultNodeID := strings.TrimSpace(cfg.DefaultNodeID)
	if defaultNodeID != "" {
		if _, ok := seen[defaultNodeID]; !ok {
			return nil, fmt.Errorf("browser proxy default node %q does not exist", defaultNodeID)
		}
	}

	ruleSets := make([]RuleSet, 0, len(cfg.RuleSets))
	for _, item := range cfg.RuleSets {
		id := strings.TrimSpace(item.ID)
		if id == "" {
			return nil, errors.New("browser proxy rule set id is required")
		}
		ruleSets = append(ruleSets, RuleSet{
			ID: id, Name: strings.TrimSpace(item.Name), Version: strings.TrimSpace(item.Version),
			Domains: normalizeDomains(item.Domains),
		})
	}

	payload := struct {
		Enabled       bool
		DefaultMode   string
		DefaultNodeID string
		Nodes         []Node
		RuleSets      []RuleSet
	}{cfg.Enabled, mode, defaultNodeID, nodes, ruleSets}
	raw, _ := json.Marshal(payload)
	sum := sha256.Sum256(raw)
	return &Service{bootstrap: Bootstrap{
		Enabled: cfg.Enabled, Revision: hex.EncodeToString(sum[:]), DefaultMode: mode,
		DefaultNodeID: defaultNodeID, Nodes: nodes, RuleSets: ruleSets,
	}}, nil
}

func normalizeNode(item config.BrowserProxyNodeConfig) (Node, error) {
	node := Node{
		ID: strings.TrimSpace(item.ID), Name: strings.TrimSpace(item.Name),
		Scheme: strings.ToLower(strings.TrimSpace(item.Scheme)), Host: strings.TrimSpace(item.Host),
		Port: item.Port, Username: item.Username, Password: item.Password, Enabled: item.Enabled,
		Status: strings.TrimSpace(item.Status), LatencyMs: item.LatencyMs,
	}
	if node.ID == "" || node.Name == "" {
		return Node{}, errors.New("browser proxy node id and name are required")
	}
	if node.Scheme != "http" && node.Scheme != "https" && node.Scheme != "socks5" {
		return Node{}, fmt.Errorf("browser proxy node %q has unsupported scheme %q", node.ID, node.Scheme)
	}
	if node.Host == "" || node.Port < 1 || node.Port > 65535 {
		return Node{}, fmt.Errorf("browser proxy node %q has invalid endpoint", node.ID)
	}
	if node.Status == "" {
		node.Status = "unknown"
	}
	return node, nil
}

func normalizeDomains(domains []string) []string {
	out := make([]string, 0, len(domains))
	seen := make(map[string]struct{}, len(domains))
	for _, domain := range domains {
		domain = strings.ToLower(strings.TrimSpace(domain))
		domain = strings.TrimPrefix(domain, "*.")
		domain = strings.TrimPrefix(domain, ".")
		if domain == "" {
			continue
		}
		if _, ok := seen[domain]; ok {
			continue
		}
		seen[domain] = struct{}{}
		out = append(out, domain)
	}
	return out
}

func validMode(mode string) bool {
	return mode == ModeRule || mode == ModeGlobal || mode == ModeCN || mode == ModeDirect
}

func (s *Service) Bootstrap(context.Context) Bootstrap {
	return s.bootstrap
}

func (s *Service) Health(ctx context.Context) []NodeHealth {
	items := make([]NodeHealth, 0, len(s.bootstrap.Nodes))
	for _, node := range s.bootstrap.Nodes {
		health := NodeHealth{ID: node.ID, Status: node.Status, LatencyMs: node.LatencyMs}
		if !node.Enabled {
			health.Status = "disabled"
			items = append(items, health)
			continue
		}
		start := time.Now()
		dialer := net.Dialer{Timeout: 2 * time.Second}
		conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", node.Host, node.Port))
		if err != nil {
			health.Status = "offline"
			health.LatencyMs = 0
		} else {
			_ = conn.Close()
			health.Status = "ready"
			health.LatencyMs = int(time.Since(start).Milliseconds())
		}
		items = append(items, health)
	}
	return items
}
