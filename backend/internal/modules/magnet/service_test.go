package magnet

import (
	"encoding/base32"
	"encoding/hex"
	"testing"
	"time"
)

func TestParseHexAndTrackerFiltering(t *testing.T) {
	result, err := Parse("magnet:?xt=urn:btih:0123456789abcdef0123456789abcdef01234567&dn=Example%20Name&tr=udp%3A%2F%2Ftracker.example%3A80&tr=http%3A%2F%2F127.0.0.1%3A8080")
	if err != nil {
		t.Fatal(err)
	}
	if result.InfoHash != "0123456789abcdef0123456789abcdef01234567" || result.DisplayName != "Example Name" {
		t.Fatalf("unexpected parse result: %#v", result)
	}
	if len(result.Trackers) != 1 || result.Trackers[0] != "udp://tracker.example:80" {
		t.Fatalf("unexpected trackers: %#v", result.Trackers)
	}
}

func TestParseBase32(t *testing.T) {
	hexHash := "0123456789abcdef0123456789abcdef01234567"
	bytes, err := hex.DecodeString(hexHash)
	if err != nil {
		t.Fatal(err)
	}
	base32Hash := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes)
	result, err := Parse("magnet:?xt=urn:btih:" + base32Hash)
	if err != nil {
		t.Fatal(err)
	}
	if result.InfoHash != hexHash {
		t.Fatalf("unexpected hash: %s", result.InfoHash)
	}
}

func TestCacheTTL(t *testing.T) {
	s, err := NewService(Config{CacheDir: t.TempDir(), CacheTTL: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	if s.cacheTTL != time.Hour {
		t.Fatalf("unexpected cache ttl: %s", s.cacheTTL)
	}
	hash := "0123456789abcdef0123456789abcdef01234567"
	if err := s.writeCache(hash, Preview{InfoHash: hash, RetrievedAt: time.Now().Add(-2 * time.Hour)}); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.readCache(hash); ok {
		t.Fatal("expired cache was returned")
	}
}
