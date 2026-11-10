package service

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"

	tokenCore "github.com/CIPFZ/gowebframe/internal/core/token"
	"github.com/CIPFZ/gowebframe/internal/modules/system/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"github.com/CIPFZ/gowebframe/internal/modules/system/repository"
)

type IApiTokenService interface {
	CreateApiToken(ctx context.Context, actor TokenActor, req dto.CreateApiTokenReq) (*dto.ApiTokenSecretResponse, error)
	GetApiTokenList(ctx context.Context, req dto.SearchApiTokenReq) ([]dto.ApiTokenResponse, int64, error)
	GetApiTokenDetail(ctx context.Context, actor TokenActor, id uint) (*dto.ApiTokenResponse, error)
	UpdateApiToken(ctx context.Context, actor TokenActor, req dto.UpdateApiTokenReq) error
	DeleteApiToken(ctx context.Context, actor TokenActor, req dto.DeleteApiTokenReq) error
	ResetApiToken(ctx context.Context, actor TokenActor, id uint) (*dto.ApiTokenSecretResponse, error)
	EnableApiToken(ctx context.Context, actor TokenActor, id uint) error
	DisableApiToken(ctx context.Context, actor TokenActor, id uint) error
	EffectiveLimits(ctx context.Context, userID uint) (dto.TokenLimits, error)
	CountTokens(ctx context.Context, userID uint) (int64, error)
	GetLimitDetail(ctx context.Context, userID uint) (*dto.TokenLimitDetail, error)
	UpdateLimit(ctx context.Context, req dto.TokenLimitSettings) error
}

// TokenActor 是操作 Token 的人。Manage（管理员）可以管理所有人的 Token，且不受申请限制约束。
type TokenActor struct {
	UserID uint
	Manage bool
}

// TokenPermissionCheck 判断用户现在能否调用某接口（所有角色的权限取并集）。
type TokenPermissionCheck func(ctx context.Context, userID uint, path, method string) (bool, error)

type ApiTokenService struct {
	tokenRepo repository.IApiTokenRepository
	allows    TokenPermissionCheck
}

func NewApiTokenService(tokenRepo repository.IApiTokenRepository, allows TokenPermissionCheck) IApiTokenService {
	return &ApiTokenService{tokenRepo: tokenRepo, allows: allows}
}

// resolvePermissions 决定 Token 的接口：跟随主人时不存接口；否则逐个校验（本人只能选自己有的）。
func (s *ApiTokenService) resolvePermissions(ctx context.Context, actor TokenActor, ownerID uint, followOwner bool, apiIDs []uint) ([]model.SysApi, error) {
	if followOwner {
		return []model.SysApi{}, nil
	}
	apis, err := s.resolveApis(ctx, apiIDs)
	if err != nil {
		return nil, err
	}
	if actor.Manage {
		return apis, nil
	}
	for _, api := range apis {
		ok, err := s.allows(ctx, ownerID, api.Path, api.Method)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errors.New("token.permissionDenied")
		}
	}
	return apis, nil
}

// checkQuota 校验额度不超过限制（管理员不受约束），并把没填的额度补成限制值。
func checkQuota[T int | int64](requested *T, limit T, manage bool) error {
	if *requested <= 0 {
		*requested = limit
		return nil
	}
	if *requested > limit && !manage {
		return errors.New("token.quotaTooHigh")
	}
	return nil
}

