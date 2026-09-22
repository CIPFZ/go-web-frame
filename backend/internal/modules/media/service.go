package media

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"gorm.io/gorm"
)

var allowedTypes = map[string]bool{"movie": true, "series": true, "anime": true, "variety": true}
var keyCleaner = regexp.MustCompile(`[\[\](){}【】（）「」《》]`)

type Service struct {
	db     *gorm.DB
	client *http.Client
}

type Config struct {
	DB *gorm.DB
}

type CollectSummary struct {
	RunID         uint `json:"runId"`
	SourceID      uint `json:"sourceId"`
	Fetched       int  `json:"fetched"`
	InsertedCount int  `json:"insertedCount"`
	UpdatedCount  int  `json:"updatedCount"`
	SkippedCount  int  `json:"skippedCount"`
	ErrorCount    int  `json:"errorCount"`
}

func NewService(cfg Config) (*Service, error) {
	if cfg.DB == nil {
		return nil, errors.New("media collector requires a database")
	}
	return &Service{db: cfg.DB, client: &http.Client{Timeout: 15 * time.Second}}, nil
}

func (s *Service) ListSources(ctx context.Context) ([]Source, error) {
	var items []Source
	err := s.db.WithContext(ctx).Order("priority asc, id asc").Find(&items).Error
	return items, err
}

func (s *Service) GetSource(ctx context.Context, id uint) (Source, error) {
	var item Source
	err := s.db.WithContext(ctx).First(&item, id).Error
	return item, err
}

func (s *Service) CreateSource(ctx context.Context, item Source) (Source, error) {
	if err := validateSource(item); err != nil {
		return item, err
	}
	if item.APIType == "" {
		item.APIType = "json"
	}
	if item.RequestIntervalSeconds <= 0 {
		item.RequestIntervalSeconds = 3
	}
	if item.TimeoutSeconds <= 0 || item.TimeoutSeconds > 120 {
		item.TimeoutSeconds = 15
	}
	if err := s.db.WithContext(ctx).Create(&item).Error; err != nil {
		return item, err
	}
	return item, nil
}

func (s *Service) UpdateSource(ctx context.Context, item Source) (Source, error) {
	if err := validateSource(item); err != nil {
		return item, err
	}
	var existing Source
	if err := s.db.WithContext(ctx).First(&existing, item.ID).Error; err != nil {
		return item, err
	}
	item.CreatedAt = existing.CreatedAt
	if err := s.db.WithContext(ctx).Model(&existing).Select(
		"name", "api_url", "api_type", "enabled", "priority", "allow_categories",
		"deny_categories", "request_interval_seconds", "timeout_seconds",
	).Updates(&item).Error; err != nil {
		return item, err
	}
	return s.GetSource(ctx, item.ID)
}

func (s *Service) HealthCheck(ctx context.Context, id uint) (HealthCheck, error) {
	source, err := s.GetSource(ctx, id)
	if err != nil {
		return HealthCheck{}, err
	}
	start := time.Now()
	check := HealthCheck{SourceID: id, CheckedAt: start, Result: "failed"}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source.APIURL, nil)
	if err == nil {
		timeout := time.Duration(source.TimeoutSeconds) * time.Second
		client := *s.client
		client.Timeout = timeout
		req.Header.Set("Accept", "application/json, application/xml;q=0.9")
		resp, requestErr := client.Do(req)
		if requestErr == nil {
			body, readErr := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
			resp.Body.Close()
			check.HTTPStatus = resp.StatusCode
			check.ResponseBytes = len(body)
			check.LatencyMS = int(time.Since(start).Milliseconds())
			if readErr == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 && validPayload(body, source.APIType) {
				check.Result = "ok"
				now := time.Now()
				source.LastOKAt = &now
				source.LastError = ""
				source.LastErrorAt = nil
				_ = s.db.WithContext(ctx).Model(&source).Updates(map[string]any{
					"last_ok_at": now, "last_error": "", "last_error_at": nil,
				}).Error
			} else {
				check.ErrorMessage = payloadError(resp.StatusCode, readErr)
				now := time.Now()
				source.LastErrorAt = &now
				source.LastError = check.ErrorMessage
				_ = s.db.WithContext(ctx).Model(&source).Updates(map[string]any{
					"last_error_at": now, "last_error": check.ErrorMessage,
				}).Error
			}
		} else {
			check.ErrorMessage = requestErr.Error()
		}
	} else {
		check.ErrorMessage = err.Error()
	}
	if check.Result != "ok" && check.ErrorMessage == "" {
		check.ErrorMessage = "invalid upstream response"
	}
	if check.Result != "ok" {
		now := time.Now()
		_ = s.db.WithContext(ctx).Model(&source).Updates(map[string]any{
			"last_error_at": now, "last_error": check.ErrorMessage,
		}).Error
	}
	if err := s.db.WithContext(ctx).Create(&check).Error; err != nil {
		return check, err
	}
	return check, nil
}

