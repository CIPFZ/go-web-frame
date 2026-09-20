package browserproxy

import (
	"testing"

	"github.com/CIPFZ/gowebframe/internal/core/config"
)

func TestNewServiceNormalizesAndRevisionsConfig(t *testing.T) {
	service, err := NewService(config.BrowserProxyConfig{
		Enabled: true, DefaultMode: "GLOBAL", DefaultNodeID: "hk",
		Nodes: []config.BrowserProxyNodeConfig{{
			ID: "hk", Name: "香港", Scheme: "HTTPS", Host: "proxy.example",
			Port: 8443, Enabled: true,
		}}, RuleSets: []config.BrowserProxyRuleSetConfig{{
			ID: "blocked", Domains: []string{"Example.com", "*.example.com"},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	data := service.Bootstrap(nil)
	if data.DefaultMode != ModeGlobal || data.DefaultNodeID != "hk" || data.Revision == "" {
		t.Fatalf("unexpected bootstrap: %+v", data)
	}
	if data.Nodes[0].Scheme != "https" || len(data.RuleSets[0].Domains) != 1 {
		t.Fatalf("normalization failed: %+v", data)
	}
}

func TestNewServiceRejectsInvalidNode(t *testing.T) {
	_, err := NewService(config.BrowserProxyConfig{Nodes: []config.BrowserProxyNodeConfig{{ID: "x", Name: "x", Scheme: "ftp", Host: "x", Port: 1}}})
	if err == nil {
		t.Fatal("expected unsupported scheme error")
	}
}