func (s *ApiTokenService) CreateApiToken(ctx context.Context, actor TokenActor, req dto.CreateApiTokenReq) (*dto.ApiTokenSecretResponse, error) {
	if strings.TrimSpace(req.Name) == "" || utf8.RuneCountInString(req.Name) > 100 || utf8.RuneCountInString(req.Description) > 255 {
		return nil, errors.New("validation.invalid")
	}
	limits, err := s.EffectiveLimits(ctx, actor.UserID)
	if err != nil {
		return nil, err
	}
	if !actor.Manage {
		n, err := s.tokenRepo.CountByOwner(ctx, actor.UserID)
		if err != nil {
			return nil, err
		}
		if int(n) >= limits.MaxTokens {
			return nil, errors.New("token.tooMany")
		}
	}
	expiresAt := time.Now().AddDate(0, 0, limits.MaxTTLDays) // 只填名字时按最长有效期
	if req.ExpiresAt != nil || req.NeverExpire {
		parsed, err := parseExpiresAt(req.ExpiresAt, req.NeverExpire)
		if err != nil {
			return nil, err
		}
		if err := checkExpiry(*parsed, limits, actor.Manage); err != nil {
			return nil, err
		}
		expiresAt = *parsed
	}
	if err := checkQuota(&req.MaxConcurrency, limits.MaxConcurrency, actor.Manage); err != nil {
		return nil, err
	}
	apis, err := s.resolvePermissions(ctx, actor, actor.UserID, req.FollowOwner, req.ApiIds)
	if err != nil {
		return nil, err
	}

	rawToken := tokenCore.GenerateRawToken()
	token := &model.SysApiToken{
		TokenHash:      tokenCore.HashToken(rawToken),
		TokenPrefix:    buildTokenPrefix(rawToken),
		Name:           strings.TrimSpace(req.Name),
		Description:    strings.TrimSpace(req.Description),
		ExpiresAt:      &expiresAt,
		MaxConcurrency: req.MaxConcurrency,
		Enabled:        true,
		CreatedBy:      actor.UserID,
		FollowOwner:    req.FollowOwner,
		Apis:           apis,
	}

	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return nil, err
	}

	return buildSecretResponse(token, rawToken), nil
}

func checkExpiry(expiresAt time.Time, limits dto.TokenLimits, manage bool) error {
	if !manage && expiresAt.After(time.Now().AddDate(0, 0, limits.MaxTTLDays)) {
		return errors.New("token.expiryTooLong")
	}
	return nil
}

func (s *ApiTokenService) GetApiTokenList(ctx context.Context, req dto.SearchApiTokenReq) ([]dto.ApiTokenResponse, int64, error) {
	list, total, err := s.tokenRepo.GetList(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	ownerIDs := make([]uint, 0, len(list))
	for _, item := range list {
		ownerIDs = append(ownerIDs, item.CreatedBy)
	}
	owners, err := s.tokenRepo.OwnerNames(ctx, ownerIDs)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]dto.ApiTokenResponse, 0, len(list))
	for _, item := range list {
		r := buildApiTokenResponse(&item)
		r.OwnerName = owners[item.CreatedBy]
		resp = append(resp, r)
	}
	return resp, total, nil
}

// ownedToken 取一个 Token；不是本人的（且不是管理员）与不存在不可区分。
func (s *ApiTokenService) ownedToken(ctx context.Context, actor TokenActor, id uint) (*model.SysApiToken, error) {
	token, err := s.tokenRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !actor.Manage && token.CreatedBy != actor.UserID {
		return nil, gorm.ErrRecordNotFound
	}
	return token, nil
}

func (s *ApiTokenService) GetApiTokenDetail(ctx context.Context, actor TokenActor, id uint) (*dto.ApiTokenResponse, error) {
	token, err := s.ownedToken(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	resp := buildApiTokenResponse(token)
	return &resp, nil
}

func (s *ApiTokenService) UpdateApiToken(ctx context.Context, actor TokenActor, req dto.UpdateApiTokenReq) error {
	token, err := s.ownedToken(ctx, actor, req.ID)
	if err != nil {
		return err
	}

	if strings.TrimSpace(req.Name) == "" || utf8.RuneCountInString(req.Name) > 100 || utf8.RuneCountInString(req.Description) > 255 {
		return errors.New("validation.invalid")
	}
	limits, err := s.EffectiveLimits(ctx, token.CreatedBy)
	if err != nil {
		return err
	}
	expiresAt, err := parseExpiresAt(req.ExpiresAt, req.NeverExpire)
	if err != nil {
		return err
	}
	if err := checkExpiry(*expiresAt, limits, actor.Manage); err != nil {
		return err
	}
	apis, err := s.resolvePermissions(ctx, actor, token.CreatedBy, req.FollowOwner, req.ApiIds)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"name":         strings.TrimSpace(req.Name),
		"description":  strings.TrimSpace(req.Description),
		"expires_at":   expiresAt,
		"follow_owner": req.FollowOwner,
	}
	// 并发上限：请求未携带时保持原值
	if req.MaxConcurrency > 0 {
		if err := checkQuota(&req.MaxConcurrency, limits.MaxConcurrency, actor.Manage); err != nil {
			return err
		}
		updates["max_concurrency"] = req.MaxConcurrency
	}
	return s.tokenRepo.UpdateWithAPIs(ctx, token, updates, apis)
}