func (s *Service) Collect(ctx context.Context, id uint, limit int) (CollectSummary, error) {
	source, err := s.GetSource(ctx, id)
	if err != nil {
		return CollectSummary{}, err
	}
	if !source.Enabled {
		return CollectSummary{}, errors.New("source is disabled")
	}
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	started := time.Now()
	run := CollectionRun{SourceID: id, Mode: "incremental", StartedAt: started, Status: "running"}
	if err := s.db.WithContext(ctx).Create(&run).Error; err != nil {
		return CollectSummary{}, err
	}
	summary := CollectSummary{RunID: run.ID, SourceID: id}
	defer func() {
		finished := time.Now()
		updates := map[string]any{
			"finished_at": finished, "status": "success", "fetched": summary.Fetched,
			"inserted_count": summary.InsertedCount, "updated_count": summary.UpdatedCount,
			"skipped_count": summary.SkippedCount, "error_count": summary.ErrorCount,
		}
		if summary.ErrorCount > 0 {
			updates["status"] = "partial"
		}
		_ = s.db.WithContext(context.Background()).Model(&run).Updates(updates).Error
	}()
	body, err := s.fetch(ctx, source, limit)
	if err != nil {
		summary.ErrorCount++
		_ = s.db.WithContext(ctx).Model(&run).Updates(map[string]any{"status": "failed", "error_message": err.Error(), "finished_at": time.Now()}).Error
		return summary, err
	}
	records, err := decodeRecords(body, source.APIType)
	if err != nil {
		summary.ErrorCount++
		_ = s.db.WithContext(ctx).Model(&run).Updates(map[string]any{"status": "failed", "error_message": err.Error(), "finished_at": time.Now()}).Error
		return summary, err
	}
	if len(records) > limit {
		records = records[:limit]
	}
	for index, record := range records {
		summary.Fetched++
		if err := s.upsertRecord(ctx, source, record, index, &summary); err != nil {
			summary.ErrorCount++
		}
	}
	return summary, nil
}

func (s *Service) ListItems(ctx context.Context, limit int) ([]Item, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	var items []Item
	err := s.db.WithContext(ctx).Order("updated_at desc, id desc").Limit(limit).Find(&items).Error
	return items, err
}

type ItemDetail struct {
	Item      Item             `json:"item"`
	Playbacks []SourcePlayback `json:"playbacks"`
}

type SourcePlayback struct {
	Source     Source     `json:"source"`
	SourceItem SourceItem `json:"sourceItem"`
	Lines      []PlayLine `json:"lines"`
}

