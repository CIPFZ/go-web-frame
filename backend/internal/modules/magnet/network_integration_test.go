package magnet

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// Opt-in smoke test with a public, freely distributable resource. Normal tests
// are deterministic and do not depend on trackers or peer availability.
func TestPublicMetadataSmoke(t *testing.T) {
	path := os.Getenv("MAGNET_SMOKE_URI_FILE")
	if path == "" {
		t.Skip("set MAGNET_SMOKE_URI_FILE to a public magnet fixture")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	cache := t.TempDir()
	s, err := NewService(Config{CacheDir: cache, MetadataTimeout: 45 * time.Second, FetchCover: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	result, err := s.Preview(context.Background(), strings.TrimSpace(string(data)))
	if err != nil {
		t.Fatal(err)
	}
	if result.Name == "" || result.TotalSize <= 0 || result.FileCount == 0 {
		t.Fatalf("incomplete metadata: %+v", result)
	}
	entries, err := os.ReadDir(cache)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "pieces-") {
			t.Fatal("temporary payload not cleaned")
		}
	}
	t.Logf("name=%s type=%s size=%d files=%d cover=%v", result.Name, result.ContentType, result.TotalSize, result.FileCount, result.Cover != nil)
}
