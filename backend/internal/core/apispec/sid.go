package apispec

import (
	"crypto/rand"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	sidMu      sync.Mutex
	sidEntropy = ulid.Monotonic(rand.Reader, 0)
)

// NewSID 生成带类型前缀的不透明资源标识，如 app_01JB8Z…。
func NewSID(prefix string) string {
	sidMu.Lock()
	id := ulid.MustNew(ulid.Timestamp(time.Now()), sidEntropy)
	sidMu.Unlock()
	return prefix + "_" + strings.ToLower(id.String())
}

// IsSID 校验标识是否为指定前缀的合法 sid。
func IsSID(prefix, sid string) bool {
	rest, ok := strings.CutPrefix(sid, prefix+"_")
	if !ok || len(rest) != ulid.EncodedSize {
		return false
	}
	_, err := ulid.ParseStrict(strings.ToUpper(rest))
	return err == nil
}
