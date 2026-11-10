package apispec

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"sync"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// PageRequest 是所有 List 请求的分页参数。
type PageRequest struct {
	PageSize  int32  `json:"page_size"`  // 默认 20，上限 100；<=0 使用默认值
	PageToken string `json:"page_token"` // 不透明 cursor；为空表示第一页
}

// PageResponse 是所有 List 响应的分页信息。
type PageResponse struct {
	NextPageToken string `json:"next_page_token,omitempty"`
	Total         *int32 `json:"total,omitempty"` // 未计算时省略
	PageSize      int32  `json:"page_size"`
}

// ListData 是 List 响应的 data。
type ListData[T any] struct {
	Items []T          `json:"items"`
	Page  PageResponse `json:"page"`
}

var (
	pageKeyMu sync.RWMutex
	pageKey   = []byte("apispec-page-token-dev-key")
)

// SetPageTokenKey 设置 page token 的签名密钥（启动时由 JWT 密钥派生）。
func SetPageTokenKey(secret string) {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("apispec/page-token/v1"))
	pageKeyMu.Lock()
	pageKey = mac.Sum(nil)
	pageKeyMu.Unlock()
}

type pageCursor struct {
	Version int    `json:"v"`
	Offset  int    `json:"o"`
	Scope   string `json:"s"` // 绑定 List 方法与过滤条件，换条件复用 token 视为非法
}

// Page 是解析后的分页位置。
type Page struct {
	Size   int
	Offset int
	scope  string
}

// Resolve 规范化页大小并校验 page token。scope 应包含方法名与过滤/排序条件的稳定表示。
func (p PageRequest) Resolve(scope string) (Page, error) {
	size := int(p.PageSize)
	if size <= 0 {
		size = DefaultPageSize
	}
	size = min(size, MaxPageSize)
	page := Page{Size: size, scope: scopeHash(scope)}
	if p.PageToken == "" {
		return page, nil
	}
	cur, ok := parsePageToken(p.PageToken)
	if !ok || cur.Version != 1 || cur.Offset < 0 || cur.Scope != page.scope {
		return Page{}, InvalidParameter("page.token_invalid", "page_token is invalid or does not match the query", FieldViolation{
			Field: "page.page_token", Reason: "format.invalid", Issue: "token is malformed, expired or bound to a different query",
		})
	}
	page.Offset = cur.Offset
	return page, nil
}

// Response 根据本页实际条数生成分页信息；fetched 应为按 Size+1 查询得到的条数。
func (p Page) Response(fetched int, total *int32) PageResponse {
	resp := PageResponse{PageSize: int32(p.Size), Total: total}
	if fetched > p.Size {
		resp.NextPageToken = signPageToken(pageCursor{Version: 1, Offset: p.Offset + p.Size, Scope: p.scope})
	}
	return resp
}

// Trim 截掉为判断下一页多取的一条。
func Trim[T any](items []T, p Page) []T {
	if items == nil {
		return []T{}
	}
	if len(items) > p.Size {
		return items[:p.Size]
	}
	return items
}

func scopeHash(scope string) string {
	sum := sha256.Sum256([]byte(scope))
	return base64.RawURLEncoding.EncodeToString(sum[:9])
}

func signPageToken(cur pageCursor) string {
	payload, _ := json.Marshal(cur)
	body := base64.RawURLEncoding.EncodeToString(payload)
	return body + "." + pageSignature(body)
}

func parsePageToken(token string) (pageCursor, bool) {
	body, sig, ok := strings.Cut(token, ".")
	if !ok || !hmac.Equal([]byte(sig), []byte(pageSignature(body))) {
		return pageCursor{}, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(body)
	if err != nil {
		return pageCursor{}, false
	}
	var cur pageCursor
	if json.Unmarshal(raw, &cur) != nil {
		return pageCursor{}, false
	}
	return cur, true
}

func pageSignature(body string) string {
	pageKeyMu.RLock()
	mac := hmac.New(sha256.New, pageKey)
	pageKeyMu.RUnlock()
	mac.Write([]byte(body))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)[:16])
}
