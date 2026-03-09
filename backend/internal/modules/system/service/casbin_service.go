package service

import (
	"context"

	"github.com/CIPFZ/gowebframe/internal/modules/system/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/system/repository"
	"github.com/CIPFZ/gowebframe/internal/svc"
)

// ICasbinService 定义了 Casbin 权限管理的服务层接口
type ICasbinService interface {
	// UpdateCasbin 更新指定角色的 API 权限（先清空后添加）
	UpdateCasbin(ctx context.Context, authorityId string, casbinInfos []dto.CasbinInfo) error
	// GetPolicyPathByAuthorityId 获取指定角色的所有 API 权限
	GetPolicyPathByAuthorityId(ctx context.Context, authorityId string) ([]dto.CasbinInfo, error)
}

// CasbinService 是 ICasbinService 的实现
type CasbinService struct {
	svcCtx     *svc.ServiceContext
	casbinRepo repository.ICasbinRepository // 依赖注入 CasbinRepository
}

// NewCasbinService 创建一个新的 CasbinService 实例
func NewCasbinService(svcCtx *svc.ServiceContext, casbinRepo repository.ICasbinRepository) ICasbinService {
	return &CasbinService{
		svcCtx:     svcCtx,
		casbinRepo: casbinRepo,
	}
}

// UpdateCasbin 更新角色的 API 权限。此操作是覆盖性的，会先删除该角色的所有旧 API 策略，然后添加新的策略。
func (s *CasbinService) UpdateCasbin(ctx context.Context, authorityId string, casbinInfos []dto.CasbinInfo) error {
	// 1. 调用仓库层，清除该角色所有旧的 API 权限策略
	if err := s.casbinRepo.ClearPolicy(ctx, authorityId); err != nil {
		return err
	}

	// 2. 组装新的规则列表
	// Casbin 规则的格式为: [subject, object, action] -> [角色ID, API路径, 请求方法]
	var rules [][]string
	for _, info := range casbinInfos {
		rules = append(rules, []string{authorityId, info.Path, info.Method})
	}

	// 3. 如果有新规则，则调用仓库层批量添加
	if len(rules) > 0 {
		return s.casbinRepo.AddPolicies(ctx, rules)
	}

	return nil
}

// GetPolicyPathByAuthorityId 获取指定角色当前拥有的所有 API 权限
func (s *CasbinService) GetPolicyPathByAuthorityId(ctx context.Context, authorityId string) ([]dto.CasbinInfo, error) {
	// 1. 从仓库获取原始的 Casbin 策略数据
	list, err := s.casbinRepo.GetPolicy(ctx, authorityId)
	if err != nil {
		return nil, err
	}

	// 2. 将原始策略数据（[][]string）转换为对前端友好的 DTO 格式（[]dto.CasbinInfo）
	var infos []dto.CasbinInfo
	for _, v := range list {
		// 原始数据 v 的格式是 [v0, v1, v2]，分别对应 subject, object, action
		// 即 v[0]=角色ID, v[1]=API路径, v[2]=请求方法
		if len(v) >= 3 {
			infos = append(infos, dto.CasbinInfo{
				Path:   v[1],
				Method: v[2],
			})
		}
	}
	return infos, nil
}