func (s *Service) GetItemDetail(ctx context.Context, id uint) (ItemDetail, error) {
	var item Item
	if err := s.db.WithContext(ctx).First(&item, id).Error; err != nil {
		return ItemDetail{}, err
	}
	var sourceItems []SourceItem
	if err := s.db.WithContext(ctx).Where("cms_media_id = ?", id).Order("id asc").Find(&sourceItems).Error; err != nil {
		return ItemDetail{}, err
	}
	playbacks := make([]SourcePlayback, 0, len(sourceItems))
	for _, sourceItem := range sourceItems {
		var source Source
		if err := s.db.WithContext(ctx).First(&source, sourceItem.SourceID).Error; err != nil {
			continue
		}
		var lines []PlayLine
		if err := s.db.WithContext(ctx).Where("source_item_id = ? AND enabled = ?", sourceItem.ID, true).
			Order("sort_order asc, id asc").Find(&lines).Error; err != nil {
			return ItemDetail{}, err
		}
		for i := range lines {
			if err := s.db.WithContext(ctx).Where("play_line_id = ?", lines[i].ID).
				Order("episode_no asc").Find(&lines[i].Episodes).Error; err != nil {
				return ItemDetail{}, err
			}
			lines[i].SourceName = source.Name
		}
		playbacks = append(playbacks, SourcePlayback{Source: source, SourceItem: sourceItem, Lines: lines})
	}
	return ItemDetail{Item: item, Playbacks: playbacks}, nil
}

func (s *Service) fetch(ctx context.Context, source Source, limit int) ([]byte, error) {
	u, err := url.Parse(source.APIURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("pg", "1")
	q.Set("ac", "detail")
	q.Set("pagesize", strconv.Itoa(limit))
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json, application/xml;q=0.9")
	client := *s.client
	client.Timeout = time.Duration(source.TimeoutSeconds) * time.Second
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("upstream status %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 20*1024*1024))
}

type rawRecord map[string]any

func decodeRecords(body []byte, apiType string) ([]rawRecord, error) {
	if strings.EqualFold(apiType, "xml") {
		return nil, errors.New("xml collection is reserved for the XML adapter")
	}
	var root any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&root); err != nil {
		return nil, fmt.Errorf("invalid json payload: %w", err)
	}
	list := findRecordList(root)
	if len(list) == 0 {
		return nil, errors.New("upstream payload contains no media list")
	}
	return list, nil
}

func findRecordList(value any) []rawRecord {
	switch node := value.(type) {
	case []any:
		items := make([]rawRecord, 0, len(node))
		for _, item := range node {
			if record, ok := item.(map[string]any); ok && looksLikeMedia(record) {
				items = append(items, rawRecord(record))
			}
		}
		if len(items) > 0 {
			return items
		}
		for _, item := range node {
			if nested := findRecordList(item); len(nested) > 0 {
				return nested
			}
		}
	case map[string]any:
		for _, key := range []string{"list", "data", "results", "result"} {
			if nested, ok := node[key]; ok {
				if records := findRecordList(nested); len(records) > 0 {
					return records
				}
			}
		}
		for _, nested := range node {
			if records := findRecordList(nested); len(records) > 0 {
				return records
			}
		}
	}
	return nil
}

func looksLikeMedia(record map[string]any) bool {
	for _, key := range []string{"vod_name", "name", "title", "vod_id"} {
		if strings.TrimSpace(asString(record[key])) != "" {
			return true
		}
	}
	return false
}

func (s *Service) upsertRecord(ctx context.Context, source Source, record rawRecord, index int, summary *CollectSummary) error {
	title := firstString(record, "vod_name", "name", "title")
	if title == "" {
		summary.SkippedCount++
		return nil
	}
	mediaType := mapMediaType(firstString(record, "type_name", "type", "category"))
	if !allowedTypes[mediaType] {
		summary.SkippedCount++
		return nil
	}
	externalID := firstString(record, "vod_id", "id", "uuid")
	if externalID == "" {
		externalID = fmt.Sprintf("generated-%x", sha256.Sum256([]byte(title+firstString(record, "vod_year", "year"))))
	}
	year := firstString(record, "vod_year", "year")
	area := firstString(record, "vod_area", "area", "region")
	normalized := normalizeKey(title, year, mediaType, area)
	rawBytes, _ := json.Marshal(record)
	rawHashBytes := sha256.Sum256(rawBytes)
	rawHash := hex.EncodeToString(rawHashBytes[:])
	now := time.Now()
	item := Item{
		Title: title, OriginalTitle: firstString(record, "vod_en", "original_title", "vod_name"),
		MediaType: mediaType, Year: year, Area: area,
		Language:      firstString(record, "vod_lang", "language"),
		Director:      firstString(record, "vod_director", "director"),
		Actors:        firstString(record, "vod_actor", "actors", "actor"),
		Description:   firstString(record, "vod_content", "content", "description"),
		CoverURL:      firstString(record, "vod_pic", "cover", "cover_url"),
		NormalizedKey: normalized, Status: "active",
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing Item
		itemCreated := false
		err := tx.Where("normalized_key = ?", normalized).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
			existing = item
			itemCreated = true
		} else if err != nil {
			return err
		} else {
			updates := map[string]any{
				"title": item.Title, "original_title": item.OriginalTitle, "media_type": item.MediaType,
				"year": item.Year, "area": item.Area, "language": item.Language, "director": item.Director,
				"actors": item.Actors, "description": item.Description, "cover_url": item.CoverURL, "status": "active",
			}
			if err := tx.Model(&existing).Updates(updates).Error; err != nil {
				return err
			}
			if existing.Description != item.Description || existing.CoverURL != item.CoverURL {
				itemCreated = false
			}
		}
		if itemCreated {
			summary.InsertedCount++
		} else {
			summary.UpdatedCount++
		}
		external := SourceItem{
			SourceID: source.ID, CMSMediaID: &existing.ID, ExternalID: externalID,
			ExternalTitle: title, FirstSeenAt: now, LastSeenAt: now, RawHash: rawHash,
			State: "active", RawMetadata: string(rawBytes),
		}
		var old SourceItem
		err = tx.Where("source_id = ? AND external_id = ?", source.ID, externalID).First(&old).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := tx.Create(&external).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			if err := tx.Model(&old).Updates(map[string]any{
				"cms_media_id": existing.ID, "external_title": title, "last_seen_at": now,
				"raw_hash": rawHash, "state": "active", "raw_metadata": string(rawBytes),
			}).Error; err != nil {
				return err
			}
			external = old
		}
		if err := s.upsertPlayback(tx, external.ID, record); err != nil {
			return err
		}
		return nil
	})
	_ = index
	return err
}

