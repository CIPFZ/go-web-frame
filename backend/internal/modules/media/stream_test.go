package media

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestPlaybackManifestRewritesNestedPlaylistSegmentsAndKeys(t *testing.T) {
	base, _ := url.Parse("https://video.example.com/show/master.m3u8")
	proxy := "/api/v1/media/episodes/7/stream"
	manifest := "#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=1000\n720/index.m3u8\n#EXT-X-KEY:METHOD=AES-128,URI=\"../key?id=2\"\n#EXT-X-MAP:URI=\"init.mp4\"\n#EXTINF:6,\nsegment.ts?token=a%2Bb"
	got := rewriteManifest(manifest, base, proxy)
	for _, target := range []string{"https://video.example.com/show/720/index.m3u8", "https://video.example.com/key?id=2", "https://video.example.com/show/init.mp4", "https://video.example.com/show/segment.ts?token=a%2Bb"} {
		if !strings.Contains(got, proxy+"?target="+url.QueryEscape(target)) {
			t.Fatalf("manifest did not rewrite %s: %s", target, got)
		}
	}
}

func TestPlaybackRejectsUnrelatedAndCredentialURLs(t *testing.T) {
	root, _ := url.Parse("https://video.example.com/root.m3u8")
	for _, raw := range []string{"https://evil.example/root", "https://example.com/root", "https://video.example.com.evil/root", "https://user:password@video.example.com/root", "file:///tmp/video"} {
		target, _ := url.Parse(raw)
		if validPlaybackURL(target) && allowedPlaybackTarget(root, target) {
			t.Fatalf("unexpected permitted target: %s", raw)
		}
	}
}

func TestReadPlaybackManifestRejectsHTMLAndOversize(t *testing.T) {
	base, _ := url.Parse("https://video.example.com/index.m3u8")
	for _, body := range []string{"<html>access denied</html>", "#EXTM3U\n" + strings.Repeat("x", 2<<20)} {
		resp := &http.Response{Body: io.NopCloser(strings.NewReader(body))}
		if _, err := readManifest(resp, base, "application/vnd.apple.mpegurl", "/stream"); err == nil {
			t.Fatal("expected invalid manifest error")
		}
	}
}
