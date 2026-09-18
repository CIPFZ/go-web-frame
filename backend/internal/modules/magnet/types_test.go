package magnet

import (
	"context"
	"net"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestFileTypesAndCompanionArtwork(t *testing.T) {
	cases := []struct {
		files []string
		sizes []int64
		want  string
	}{
		{[]string{"cover.jpg", "movie.mkv"}, []int64{1000, 900000}, "video"},
		{[]string{"folder.png", "song.FLAC"}, []int64{1000, 100000}, "audio"},
		{[]string{"readme.txt", "cover.jpg", "linux.iso"}, []int64{100, 1000, 900000}, "disk_image"},
		{[]string{"readme.txt", "data.7z"}, []int64{100, 900000}, "archive"},
		{[]string{"cover.jpg", "book.epub"}, []int64{1000, 900000}, "document"},
		{[]string{"01.png", "02.jpg"}, []int64{5000, 9000}, "image"},
		{[]string{"README.txt", "01.png"}, []int64{10, 9000}, "image"},
		{[]string{"data.unknown"}, []int64{9000}, "other"},
		{[]string{"sample.mp4", "album.flac"}, []int64{100, 9000}, "audio"},
	}
	for _, tc := range cases {
		sizes, counts := map[string]int64{}, map[string]int{}
		for i, name := range tc.files {
			ext := name[strings.LastIndex(name, ".")+1:]
			kind := fileType(ext)
			sizes[kind] += tc.sizes[i]
			counts[kind]++
		}
		if got := dominantType(sizes, counts); got != tc.want {
			t.Errorf("%v: got %s want %s", tc.files, got, tc.want)
		}
	}
}

func TestSanitizedMagnetDoesNotRetainSourceURLs(t *testing.T) {
	uri := "magnet:?xt=urn:btih:0123456789abcdef0123456789abcdef01234567&xs=http://127.0.0.1/admin&ws=http://localhost/file&x.pe=127.0.0.1:80&tr=http://127.0.0.1:80&tr=https%3A%2F%2Ftracker.example%2Fannounce%3Fkey%3Da%252Bb"
	result, err := Parse(uri)
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ := url.Parse(result.URI)
	for _, key := range []string{"xs", "ws", "x.pe"} {
		if parsed.Query().Has(key) {
			t.Fatal("untrusted parameter retained", key)
		}
	}
	if len(result.Trackers) != 1 || !strings.Contains(result.Trackers[0], "key=a%2Bb") {
		t.Fatalf("tracker escaped twice: %v", result.Trackers)
	}
	if strings.Contains(result.URI, "127.0.0.1") {
		t.Fatal("private tracker retained")
	}
}

func TestInvalidMagnets(t *testing.T) {
	for _, uri := range []string{"", "https://example.com", "magnet:?xt=urn:btih:bad", "magnet:?xt=urn:btmh:123",
		"magnet:?xt=urn:btih:0123456789abcdef0123456789abcdef01234567&xt=urn:btih:0123456789abcdef0123456789abcdef01234567"} {
		if _, err := Parse(uri); err == nil {
			t.Errorf("accepted %q", uri)
		}
	}
}

func TestPublicAddressesOnly(t *testing.T) {
	for _, addr := range []string{"127.0.0.1", "10.0.0.1", "172.16.1.1", "192.168.1.1", "169.254.169.254", "::1", "fc00::1", "100.64.0.1", "0.0.0.0", "224.0.0.1"} {
		if publicIP(net.ParseIP(addr)) {
			t.Errorf("accepted private address %s", addr)
		}
	}
	if !publicIP(net.ParseIP("1.1.1.1")) {
		t.Fatal("public address blocked")
	}
}

func TestCacheAndBusyRequests(t *testing.T) {
	s, err := NewService(Config{CacheDir: t.TempDir(), MaxConcurrent: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	hash := "0123456789abcdef0123456789abcdef01234567"
	original := Preview{InfoHash: hash, Name: "archive.iso", ContentType: "disk_image", RetrievedAt: time.Now()}
	if err := s.writeCache(hash, original); err != nil {
		t.Fatal(err)
	}
	got, err := s.Preview(context.Background(), "magnet:?xt=urn:btih:"+hash)
	if err != nil || !got.Cached || got.Name != original.Name || got.Cover != nil {
		t.Fatalf("cache: %#v %v", got, err)
	}
	original.RetrievedAt = time.Now().Add(-25 * time.Hour)
	if err := s.writeCache(hash, original); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.readCache(hash); ok {
		t.Fatal("expired cache used")
	}
	s.limiter <- struct{}{}
	if _, err := s.Preview(context.Background(), "magnet:?xt=urn:btih:"+hash); err == nil {
		t.Fatal("busy request accepted")
	}
	<-s.limiter
	if _, _, err := s.ServeCover("../test"); err == nil {
		t.Fatal("path traversal accepted")
	}
}
