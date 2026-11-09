package apispec

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func init() { gin.SetMode(gin.TestMode) }

func serve(h gin.HandlerFunc, body string, headers ...string) *httptest.ResponseRecorder {
	r := gin.New()
	r.POST("/x", Observe(), h)
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &m))
	return m
}

func TestWriteOKShape(t *testing.T) {
	w := serve(func(c *gin.Context) { WriteOK(c, map[string]string{"sid": "app_x"}) }, "", HeaderRequestID, "req-abcdefgh")
	require.Equal(t, http.StatusOK, w.Code)
	m := decodeBody(t, w)
	require.EqualValues(t, 0, m["code"])
	require.Equal(t, "", m["code_name"])
	require.Equal(t, "success", m["message"])
	require.NotContains(t, m, "error")
	meta := m["meta"].(map[string]any)
	require.Equal(t, "req-abcdefgh", meta["request_id"])
	require.Equal(t, "req-abcdefgh", w.Header().Get(HeaderRequestID))
	require.NotEmpty(t, w.Header().Get(HeaderLatencyMS))
}

func TestWriteOKGeneratesRequestIDForInvalidInbound(t *testing.T) {
	w := serve(func(c *gin.Context) { WriteOK(c, nil) }, "", HeaderRequestID, "bad id with spaces")
	id := w.Header().Get(HeaderRequestID)
	require.True(t, IsSID("req", id), id)
	require.Equal(t, map[string]any{}, decodeBody(t, w)["data"])
}

func TestWriteErrorDomainAndInternal(t *testing.T) {
	w := serve(func(c *gin.Context) {
		WriteError(c, FailedPrecondition("task.already_claimed", "task is already claimed").WithMetadata("task_sid", "tsk_1"))
	}, "")
	require.Equal(t, http.StatusBadRequest, w.Code)
	m := decodeBody(t, w)
	require.EqualValues(t, 100010, m["code"])
	require.Equal(t, "FailedPrecondition", m["code_name"])
	require.Nil(t, m["data"])
	e := m["error"].(map[string]any)
	require.Equal(t, "task.already_claimed", e["reason"])
	require.Equal(t, false, e["retryable"])
	require.Equal(t, "tsk_1", e["metadata"].(map[string]any)["task_sid"])

	w = serve(func(c *gin.Context) { WriteError(c, errors.New("pq: password authentication failed for user x")) }, "")
	require.Equal(t, http.StatusInternalServerError, w.Code)
	require.NotContains(t, w.Body.String(), "pq:")
	require.Equal(t, true, decodeBody(t, w)["error"].(map[string]any)["retryable"])
}

func TestRateLimitedSetsRetryAfter(t *testing.T) {
	w := serve(func(c *gin.Context) { WriteError(c, RateLimited(7)) }, "")
	require.Equal(t, http.StatusTooManyRequests, w.Code)
	require.Equal(t, "7", w.Header().Get("Retry-After"))
	e := decodeBody(t, w)["error"].(map[string]any)
	require.EqualValues(t, 7, e["retry_after_seconds"])
	require.Equal(t, true, e["retryable"])
}

func TestQuotaExceededRetryableOnlyWithWait(t *testing.T) {
	require.False(t, QuotaExceeded("quota.daily_upload", "daily upload quota exhausted", 0).Retryable)
	require.True(t, QuotaExceeded("quota.daily_upload", "daily upload quota exhausted", 60).Retryable)
}

type decodeReq struct {
	App    string   `json:"app" validate:"required"`
	Kind   string   `json:"kind" validate:"omitempty,oneof=software config"`
	Size   int32    `json:"size" validate:"gte=0,lte=10"`
	AppSID string   `json:"app_sid" validate:"omitempty,sid=app"`
	Nested *nested  `json:"nested"`
	Tags   []string `json:"tags" validate:"max=2"`
}

type nested struct {
	Name string `json:"name" validate:"required"`
}

func decodeWith(body string) (*httptest.ResponseRecorder, *decodeReq) {
	var got decodeReq
	w := serve(func(c *gin.Context) {
		if err := Decode(c, &got); err != nil {
			WriteError(c, err)
			return
		}
		WriteOK(c, nil)
	}, body)
	return w, &got
}

func violations(t *testing.T, w *httptest.ResponseRecorder) []map[string]any {
	t.Helper()
	e := decodeBody(t, w)["error"].(map[string]any)
	raw := e["field_violations"].([]any)
	out := make([]map[string]any, len(raw))
	for i, v := range raw {
		out[i] = v.(map[string]any)
	}
	return out
}