func (s *ApiTokenService) DeleteApiToken(ctx context.Context, actor TokenActor, req dto.DeleteApiTokenReq) error {
	ids := make([]uint, 0, len(req.IDs)+1)
	if req.ID != 0 {
		ids = append(ids, req.ID)
	}
	ids = append(ids, req.IDs...)
	if len(ids) == 0 {
		return errors.New("未选择要删除的 token")
	}
	for _, id := range ids {
		if _, err := s.ownedToken(ctx, actor, id); err != nil {
			return err
		}
	}
	return s.tokenRepo.DeleteByIDs(ctx, ids)
}

func (s *ApiTokenService) ResetApiToken(ctx context.Context, actor TokenActor, id uint) (*dto.ApiTokenSecretResponse, error) {
	token, err := s.ownedToken(ctx, actor, id)
	if err != nil {
		return nil, err
	}

	rawToken := tokenCore.GenerateRawToken()
	updates := map[string]interface{}{
		"token_hash":   tokenCore.HashToken(rawToken),
		"token_prefix": buildTokenPrefix(rawToken),
	}
	if err := s.tokenRepo.UpdateColumns(ctx, id, updates); err != nil {
		return nil, err
	}

	token.TokenHash = updates["token_hash"].(string)
	token.TokenPrefix = updates["token_prefix"].(string)
	return buildSecretResponse(token, rawToken), nil
}

func (s *ApiTokenService) EnableApiToken(ctx context.Context, actor TokenActor, id uint) error {
	if _, err := s.ownedToken(ctx, actor, id); err != nil {
		return err
	}
	return s.tokenRepo.UpdateColumns(ctx, id, map[string]interface{}{"enabled": true})
}

func (s *ApiTokenService) DisableApiToken(ctx context.Context, actor TokenActor, id uint) error {
	if _, err := s.ownedToken(ctx, actor, id); err != nil {
		return err
	}
	return s.tokenRepo.UpdateColumns(ctx, id, map[string]interface{}{"enabled": false})
}

func (s *ApiTokenService) CountTokens(ctx context.Context, userID uint) (int64, error) {
	return s.tokenRepo.CountByOwner(ctx, userID)
}

// 全局默认没有设置时的取值。
var fallbackLimits = dto.TokenLimits{MaxTokens: 5, MaxTTLDays: 180, MaxConcurrency: 5}

func applyLimit(base dto.TokenLimits, l *model.SysTokenLimit) dto.TokenLimits {
	if l.MaxTokens != nil {
		base.MaxTokens = *l.MaxTokens
	}
	if l.MaxTTLDays != nil {
		base.MaxTTLDays = *l.MaxTTLDays
	}
	if l.MaxConcurrency != nil {
		base.MaxConcurrency = *l.MaxConcurrency
	}
	return base
}

func (s *ApiTokenService) defaultLimits(ctx context.Context) (dto.TokenLimits, error) {
	global, err := s.tokenRepo.GetLimit(ctx, 0)
	if err != nil {
		return dto.TokenLimits{}, err
	}
	return applyLimit(fallbackLimits, global), nil
}

// EffectiveLimits 返回用户实际生效的限制：单独设置过的项用设置值，其余跟随全局默认。
func (s *ApiTokenService) EffectiveLimits(ctx context.Context, userID uint) (dto.TokenLimits, error) {
	base, err := s.defaultLimits(ctx)
	if err != nil {
		return base, err
	}
	own, err := s.tokenRepo.GetLimit(ctx, userID)
	if err != nil {
		return base, err
	}
	return applyLimit(base, own), nil
}