func (s *Service) upsertPlayback(tx *gorm.DB, sourceItemID uint, record rawRecord) error {
	froms := splitValues(firstString(record, "vod_play_from", "play_from", "source"))
	urls := splitValues(firstString(record, "vod_play_url", "play_url", "url"))
	if len(froms) == 0 || len(urls) == 0 {
		return nil
	}
	for i, rawURLList := range urls {
		lineName := "线路"
		if i < len(froms) && froms[i] != "" {
			lineName = froms[i]
		}
		lineKey := sanitizeKey(lineName)
		line := PlayLine{SourceItemID: sourceItemID, LineName: lineName, LineKey: lineKey, SortOrder: i + 1, Enabled: true}
		var existing PlayLine
		err := tx.Where("source_item_id = ? AND line_key = ?", sourceItemID, lineKey).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := tx.Create(&line).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			if err := tx.Model(&existing).Updates(map[string]any{"line_name": lineName, "sort_order": i + 1, "enabled": true}).Error; err != nil {
				return err
			}
			line = existing
		}
		for episodeNo, value := range strings.Split(rawURLList, "#") {
			episodeURL, episodeName := splitEpisode(strings.TrimSpace(value))
			if episodeURL == "" {
				continue
			}
			urlHashBytes := sha256.Sum256([]byte(episodeURL))
			episode := PlayEpisode{
				PlayLineID: line.ID, EpisodeNo: episodeNo + 1, EpisodeName: episodeName,
				StreamURL: episodeURL, StreamType: streamType(episodeURL), URLHash: hex.EncodeToString(urlHashBytes[:]),
			}
			var existingEpisode PlayEpisode
			err := tx.Where("play_line_id = ? AND episode_no = ?", line.ID, episode.EpisodeNo).First(&existingEpisode).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(&episode).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else if err := tx.Model(&existingEpisode).Updates(map[string]any{
				"episode_name": episode.EpisodeName, "stream_url": episode.StreamURL,
				"stream_type": episode.StreamType, "url_hash": episode.URLHash,
			}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func validateSource(item Source) error {
	if strings.TrimSpace(item.Name) == "" {
		return errors.New("source name is required")
	}
	parsed, err := url.ParseRequestURI(strings.TrimSpace(item.APIURL))
	if err != nil || parsed.Host == "" || parsed.Scheme != "https" {
		return errors.New("source API URL must be an absolute HTTPS URL")
	}
	if item.APIType != "" && item.APIType != "json" && item.APIType != "xml" {
		return errors.New("source API type must be json or xml")
	}
	return nil
}

func validPayload(body []byte, apiType string) bool {
	if strings.EqualFold(apiType, "xml") {
		return bytes.Contains(bytes.ToLower(body), []byte("<"))
	}
	var value any
	return json.Unmarshal(body, &value) == nil
}

func payloadError(status int, readErr error) string {
	if readErr != nil {
		return readErr.Error()
	}
	return fmt.Sprintf("upstream status %d or invalid payload", status)
}

func firstString(record map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(asString(record[key])); value != "" {
			return value
		}
	}
	return ""
}

func asString(value any) string {
	switch item := value.(type) {
	case string:
		return item
	case json.Number:
		return item.String()
	case float64:
		return strconv.FormatFloat(item, 'f', -1, 64)
	case int:
		return strconv.Itoa(item)
	default:
		return ""
	}
}

func mapMediaType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch {
	case strings.Contains(value, "动漫"), strings.Contains(value, "动画"), strings.Contains(value, "anime"):
		return "anime"
	case strings.Contains(value, "综艺"), strings.Contains(value, "variety"), strings.Contains(value, "show"):
		return "variety"
	case strings.Contains(value, "剧"), strings.Contains(value, "series"), strings.Contains(value, "电视剧"):
		return "series"
	case strings.Contains(value, "电影"), strings.Contains(value, "movie"):
		return "movie"
	default:
		return "movie"
	}
}

func normalizeKey(title, year, mediaType, area string) string {
	title = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, strings.ToLower(strings.TrimSpace(title)))
	title = keyCleaner.ReplaceAllString(title, "")
	for _, marker := range []string{"1080p", "2160p", "4k", "高清", "无删减"} {
		title = strings.Replace(title, marker, "", 1)
	}
	return strings.Join([]string{title, strings.TrimSpace(year), mediaType, strings.TrimSpace(area)}, "|")
}

func splitValues(value string) []string {
	var result []string
	for _, part := range strings.Split(value, "$$$") {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func splitEpisode(value string) (string, string) {
	parts := strings.SplitN(value, "$", 2)
	if len(parts) == 1 {
		return strings.TrimSpace(parts[0]), ""
	}
	return strings.TrimSpace(parts[1]), strings.TrimSpace(parts[0])
}

func sanitizeKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return '_'
	}, value)
	value = strings.Trim(value, "_")
	if value == "" {
		return "line"
	}
	return value[:min(len(value), 100)]
}

func streamType(value string) string {
	lower := strings.ToLower(value)
	switch {
	case strings.Contains(lower, ".m3u8"):
		return "m3u8"
	case strings.Contains(lower, ".mp4"):
		return "mp4"
	default:
		return "other"
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
