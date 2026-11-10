package apispec

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	HeaderIdempotencyKey      = "Idempotency-Key"
	HeaderIdempotentReplayed  = "X-Idempotent-Replayed"
	idempotencyRecordTTL      = 24 * time.Hour
	idempotencyLease          = 30 * time.Second
	idempotencyPollInterval   = 200 * time.Millisecond
	idempotencyStatePending   = "pending"
	idempotencyStateCompleted = "completed"
)

var idempotencyKeyFormat = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,255}$`)

// IdempotencyRecord 是一次带幂等键请求的认领或完成状态。
type IdempotencyRecord struct {
	State       string `json:"state"`
	Owner       string `json:"owner,omitempty"`
	RequestHash string `json:"request_hash"`
	Status      int    `json:"status,omitempty"`
	Body        []byte `json:"body,omitempty"`
}

// IdempotencyStore 需在多副本间共享：生产用 Redis，单进程可用内存实现。
type IdempotencyStore interface {
	// Claim 原子认领 key；已存在时返回现有记录且 claimed=false。
	Claim(ctx context.Context, key string, rec IdempotencyRecord, lease time.Duration) (existing *IdempotencyRecord, claimed bool, err error)
	// Complete 在仍持有认领时写入终态。
	Complete(ctx context.Context, key, owner string, rec IdempotencyRecord, ttl time.Duration) error
	// Release 在仍持有认领时释放，允许重试。
	Release(ctx context.Context, key, owner string) error
}

// IdempotencyMiddleware 处理 optional_key / required_key 方法。
// 认证在其之前执行，因此每次执行与重放都会重新鉴权。
func IdempotencyMiddleware(store IdempotencyStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		policy := PolicyOf(c)
		key := c.GetHeader(HeaderIdempotencyKey)
		if key == "" {
			if policy.Idempotency == IdempotencyRequiredKey {
				AbortWithError(c, missingKeyError())
			} else {
				c.Next()
			}
			return
		}
		if !idempotencyKeyFormat.MatchString(key) {
			AbortWithError(c, InvalidParameter("idempotency.key_invalid", "Idempotency-Key is invalid", FieldViolation{
				Field: "header." + HeaderIdempotencyKey, Reason: "format.invalid", Issue: "key has an invalid format", ExpectedFormat: "1-255 chars of [A-Za-z0-9._:-]",
			}))
			return
		}

		body, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, MaxRequestBytes))
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		hash, ok := canonicalHash(body)
		if err != nil || !ok {
			c.Next() // 解码/校验错误发生在业务执行前，不建立幂等记录
			return
		}

		scope := sha256.Sum256([]byte(PrincipalOf(c).String() + "\x00" + MethodName(c) + "\x00" + c.FullPath() + "\x00" + key))
		storeKey := hex.EncodeToString(scope[:])
		owner := NewSID("idem")
		ctx := c.Request.Context()
		deadline := time.Now().Add(idempotencyLease)

		for {
			existing, claimed, err := store.Claim(ctx, storeKey, IdempotencyRecord{State: idempotencyStatePending, Owner: owner, RequestHash: hash}, idempotencyLease)
			if err != nil {
				AbortWithError(c, Unavailable("idempotency.store_unavailable", err))
				return
			}
			if claimed {
				break
			}
			if existing.RequestHash != hash {
				AbortWithError(c, NewError(CodeConflict, "idempotency.key_reused", "Idempotency-Key was used with a different request"))
				return
			}
			if existing.State == idempotencyStateCompleted {
				replay(c, existing)
				return
			}
			if time.Now().After(deadline) {
				AbortWithError(c, Unavailable("idempotency.in_progress", errors.New("previous request still in progress")).WithRetryAfter(1))
				return
			}
			select {
			case <-ctx.Done():
				c.Abort()
				return
			case <-time.After(idempotencyPollInterval):
			}
		}

		rec := &recorder{ResponseWriter: c.Writer}
		c.Writer = rec
		c.Next()

		// 存储操作不受客户端断开影响
		bg, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if cacheable(rec.Status(), rec.body.Bytes()) {
			_ = store.Complete(bg, storeKey, owner, IdempotencyRecord{
				State: idempotencyStateCompleted, RequestHash: hash, Status: rec.Status(), Body: rec.body.Bytes(),
			}, idempotencyRecordTTL)
		} else {
			_ = store.Release(bg, storeKey, owner)
		}
	}
}

func missingKeyError() *Error {
	e := NewError(CodeMissingParameter, "idempotency.key_missing", "Idempotency-Key header is required")
	e.FieldViolations = []FieldViolation{{Field: "header." + HeaderIdempotencyKey, Reason: "field.required", Issue: "header is required"}}
	return e
}

// canonicalHash 以规范化 JSON（键排序、数字原样）计算请求哈希，避免字段顺序/空白造成假冲突。
func canonicalHash(body []byte) (string, bool) {
	if len(bytes.TrimSpace(body)) == 0 {
		body = []byte("{}")
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var v any
	if dec.Decode(&v) != nil || dec.More() {
		return "", false
	}
	canonical, err := json.Marshal(v)
	if err != nil {
		return "", false
	}
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:]), true
}

// cacheable：成功与执行后确定的不可重试业务错误写入并重放；校验、鉴权与可重试错误释放认领。
func cacheable(status int, body []byte) bool {
	if status >= 200 && status < 300 {
		return true
	}
	var resp struct {
		CodeName string `json:"code_name"`
	}
	if json.Unmarshal(body, &resp) != nil {
		return false
	}
	switch resp.CodeName {
	case CodeResourceNotFound.Name, CodeFailedPrecondition.Name, CodeResourceAlreadyExists.Name, CodeUnprocessableEntity.Name:
		return true
	}
	return false
}

func replay(c *gin.Context, rec *IdempotencyRecord) {
	var body map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body, &body); err != nil {
		AbortWithError(c, Internal(err))
		return
	}
	m := meta(c) // meta 与当前响应头保持一致
	body["meta"], _ = json.Marshal(m)
	c.Header(HeaderRequestID, m.RequestID)
	c.Header(HeaderLatencyMS, strconv.Itoa(int(m.LatencyMS)))
	if m.TraceID != "" {
		c.Header(HeaderTraceID, m.TraceID)
	}
	c.Header(HeaderIdempotentReplayed, "true")
	c.JSON(rec.Status, body)
	c.Abort()
}

type recorder struct {
	gin.ResponseWriter
	body bytes.Buffer
}

func (r *recorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r *recorder) WriteString(s string) (int, error) {
	r.body.WriteString(s)
	return r.ResponseWriter.WriteString(s)
}

// RedisIdempotencyStore 使用 Redis 实现跨副本的原子认领。
type RedisIdempotencyStore struct {
	client redis.UniversalClient
	prefix string
}

func NewRedisIdempotencyStore(client redis.UniversalClient, prefix string) *RedisIdempotencyStore {
	return &RedisIdempotencyStore{client: client, prefix: prefix}
}

// 仅当记录仍由 owner 认领时才覆盖或删除。
var ownedSet = redis.NewScript(`
local cur = redis.call('GET', KEYS[1])
if not cur then return 0 end
if cjson.decode(cur)['owner'] ~= ARGV[1] then return 0 end
if ARGV[2] == '' then return redis.call('DEL', KEYS[1]) end
redis.call('SET', KEYS[1], ARGV[2], 'PX', ARGV[3])
return 1`)

func (s *RedisIdempotencyStore) Claim(ctx context.Context, key string, rec IdempotencyRecord, lease time.Duration) (*IdempotencyRecord, bool, error) {
	raw, _ := json.Marshal(rec)
	ok, err := s.client.SetNX(ctx, s.prefix+key, raw, lease).Result()
	if err != nil {
		return nil, false, err
	}
	if ok {
		return nil, true, nil
	}
	cur, err := s.client.Get(ctx, s.prefix+key).Bytes()
	if errors.Is(err, redis.Nil) {
		return s.Claim(ctx, key, rec, lease) // 记录恰好过期
	}
	if err != nil {
		return nil, false, err
	}
	var existing IdempotencyRecord
	if err := json.Unmarshal(cur, &existing); err != nil {
		return nil, false, err
	}
	return &existing, false, nil
}

func (s *RedisIdempotencyStore) Complete(ctx context.Context, key, owner string, rec IdempotencyRecord, ttl time.Duration) error {
	raw, _ := json.Marshal(rec)
	return ownedSet.Run(ctx, s.client, []string{s.prefix + key}, owner, raw, ttl.Milliseconds()).Err()
}

func (s *RedisIdempotencyStore) Release(ctx context.Context, key, owner string) error {
	return ownedSet.Run(ctx, s.client, []string{s.prefix + key}, owner, "", 0).Err()
}

// MemoryIdempotencyStore 是单进程实现（未启用 Redis 或测试时使用）。
type MemoryIdempotencyStore struct {
	mu      sync.Mutex
	records map[string]memoryRecord
}

type memoryRecord struct {
	rec     IdempotencyRecord
	expires time.Time
}

func NewMemoryIdempotencyStore() *MemoryIdempotencyStore {
	return &MemoryIdempotencyStore{records: map[string]memoryRecord{}}
}

func (s *MemoryIdempotencyStore) Claim(_ context.Context, key string, rec IdempotencyRecord, lease time.Duration) (*IdempotencyRecord, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cur, ok := s.records[key]; ok && time.Now().Before(cur.expires) {
		r := cur.rec
		return &r, false, nil
	}
	s.records[key] = memoryRecord{rec: rec, expires: time.Now().Add(lease)}
	return nil, true, nil
}

func (s *MemoryIdempotencyStore) Complete(_ context.Context, key, owner string, rec IdempotencyRecord, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cur, ok := s.records[key]; ok && cur.rec.Owner == owner {
		s.records[key] = memoryRecord{rec: rec, expires: time.Now().Add(ttl)}
	}
	return nil
}

func (s *MemoryIdempotencyStore) Release(_ context.Context, key, owner string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cur, ok := s.records[key]; ok && cur.rec.Owner == owner {
		delete(s.records, key)
	}
	return nil
}
