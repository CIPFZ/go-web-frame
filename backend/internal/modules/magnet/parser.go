package magnet

import (
	"encoding/base32"
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)

var hexHash = regexp.MustCompile("^[0-9a-fA-F]{40}$")
var base32Hash = regexp.MustCompile("^[A-Za-z2-7]{32}$")

var allowedTrackerSchemes = map[string]bool{
	"http":  true,
	"https": true,
	"udp":   true,
}

type ParsedMagnet struct {
	URI         string
	InfoHash    string
	DisplayName string
	Trackers    []string
}

func Parse(value string) (ParsedMagnet, error) {
	uri := strings.TrimSpace(value)
	if uri == "" || len(uri) > 16384 {
		return ParsedMagnet{}, errors.New("magnet link is required")
	}
	parsed, err := url.Parse(uri)
	if err != nil || !strings.EqualFold(parsed.Scheme, "magnet") {
		return ParsedMagnet{}, errors.New("only magnet URIs are supported")
	}

	values := parsed.Query()["xt"]
	hashes := make([]string, 0, 1)
	for _, item := range values {
		item = strings.TrimSpace(item)
		if strings.HasPrefix(strings.ToLower(item), "urn:btih:") {
			hashes = append(hashes, item[9:])
		}
	}
	if len(hashes) != 1 {
		return ParsedMagnet{}, errors.New("magnet must contain exactly one urn:btih xt value")
	}
	infoHash, err := normalizeHash(hashes[0])
	if err != nil {
		return ParsedMagnet{}, err
	}

	trackers := make([]string, 0, 8)
	seen := make(map[string]struct{})
	for _, raw := range parsed.Query()["tr"] {
		tracker := strings.TrimSpace(raw)
		if len(tracker) == 0 || len(tracker) > 2048 {
			continue
		}
		trackerURL, err := url.Parse(tracker)
		if err != nil || !allowedTrackerSchemes[strings.ToLower(trackerURL.Scheme)] || trackerURL.Host == "" || trackerURL.User != nil || blockedHost(trackerURL.Hostname()) {
			continue
		}
		if _, ok := seen[tracker]; ok {
			continue
		}
		seen[tracker] = struct{}{}
		trackers = append(trackers, tracker)
		if len(trackers) == 32 {
			break
		}
	}

	displayName := strings.TrimSpace(parsed.Query().Get("dn"))
	if len(displayName) > 512 {
		displayName = displayName[:512]
	}
	query := url.Values{"xt": {"urn:btih:" + infoHash}, "dn": {displayName}, "tr": trackers}
	return ParsedMagnet{
		URI:         "magnet:?" + query.Encode(),
		InfoHash:    infoHash,
		DisplayName: displayName,
		Trackers:    trackers,
	}, nil
}

func normalizeHash(value string) (string, error) {
	value = strings.TrimSpace(value)
	switch {
	case hexHash.MatchString(value):
		return strings.ToLower(value), nil
	case base32Hash.MatchString(value):
		decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(value))
		if err != nil || len(decoded) != 20 {
			return "", errors.New("magnet xt contains an invalid base32 BTIH")
		}
		return fmt.Sprintf("%x", decoded), nil
	default:
		return "", errors.New("magnet xt must contain a 40-character hex or 32-character base32 BTIH")
	}
}

func blockedHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" || host == "localhost" || host == "localhost.localdomain" {
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() || ip.IsMulticast()
}