func TestDecode(t *testing.T) {
	w, got := decodeWith(`{"app":"qq","kind":"software","size":3}`)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "qq", got.App)

	w, _ = decodeWith(``)
	require.Equal(t, "MissingParameter", decodeBody(t, w)["code_name"], "empty body is {} and app is required")
	require.Equal(t, "app", violations(t, w)[0]["field"])

	w, _ = decodeWith(`{"app":"qq","colour":"red"}`)
	require.Equal(t, "InvalidParameter", decodeBody(t, w)["code_name"])
	require.Equal(t, "field.unknown", violations(t, w)[0]["reason"])
	require.NotContains(t, w.Body.String(), "red", "rejected values are never echoed")

	w, _ = decodeWith(`{"app":"qq","size":"big"}`)
	require.Equal(t, "type.mismatch", violations(t, w)[0]["reason"])
	require.Equal(t, "size", violations(t, w)[0]["field"])

	w, _ = decodeWith(`{"app":"qq",`)
	require.Equal(t, "BadRequest", decodeBody(t, w)["code_name"])

	w, _ = decodeWith(`{"app":"qq"} {"app":"x"}`)
	require.Equal(t, "BadRequest", decodeBody(t, w)["code_name"])

	w, _ = decodeWith(`{"app":"qq","kind":"plugin","size":11,"nested":{},"tags":["a","b","c"],"app_sid":"app_123"}`)
	require.Equal(t, "InvalidParameter", decodeBody(t, w)["code_name"])
	byField := map[string]map[string]any{}
	for _, v := range violations(t, w) {
		byField[v["field"].(string)] = v
	}
	require.Equal(t, "enum.undefined", byField["kind"]["reason"])
	require.Equal(t, "one of: software, config", byField["kind"]["expected_format"])
	require.Equal(t, "value.out_of_range", byField["size"]["reason"])
	require.Equal(t, "field.required", byField["nested.name"]["reason"])
	require.Equal(t, "length.out_of_range", byField["tags"]["reason"])
	require.Equal(t, "format.invalid", byField["app_sid"]["reason"])

	w, _ = decodeWith(`{"app":"` + strings.Repeat("x", MaxRequestBytes) + `"}`)
	require.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
}

func TestSID(t *testing.T) {
	a, b := NewSID("app"), NewSID("app")
	require.NotEqual(t, a, b)
	require.True(t, IsSID("app", a))
	require.False(t, IsSID("ver", a))
	require.False(t, IsSID("app", "app_short"))
	require.Equal(t, strings.ToLower(a), a)
	require.Less(t, a, b, "monotonic within the same process")
}

func TestPageTokens(t *testing.T) {
	SetPageTokenKey("test-secret")
	p, err := PageRequest{}.Resolve("ListApps|q=")
	require.NoError(t, err)
	require.Equal(t, DefaultPageSize, p.Size)
	p, _ = PageRequest{PageSize: 500}.Resolve("ListApps|q=")
	require.Equal(t, MaxPageSize, p.Size)

	p, _ = PageRequest{PageSize: 2}.Resolve("ListApps|q=")
	resp := p.Response(3, nil)
	require.NotEmpty(t, resp.NextPageToken)
	require.Nil(t, resp.Total)
	require.Equal(t, []int{1, 2}, Trim([]int{1, 2, 3}, p))
	require.Equal(t, []int{}, Trim[int](nil, p))

	next, err := PageRequest{PageSize: 2, PageToken: resp.NextPageToken}.Resolve("ListApps|q=")
	require.NoError(t, err)
	require.Equal(t, 2, next.Offset)
	require.Empty(t, next.Response(1, nil).NextPageToken)

	_, err = PageRequest{PageToken: resp.NextPageToken}.Resolve("ListApps|q=other")
	require.Equal(t, "page.token_invalid", AsError(err).Reason, "token bound to its query")
	_, err = PageRequest{PageToken: resp.NextPageToken + "x"}.Resolve("ListApps|q=")
	require.Error(t, err, "tampered token")
	_, err = PageRequest{PageToken: "garbage"}.Resolve("ListApps|q=")
	require.Equal(t, CodeInvalidParameter, AsError(err).Code)
}

func TestSortSpec(t *testing.T) {
	spec := SortSpec{Columns: map[string]string{"name": "name", "created_at": "created_at"}, Default: []SortClause{{Field: "created_at", Order: "desc"}}}
	got, err := spec.OrderBy(nil)
	require.NoError(t, err)
	require.Equal(t, "created_at DESC, id ASC", got)
	got, err = spec.OrderBy([]SortClause{{Field: "name", Order: "asc"}})
	require.NoError(t, err)
	require.Equal(t, "name ASC, id ASC", got)
	_, err = spec.OrderBy([]SortClause{{Field: "password", Order: "asc"}})
	require.Equal(t, "sort[0].field", AsError(err).FieldViolations[0].Field)
	_, err = spec.OrderBy([]SortClause{{Field: "name"}})
	require.Error(t, err)
}

func TestFieldViolationJSONOmitsEmpty(t *testing.T) {
	b, _ := json.Marshal(FieldViolation{Field: "a", Reason: "r", Issue: "i"})
	require.False(t, bytes.Contains(b, []byte("suggestion")))
}