func (s *ApiTokenService) GetLimitDetail(ctx context.Context, userID uint) (*dto.TokenLimitDetail, error) {
	own, err := s.tokenRepo.GetLimit(ctx, userID)
	if err != nil {
		return nil, err
	}
	base, err := s.defaultLimits(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.TokenLimitDetail{
		TokenLimitSettings: dto.TokenLimitSettings{UserID: userID, MaxTokens: own.MaxTokens, MaxTTLDays: own.MaxTTLDays, MaxConcurrency: own.MaxConcurrency},
		Effective:          applyLimit(base, own),
		Default:            base,
	}, nil
}

// UpdateLimit 保存一份设置，只影响之后的申请与修改，已有的 Token 不变。
func (s *ApiTokenService) UpdateLimit(ctx context.Context, req dto.TokenLimitSettings) error {
	return s.tokenRepo.SaveLimit(ctx, &model.SysTokenLimit{UserID: req.UserID, MaxTokens: req.MaxTokens, MaxTTLDays: req.MaxTTLDays, MaxConcurrency: req.MaxConcurrency})
}

func (s *ApiTokenService) resolveApis(ctx context.Context, apiIDs []uint) ([]model.SysApi, error) {
	if len(apiIDs) == 0 {
		return nil, errors.New("请至少选择一个 API")
	}
	unique := make([]uint, 0, len(apiIDs))
	seen := map[uint]bool{}
	for _, id := range apiIDs {
		if !seen[id] {
			unique = append(unique, id)
			seen[id] = true
		}
	}
	apiIDs = unique
	apis, err := s.tokenRepo.LoadApisByIDs(ctx, apiIDs)
	if err != nil {
		return nil, err
	}
	if len(apis) != len(apiIDs) {
		return nil, errors.New("存在无效的 API 选择")
	}
	return apis, nil
}

func parseExpiresAt(raw *string, neverExpire bool) (*time.Time, error) {
	if neverExpire || raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, errors.New("token.expiryRequired")
	}
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*raw))
	if err != nil {
		return nil, errors.New("过期时间格式错误")
	}
	if parsed.Before(time.Now()) {
		return nil, errors.New("过期时间必须晚于当前时间")
	}
	return &parsed, nil
}

func buildTokenPrefix(raw string) string {
	if len(raw) <= 8 {
		return raw
	}
	return raw[:8]
}

func buildApiTokenResponse(token *model.SysApiToken) dto.ApiTokenResponse {
	resp := dto.ApiTokenResponse{
		ID:             token.ID,
		TokenPrefix:    token.TokenPrefix,
		Name:           token.Name,
		Description:    token.Description,
		NeverExpire:    token.ExpiresAt == nil,
		MaxConcurrency: token.MaxConcurrency,
		Enabled:        token.Enabled,
		CreatedAt:      token.CreatedAt.Format(time.RFC3339),
		CreatedBy:      token.CreatedBy,
		FollowOwner:    token.FollowOwner,
		Apis:           buildApiItems(token.Apis),
	}
	if token.ExpiresAt != nil {
		value := token.ExpiresAt.Format(time.RFC3339)
		resp.ExpiresAt = &value
	}
	if token.LastUsedAt != nil {
		value := token.LastUsedAt.Format(time.RFC3339)
		resp.LastUsedAt = &value
	}
	return resp
}

func buildSecretResponse(token *model.SysApiToken, rawToken string) *dto.ApiTokenSecretResponse {
	resp := buildApiTokenResponse(token)
	return &dto.ApiTokenSecretResponse{
		ApiTokenResponse: resp,
		Token:            rawToken,
	}
}

func buildApiItems(apis []model.SysApi) []dto.ApiSimpleItem {
	items := make([]dto.ApiSimpleItem, 0, len(apis))
	for _, api := range apis {
		items = append(items, dto.ApiSimpleItem{
			ID:          api.ID,
			Path:        api.Path,
			Method:      api.Method,
			ApiGroup:    api.ApiGroup,
			Description: api.Description,
		})
	}
	return items
}

var _ = gorm.ErrRecordNotFound
