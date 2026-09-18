package magnet

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
)

type Config struct {
	CacheDir        string
	CacheTTL        time.Duration
	MetadataTimeout time.Duration
	MaxFiles        int
	FetchCover      bool
	MaxCoverBytes   int64
	MaxConcurrent   int
}

type Service struct {
	ctx          context.Context
	cancel       context.CancelFunc
	cacheDir     string
	cacheTTL     time.Duration
	timeout      time.Duration
	maxFiles     int
	fetchCover   bool
	maxCoverSize int64
	limiter      chan struct{}
}

type Preview struct {
	InfoHash       string    `json:"info_hash"`
	Name           string    `json:"name"`
	Title          string    `json:"title"`
	DisplayName    string    `json:"display_name"`
	Trackers       []string  `json:"trackers"`
	FileCount      int       `json:"file_count"`
	FilesTruncated bool      `json:"files_truncated"`
	TotalSize      int64     `json:"total_size"`
	TotalSizeHuman string    `json:"total_size_human"`
	ContentType    string    `json:"content_type"`
	Files          []File    `json:"files"`
	Cover          *Cover    `json:"cover,omitempty"`
	MetadataSource string    `json:"metadata_source"`
	RetrievedAt    time.Time `json:"retrieved_at"`
	Cached         bool      `json:"cached"`
}

