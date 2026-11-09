package apispec

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

var (
	readPolicy  = MethodPolicy{Action: ActionRead, Sensitivity: SensitivityRead, Visibility: VisibilitySelected, Idempotency: IdempotencyNatural, Authentication: AuthAnonymous, Command: "orders order get"}
	writePolicy = MethodPolicy{Action: ActionCreate, Sensitivity: SensitivityWriteSafe, Visibility: VisibilitySelected, Idempotency: IdempotencyOptionalKey, Authentication: AuthToken, Command: "orders report create"}
)

func TestValidateMethodInvariants(t *testing.T) {
	require.NoError(t, validateMethod("orders", "get", "GetOrder", readPolicy))
	require.NoError(t, validateMethod("reports", "create", "CreateReport", writePolicy))
	require.NoError(t, validateMethod("jobs", "create-from-templates", "CreateFromTemplateJob", writePolicy))

	admin := MethodPolicy{Action: ActionAdmin, Sensitivity: SensitivityWriteSensitive, Visibility: VisibilityPublic, Idempotency: IdempotencyOptionalKey, Authentication: AuthUser}
	require.NoError(t, validateMethod("recipe-revisions", "update", "UpdateRecipeRevision", admin))

	cases := map[string]struct {
		resource, verb, name string
		mutate               func(*MethodPolicy)
		base                 MethodPolicy
	}{
		"unknown verb":            {"orders", "fetch", "FetchOrder", nil, readPolicy},
		"bad alias verb":          {"tasks", "claim", "ClaimTask", nil, writePolicy},
		"singular resource":       {"order", "get", "GetOrder", nil, readPolicy},
		"snake resource":          {"order_items", "get", "GetOrderItem", nil, readPolicy},
		"name mismatch":           {"orders", "get", "FetchOrder", nil, readPolicy},
		"read with write sens":    {"orders", "get", "GetOrder", func(p *MethodPolicy) { p.Sensitivity = SensitivityWriteSafe }, readPolicy},
		"read with key":           {"orders", "get", "GetOrder", func(p *MethodPolicy) { p.Idempotency = IdempotencyOptionalKey }, readPolicy},
		"create marked read":      {"orders", "create", "CreateOrder", nil, readPolicy},
		"list with side effects":  {"orders", "list", "ListOrders", nil, writePolicy},
		"delete not destructive":  {"orders", "delete", "DeleteOrder", func(p *MethodPolicy) { p.Action = ActionDelete }, writePolicy},
		"report not internal":     {"orders", "create", "CreateOrder", func(p *MethodPolicy) { p.Action = ActionReport }, writePolicy},
		"admin selected":          {"orders", "update", "UpdateOrder", func(p *MethodPolicy) { p.Visibility = VisibilitySelected }, admin},
		"admin via token":         {"orders", "update", "UpdateOrder", func(p *MethodPolicy) { p.Authentication = AuthToken }, admin},
		"anonymous write":         {"orders", "create", "CreateOrder", func(p *MethodPolicy) { p.Authentication = AuthAnonymous }, writePolicy},
		"selected without cmd":    {"orders", "get", "GetOrder", func(p *MethodPolicy) { p.Command = "" }, readPolicy},
		"selected write no idem":  {"orders", "create", "CreateOrder", func(p *MethodPolicy) { p.Idempotency = IdempotencyNone }, writePolicy},
		"zero policy fails close": {"orders", "get", "GetOrder", func(p *MethodPolicy) { *p = MethodPolicy{} }, readPolicy},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			p := tc.base
			if tc.mutate != nil {
				tc.mutate(&p)
			}
			require.Error(t, validateMethod(tc.resource, tc.verb, tc.name, p))
		})
	}
}

func tokenAuth(c *gin.Context) {
	id := c.GetHeader("X-Test-Token")
	if id == "" {
		AbortWithError(c, Unauthorized("auth.token_missing", "API token is required"))
		return
	}
	SetPrincipal(c, Principal{Kind: PrincipalToken, ID: id})
	c.Next()
}

func newTestRouter(store IdempotencyStore) (*gin.Engine, *Router) {
	e := gin.New()
	r := NewRouter(e.Group("/api/v1/orders"), RouterOptions{
		Authenticators: map[Authentication]gin.HandlerFunc{AuthToken: tokenAuth},
		Idempotency:    IdempotencyMiddleware(store),
	})
	return e, r
}

func TestRouterRegistration(t *testing.T) {
	e, r := newTestRouter(NewMemoryIdempotencyStore())
	r.Handle("orders", "get", "GetOrder", readPolicy, func(c *gin.Context) {
		require.Equal(t, "GetOrder", MethodName(c))
		require.Equal(t, PrincipalAnonymous, PrincipalOf(c).Kind)
		WriteOK(c, gin.H{"object": "order"})
	})
	require.Panics(t, func() { r.Handle("orders", "search", "SearchApps", readPolicy, nil) }, "duplicate command")
	require.Panics(t, func() { r.Handle("orders", "update", "UpdateOrder", MethodPolicy{}, nil) }, "invalid policy fails closed")
	userOnly := writePolicy
	userOnly.Authentication, userOnly.Command = AuthUser, "orders order update"
	require.Panics(t, func() { r.Handle("orders", "update", "UpdateOrder", userOnly, nil) }, "missing authenticator")

	methods := r.Methods()
	require.Len(t, methods, 1)
	require.Equal(t, "/api/v1/orders/orders/get", methods[0].Path)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/orders/orders/get", nil))
	require.Equal(t, http.StatusOK, w.Code)
	w = httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/orders/orders/get", nil))
	require.NotEqual(t, http.StatusOK, w.Code, "only POST is routed")
}

