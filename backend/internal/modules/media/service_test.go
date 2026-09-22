package media

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeKeyRemovesPresentationMarkers(t *testing.T) {
	require.Equal(t, "流浪地球2|2023|movie|中国", normalizeKey(" 流浪地球2 4K [高清] ", "2023", "movie", "中国"))
}

func TestDecodeRecordsSupportsMacCMSDataList(t *testing.T) {
	body := []byte(`{"code":1,"page":1,"list":[{"vod_id":42,"vod_name":"测试电影","type_name":"电影","vod_year":"2025"}]}`)
	records, err := decodeRecords(body, "json")
	require.NoError(t, err)
	require.Len(t, records, 1)
	require.Equal(t, "测试电影", firstString(records[0], "vod_name"))
}

func TestPlaybackParsing(t *testing.T) {
	require.Equal(t, []string{"bfzym3u8", "hhyun"}, splitValues("bfzym3u8$$$hhyun"))
	url, name := splitEpisode("正片$https://cdn.example/test.m3u8")
	require.Equal(t, "https://cdn.example/test.m3u8", url)
	require.Equal(t, "正片", name)
}

func TestValidateSourceRejectsPlainHTTP(t *testing.T) {
	require.Error(t, validateSource(Source{Name: "bad", APIURL: "http://example.com/api"}))
	require.NoError(t, validateSource(Source{Name: "good", APIURL: "https://example.com/api"}))
}
