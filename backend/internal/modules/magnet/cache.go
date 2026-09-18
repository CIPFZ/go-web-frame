package magnet

import (
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Only generated hash directories are eligible for deletion; temporary torrent
// workspaces belong to their active request and are removed after client.Close.
func (s *Service) pruneCache(current string) {
	entries, err := os.ReadDir(s.cacheDir)
	if err != nil {
		return
	}
	type cached struct {
		name     string
		modified time.Time
	}
	items := []cached{}
	for _, entry := range entries {
		if !entry.IsDir() || !hexHash.MatchString(entry.Name()) || entry.Name() == current {
			continue
		}
		info, err := os.Stat(s.cachePath(entry.Name()))
		if err != nil {
			continue
		}
		if time.Since(info.ModTime()) > s.cacheTTL {
			_ = os.RemoveAll(filepath.Join(s.cacheDir, entry.Name()))
			continue
		}
		items = append(items, cached{entry.Name(), info.ModTime()})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].modified.Before(items[j].modified) })
	for len(items) > 63 {
		_ = os.RemoveAll(filepath.Join(s.cacheDir, items[0].name))
		items = items[1:]
	}
}