func call(e *gin.Engine, body string, headers ...string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/reports/create", strings.NewReader(body))
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)
	return w
}

type reportReq struct {
	Result string `json:"result" validate:"required,oneof=succeeded failed"`
}

func idempotencySuite(t *testing.T, store IdempotencyStore) {
	e, r := newTestRouter(store)
	var calls atomic.Int32
	var failNext atomic.Bool
	r.Handle("reports", "create", "CreateReport", writePolicy, func(c *gin.Context) {
		var req reportReq
		if err := Decode(c, &req); err != nil {
			WriteError(c, err)
			return
		}
		n := calls.Add(1)
		time.Sleep(50 * time.Millisecond)
		if failNext.CompareAndSwap(true, false) {
			WriteError(c, Unavailable("db.unavailable", nil)) // 可重试：释放认领
			return
		}
		WriteOK(c, gin.H{"sid": NewSID("rpt"), "call": n})
	})
	tok := []string{"X-Test-Token", "7"}

	// 无凭据：认证在幂等之前失败
	w := call(e, `{"result":"succeeded"}`, "Idempotency-Key", "k1")
	require.Equal(t, http.StatusUnauthorized, w.Code)

	// 无 key：optional_key 正常执行
	w = call(e, `{"result":"succeeded"}`, tok...)
	require.Equal(t, http.StatusOK, w.Code)
	require.EqualValues(t, 1, calls.Load())

	// 首次执行 + 重放（字段顺序/空白不同仍视为同一请求）
	first := call(e, `{"result":"succeeded"}`, append(tok, "Idempotency-Key", "k1")...)
	require.Equal(t, http.StatusOK, first.Code)
	replayed := call(e, `{ "result" : "succeeded" }`, append(tok, "Idempotency-Key", "k1")...)
	require.Equal(t, http.StatusOK, replayed.Code)
	require.Equal(t, "true", replayed.Header().Get(HeaderIdempotentReplayed))
	require.Equal(t, decodeBody(t, first)["data"], decodeBody(t, replayed)["data"])
	require.Equal(t, replayed.Header().Get(HeaderRequestID), decodeBody(t, replayed)["meta"].(map[string]any)["request_id"])
	require.EqualValues(t, 2, calls.Load())

	// 同 key 不同请求 → Conflict
	w = call(e, `{"result":"failed"}`, append(tok, "Idempotency-Key", "k1")...)
	require.Equal(t, http.StatusConflict, w.Code)
	require.Equal(t, "idempotency.key_reused", decodeBody(t, w)["error"].(map[string]any)["reason"])

	// 不同身份同 key 互不影响
	w = call(e, `{"result":"succeeded"}`, "X-Test-Token", "8", "Idempotency-Key", "k1")
	require.Empty(t, w.Header().Get(HeaderIdempotentReplayed))

	// 校验失败不写缓存
	before := calls.Load()
	w = call(e, `{"result":"maybe"}`, append(tok, "Idempotency-Key", "k2")...)
	require.Equal(t, http.StatusBadRequest, w.Code)
	w = call(e, `{"result":"maybe"}`, append(tok, "Idempotency-Key", "k2")...)
	require.Empty(t, w.Header().Get(HeaderIdempotentReplayed))
	require.Equal(t, before, calls.Load())

	// 非法 key
	w = call(e, `{}`, append(tok, "Idempotency-Key", "bad key!")...)
	require.Equal(t, "idempotency.key_invalid", decodeBody(t, w)["error"].(map[string]any)["reason"])

	// 并发同 key：只执行一次，其余等待复用结果
	before = calls.Load()
	var wg sync.WaitGroup
	results := make([]*httptest.ResponseRecorder, 5)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i] = call(e, `{"result":"succeeded"}`, append(tok, "Idempotency-Key", "k3")...)
		}(i)
	}
	wg.Wait()
	require.Equal(t, before+1, calls.Load())
	for _, res := range results {
		require.Equal(t, http.StatusOK, res.Code)
		require.Equal(t, decodeBody(t, results[0])["data"], decodeBody(t, res)["data"])
	}

	// 可重试错误释放认领，重试会再次执行
	failNext.Store(true)
	w = call(e, `{"result":"failed"}`, append(tok, "Idempotency-Key", "k4")...)
	require.Equal(t, http.StatusServiceUnavailable, w.Code)
	before = calls.Load()
	w = call(e, `{"result":"failed"}`, append(tok, "Idempotency-Key", "k4")...)
	require.Equal(t, http.StatusOK, w.Code)
	require.Empty(t, w.Header().Get(HeaderIdempotentReplayed))
	require.Equal(t, before+1, calls.Load())
}

func TestIdempotencyMemoryStore(t *testing.T) { idempotencySuite(t, NewMemoryIdempotencyStore()) }

func TestIdempotencyRedisStore(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	idempotencySuite(t, NewRedisIdempotencyStore(client, "orders:idem:"))
	require.NotEmpty(t, mr.Keys())
	for _, k := range mr.Keys() {
		require.True(t, strings.HasPrefix(k, "orders:idem:"))
		require.Greater(t, mr.TTL(k), 23*time.Hour)
	}
}

func TestIdempotencyRequiredKey(t *testing.T) {
	e, r := newTestRouter(NewMemoryIdempotencyStore())
	p := writePolicy
	p.Idempotency = IdempotencyRequiredKey
	r.Handle("reports", "create", "CreateReport", p, func(c *gin.Context) { WriteOK(c, nil) })
	w := call(e, `{}`, "X-Test-Token", "7")
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, "MissingParameter", decodeBody(t, w)["code_name"])
}
