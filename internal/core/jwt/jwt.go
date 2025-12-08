package jwt

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/CIPFZ/gowebframe/internal/core/claims"
	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/modules/system/dto"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

// 定义错误
var (
	ErrTokenExpired     = errors.New("token已过期")
	ErrTokenNotValidYet = errors.New("token尚未激活")
	ErrTokenMalformed   = errors.New("这不是一个token")
	ErrTokenInvalid     = errors.New("无法处理此token")
)

type JWT struct {
	cfg         config.JWT
	logger      *zap.Logger
	redis       redis.UniversalClient
	signingKey  []byte
	issuer      string
	expiresTime time.Duration // 预解析的时间
	bufferTime  time.Duration // 预解析的时间

	// 并发控制，防止高并发下旧 Token 换新 Token 时产生多次计算
	sf *singleflight.Group
}

// NewJWT 初始化
func NewJWT(cfg config.JWT, logger *zap.Logger, redis redis.UniversalClient) *JWT {
	// 在初始化时解析时间，避免运行时重复解析
	ep, _ := parseDuration(cfg.ExpiresTime)
	bf, _ := parseDuration(cfg.BufferTime)

	return &JWT{
		cfg:         cfg,
		logger:      logger,
		redis:       redis,
		signingKey:  []byte(cfg.SigningKey),
		issuer:      cfg.Issuer,
		expiresTime: ep,
		bufferTime:  bf,
		sf:          &singleflight.Group{},
	}
}

// CreateClaims 构建 Claims
func (j *JWT) CreateClaims(baseClaims dto.BaseClaims) claims.CustomClaims {
	// BufferTime 是 int64 (秒)
	bfSeconds := int64(j.bufferTime / time.Second)

	return claims.CustomClaims{
		BaseClaims: baseClaims,
		BufferTime: bfSeconds,
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{j.cfg.Issuer},                               // 建议放到配置中
			NotBefore: jwt.NewNumericDate(time.Now().Add(-1000 * time.Millisecond)), // 容错1秒
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.expiresTime)),
			Issuer:    j.issuer,
		},
	}
}

// CreateToken 创建 Token
func (j *JWT) CreateToken(claims claims.CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.signingKey)
}

// ParseToken 解析 Token (v5 版本)
func (j *JWT) ParseToken(tokenString string) (*claims.CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.signingKey, nil
	})

	if err != nil {
		// ✨ v5 错误处理方式
		return nil, err
	}

	if claims, ok := token.Claims.(*claims.CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

// ResolveToken 处理 Token 续期逻辑 (GVA 的核心逻辑)
// 如果 Token 过期但在 BufferTime 内，则自动刷新
func (j *JWT) ResolveToken(ctx context.Context, oldToken string, claims *claims.CustomClaims) (string, *claims.CustomClaims, error) {
	// 1. 如果没有过期，直接返回
	if claims.ExpiresAt.Unix()-time.Now().Unix() > claims.BufferTime {
		return oldToken, claims, nil
	}

	// 2. 如果已经过期且超过缓冲时间，或者不符合刷新条件，直接返回旧的（让中间件报过期）
	// 注意：jwt.ParseWithClaims 已经验证了 ExpiresAt，如果过期会报错 ErrTokenExpired
	// 这里我们处理的是 "即将过期" 或者 "刚过期但在宽限期内" 的情况
	// 这里的逻辑取决于你的 JWT 中间件是否允许 "ErrTokenExpired" 进入到这一步

	// 使用 singleflight 避免并发刷新
	// Key 使用 token 签名部分，确保唯一
	newToken, err, _ := j.sf.Do("JWT:"+oldToken, func() (interface{}, error) {
		return j.createTokenByOldToken(ctx, oldToken, *claims)
	})

	if err != nil {
		return "", nil, err
	}

	// 解析新 Token 返回新的 Claims
	newClaims, err := j.ParseToken(newToken.(string))
	return newToken.(string), newClaims, err
}

// createTokenByOldToken (内部方法) 执行真正的刷新逻辑
func (j *JWT) createTokenByOldToken(ctx context.Context, oldToken string, c claims.CustomClaims) (string, error) {
	// 1. 检查 Redis 黑名单 (防止已登出的 Token 复活)
	// 这里的 Key 策略要统一，建议 use: uuid
	isBlack, _ := j.redis.Get(ctx, "jwt_black:"+oldToken).Result()
	if isBlack != "" {
		return "", errors.New("token已被注销")
	}

	// 2. 检查是否已经在 Redis 中有并发请求生成的新 Token (可选，GVA 策略)
	// (singleflight 已经处理了进程内的并发，Redis 处理多实例并发)

	// 3. 生成新 Token
	newClaims := j.CreateClaims(c.BaseClaims)
	newToken, err := j.CreateToken(newClaims)
	if err != nil {
		return "", err
	}

	// 4. (可选) 将旧 Token 加入黑名单，防止重放，但要给一个短暂的过渡期
	// j.SetBlacklist(ctx, oldToken, 10*time.Second)

	return newToken, nil
}

// SetBlacklist 将 Token 加入黑名单 (登出时使用)
func (j *JWT) SetBlacklist(ctx context.Context, token string, expiration time.Duration) error {
	return j.redis.Set(ctx, "jwt_black:"+token, "1", expiration).Err()
}

// IsBlacklist 检查是否在黑名单
func (j *JWT) IsBlacklist(ctx context.Context, token string) bool {
	val, _ := j.redis.Get(ctx, "jwt_black:"+token).Result()
	return val != ""
}

// ---------------- 辅助函数 ----------------

// parseDuration 解析字符串时间，支持 "d" (天)
func parseDuration(d string) (time.Duration, error) {
	d = strings.TrimSpace(d)
	dr, err := time.ParseDuration(d)
	if err == nil {
		return dr, nil
	}
	if strings.HasSuffix(d, "d") {
		days := strings.TrimSuffix(d, "d")
		v, err := strconv.Atoi(days)
		if err == nil {
			return time.Hour * 24 * time.Duration(v), nil
		}
	}
	return 0, errors.New("invalid duration")
}
