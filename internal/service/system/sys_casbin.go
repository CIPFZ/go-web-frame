package system

import (
	"context"
	"errors"

	"github.com/CIPFZ/gowebframe/internal/model/system/request"
	"github.com/CIPFZ/gowebframe/internal/svc"
)

type ICasbinService interface {
	UpdateCasbin(ctx context.Context, authorityId string, casbinInfos []request.CasbinInfo) error
	GetPolicyPathByAuthorityId(ctx context.Context, authorityId string) ([]request.CasbinInfo, error)
}

type CasbinService struct {
	svcCtx *svc.ServiceContext
}

func NewCasbinService(svcCtx *svc.ServiceContext) ICasbinService {
	return &CasbinService{svcCtx: svcCtx}
}

// UpdateCasbin 更新 Casbin 权限
func (s *CasbinService) UpdateCasbin(ctx context.Context, authorityId string, casbinInfos []request.CasbinInfo) error {
	// 1. 清除该角色所有旧的 API 权限
	// 0 表示清除字段索引为 0 的列 (v0)，即 role_id
	_, err := s.svcCtx.CasbinEnforcer.RemoveFilteredPolicy(0, authorityId)
	if err != nil {
		return err
	}

	// 2. 组装新规则
	// 规则格式: [role_id, path, method]
	var rules [][]string
	for _, info := range casbinInfos {
		rules = append(rules, []string{authorityId, info.Path, info.Method})
	}

	// 3. 批量添加新规则
	if len(rules) > 0 {
		success, err := s.svcCtx.CasbinEnforcer.AddPolicies(rules)
		if !success {
			return errors.New("存在重复规则，部分添加失败")
		}
		if err != nil {
			return err
		}
	}

	// 4. (可选) 再次手动 LoadPolicy 确保内存同步，虽然 AddPolicies 通常会自动同步
	//_ = s.svcCtx.CasbinEnforcer.LoadPolicy()

	return nil
}

// GetPolicyPathByAuthorityId 获取角色当前拥有的权限
func (s *CasbinService) GetPolicyPathByAuthorityId(ctx context.Context, authorityId string) ([]request.CasbinInfo, error) {
	// 从 Casbin 中获取该角色的所有策略
	// GetFilteredPolicy(0, authorityId) -> 获取 v0 == authorityId 的所有记录
	list, err := s.svcCtx.CasbinEnforcer.GetFilteredPolicy(0, authorityId)
	if err != nil {
		return nil, err
	}

	var infos []request.CasbinInfo
	for _, v := range list {
		// v[0]=role, v[1]=path, v[2]=method
		if len(v) >= 3 {
			infos = append(infos, request.CasbinInfo{
				Path:   v[1],
				Method: v[2],
			})
		}
	}
	return infos, nil
}