type File struct {
	Index     int    `json:"index"`
	Path      string `json:"path"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	SizeHuman string `json:"size_human"`
	Extension string `json:"extension,omitempty"`
	IsImage   bool   `json:"is_image"`
	IsVideo   bool   `json:"is_video"`
	IsAudio   bool   `json:"is_audio"`
	Type      string `json:"type"`
	IsCover   bool   `json:"is_cover"`
}

type Cover struct {
	URL      string `json:"url"`
	Path     string `json:"path"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

func NewService(cfg Config) (*Service, error) {
	if cfg.CacheDir == "" {
		cfg.CacheDir = filepath.Join(os.TempDir(), "base-frame", "magnet-preview")
	}
	if cfg.CacheTTL <= 0 {
		cfg.CacheTTL = 3 * time.Hour
	}
	if cfg.MetadataTimeout <= 0 || cfg.MetadataTimeout > 45*time.Second {
		cfg.MetadataTimeout = 45 * time.Second
	}
	if cfg.MaxFiles <= 0 || cfg.MaxFiles > 10000 {
		cfg.MaxFiles = 1000
	}
	if cfg.MaxCoverBytes <= 0 || cfg.MaxCoverBytes > 5*1024*1024 {
		cfg.MaxCoverBytes = 2 * 1024 * 1024
	}
	if cfg.MaxConcurrent <= 0 || cfg.MaxConcurrent > 2 {
		cfg.MaxConcurrent = 1
	}
	if err := os.MkdirAll(cfg.CacheDir, 0750); err != nil {
		return nil, fmt.Errorf("create magnet cache: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	service := &Service{
		ctx: ctx, cancel: cancel,
		cacheDir:     cfg.CacheDir,
		cacheTTL:     cfg.CacheTTL,
		timeout:      cfg.MetadataTimeout,
		maxFiles:     cfg.MaxFiles,
		fetchCover:   cfg.FetchCover,
		maxCoverSize: cfg.MaxCoverBytes,
		limiter:      make(chan struct{}, cfg.MaxConcurrent),
	}
	service.pruneCache("")
	return service, nil
}

func (s *Service) Close() error {
	s.cancel()
	return nil
}

func (s *Service) Preview(ctx context.Context, value string) (Preview, error) {
	parsed, err := Parse(value)
	if err != nil {
		return Preview{}, err
	}
	if cached, ok := s.readCache(parsed.InfoHash); ok {
		cached.Cached = true
		return cached, nil
	}
	select {
	case s.limiter <- struct{}{}:
		defer func() { <-s.limiter }()
	case <-ctx.Done():
		return Preview{}, ctx.Err()
	default:
		return Preview{}, errors.New("预览服务正在处理其他请求，请稍后重试")
	}

	if cached, ok := s.readCache(parsed.InfoHash); ok {
		cached.Cached = true
		return cached, nil
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stop := context.AfterFunc(s.ctx, cancel)
	defer stop()
	return s.fetch(ctx, parsed)
}

func (s *Service) fetch(ctx context.Context, parsed ParsedMagnet) (Preview, error) {
	workDir, err := os.MkdirTemp(s.cacheDir, "pieces-")
	if err != nil {
		return Preview{}, err
	}
	defer os.RemoveAll(workDir)
	clientCfg := torrent.NewDefaultClientConfig()
	clientCfg.DataDir = workDir
	clientCfg.ListenPort = 0
	clientCfg.NoDefaultPortForwarding = true
	clientCfg.NoUpload = true
	clientCfg.Seed = false
	clientCfg.DisablePEX = true
	clientCfg.DisableWebtorrent = true
	clientCfg.DisableWebseeds = true
	clientCfg.AcceptPeerConnections = false
	clientCfg.DialForPeerConns = true
	clientCfg.IPBlocklist = publicPeersOnly{}
	clientCfg.TrackerDialContext = publicDialContext
	clientCfg.LookupTrackerIp = publicTrackerIPs
	clientCfg.EstablishedConnsPerTorrent = 24
	clientCfg.HalfOpenConnsPerTorrent = 8
	clientCfg.TotalHalfOpenConns = 8
	clientCfg.MaxUnverifiedBytes = 16 * 1024 * 1024
	client, err := torrent.NewClient(clientCfg)
	if err != nil {
		return Preview{}, fmt.Errorf("create torrent metadata client: %w", err)
	}
	defer client.Close()
	torrentItem, err := client.AddMagnet(parsed.URI)
	if err != nil {
		return Preview{}, fmt.Errorf("add magnet: %w", err)
	}
	defer torrentItem.Drop()

	waitCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	select {
	case <-torrentItem.GotInfo():
	case <-waitCtx.Done():
		return Preview{}, fmt.Errorf("torrent metadata unavailable within %s", s.timeout)
	}

	info := torrentItem.Info()
	if info == nil {
		return Preview{}, errors.New("torrent metadata is empty")
	}
	files := torrentItem.Files()
	records := make([]File, 0, minInt(len(files), s.maxFiles))
	var coverCandidate *torrent.File
	var coverCandidateRecord File
	var totalSize int64
	composition := map[string]int64{}
	counts := map[string]int{}
	for index, item := range files {
		path := item.DisplayPath()
		size := item.Length()
		ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
		record := File{
			Index: index, Path: path, Name: filepath.Base(path), Size: size,
			SizeHuman: humanSize(size), Extension: ext, Type: fileType(ext),
			IsImage: isImage(ext), IsVideo: isVideo(ext), IsAudio: isAudio(ext),
		}
		composition[record.Type] += size
		counts[record.Type]++
		record.IsCover = record.IsImage && isCoverName(path)
		if record.IsCover && coverCandidate == nil && size > 0 && size <= s.maxCoverSize {
			coverCandidate, coverCandidateRecord = item, record
		}
		if index < s.maxFiles {
			records = append(records, record)
		}
		totalSize += size
	}
	if coverCandidate == nil {
		for index, item := range files {
			ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(item.DisplayPath())), ".")
			if isImage(ext) && item.Length() > 0 && item.Length() <= s.maxCoverSize {
				coverCandidate = item
				coverCandidateRecord = File{
					Index: index, Path: item.DisplayPath(), Name: filepath.Base(item.DisplayPath()),
					Size: item.Length(), SizeHuman: humanSize(item.Length()), Extension: ext, IsImage: true,
				}
				break
			}
		}
	}

	result := Preview{
		InfoHash: parsed.InfoHash, Name: torrentItem.Name(),
		DisplayName: parsed.DisplayName, Trackers: parsed.Trackers,
		FileCount: len(files), FilesTruncated: len(files) > len(records),
		TotalSize: totalSize, TotalSizeHuman: humanSize(totalSize),
		ContentType: dominantType(composition, counts), MetadataSource: "bittorrent-ut_metadata",
		RetrievedAt: time.Now().UTC(), Files: records,
	}
	if result.Name == "" {
		result.Name = parsed.DisplayName
	}
	result.Title = result.Name
	if len(files) == 1 && fileType(strings.TrimPrefix(strings.ToLower(filepath.Ext(result.Name)), ".")) != "other" {
		result.Title = strings.TrimSuffix(result.Name, filepath.Ext(result.Name))
	}
	if result.Title == "" {
		result.Title = result.Name
	}

	if s.fetchCover && coverCandidate != nil && info.PieceLength <= 8*1024*1024 {
		if cover, err := s.fetchCoverFile(ctx, torrentItem, coverCandidate, coverCandidateRecord); err == nil {
			result.Cover = cover
		}
	}
	if err := s.writeCache(parsed.InfoHash, result); err != nil {
		return Preview{}, err
	}
	return result, nil
}

func (s *Service) fetchCoverFile(ctx context.Context, item *torrent.Torrent, file *torrent.File, record File) (*Cover, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	file.SetPriority(torrent.PiecePriorityNow)
	reader := file.NewReader()
	reader.SetContext(ctx)
	reader.SetReadahead(0)
	defer reader.Close()
	data, err := io.ReadAll(io.LimitReader(reader, s.maxCoverSize+1))
	if err != nil {
		return nil, fmt.Errorf("read cover file: %w", err)
	}
	if int64(len(data)) > s.maxCoverSize {
		return nil, errors.New("cover file exceeds configured limit")
	}
	imageCfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || imageCfg.Width <= 0 || imageCfg.Height <= 0 || int64(imageCfg.Width)*int64(imageCfg.Height) > 40000000 {
		return nil, errors.New("invalid or oversized cover image")
	}
	hash := strings.ToLower(item.InfoHash().String())
	coverDir := filepath.Join(s.cacheDir, hash)
	if err := os.MkdirAll(coverDir, 0750); err != nil {
		return nil, err
	}
	filename := "cover." + record.Extension
	target := filepath.Join(coverDir, filename)
	if err := os.WriteFile(target, data, 0640); err != nil {
		return nil, err
	}
	return &Cover{
		URL: "", Path: record.Path,
		Filename: filename, Size: int64(len(data)), MimeType: http.DetectContentType(data),
	}, nil
}

func (s *Service) ServeCover(hash string) (string, string, error) {
	if !hexHash.MatchString(hash) {
		return "", "", errors.New("invalid info hash")
	}
	dir := filepath.Join(s.cacheDir, strings.ToLower(hash))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", "", err
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "cover.") {
			mime := mimeType(strings.TrimPrefix(filepath.Ext(entry.Name()), "."))
			return filepath.Join(dir, entry.Name()), mime, nil
		}
	}
	return "", "", os.ErrNotExist
}

func (s *Service) cachePath(hash string) string {
	return filepath.Join(s.cacheDir, hash, "preview.json")
}

func (s *Service) readCache(hash string) (Preview, bool) {
	if !hexHash.MatchString(hash) {
		return Preview{}, false
	}
	data, err := os.ReadFile(s.cachePath(hash))
	if err != nil {
		return Preview{}, false
	}
	var preview Preview
	if json.Unmarshal(data, &preview) != nil || time.Since(preview.RetrievedAt) > s.cacheTTL {
		return Preview{}, false
	}
	return preview, true
}

func (s *Service) writeCache(hash string, preview Preview) error {
	dir := filepath.Dir(s.cachePath(hash))
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}
	data, err := json.MarshalIndent(preview, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "preview-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0640); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, s.cachePath(hash)); err != nil {
		return err
	}
	s.pruneCache(hash)
	return nil
}

func (s *Service) ReadCachedPreview(hash string) (Preview, bool) {
	preview, ok := s.readCache(strings.ToLower(hash))
	return preview, ok
}

func humanSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	value := float64(size)
	for _, unit := range []string{"KiB", "MiB", "GiB", "TiB"} {
		value /= 1024
		if value < 1024 || unit == "TiB" {
			return fmt.Sprintf("%.1f %s", value, unit)
		}
	}
	return fmt.Sprintf("%d B", size)
}

func isCoverName(path string) bool {
	lower := strings.ToLower(path)
	for _, part := range []string{"cover", "poster", "folder", "fanart", "thumb", "front"} {
		if strings.Contains(lower, part) {
			return true
		}
	}
	return false
}

func mimeType(ext string) string {
	switch ext {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "webp":
		return "image/webp"
	case "gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
