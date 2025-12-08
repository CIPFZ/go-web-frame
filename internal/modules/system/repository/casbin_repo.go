package repository

import (
	"context"
	"errors"
	"github.com/casbin/casbin/v2"
)

// ICasbinRepository 定义了与 Casbin 存储交互的接口
// 它的主要职责是封装 Casbin Enforcer 的原生方法，提供更清晰的业务语义。
type ICasbinRepository interface {
	// ClearPolicy 清除指定角色的所有策略
	// 注意：虽然 Casbin API 不强制要求 Context，但为了接口统一和未来扩展，我们保留它
	ClearPolicy(ctx context.Context, authorityId string) error

	// AddPolicies 批量添加策略
	AddPolicies(ctx context.Context, rules [][]string) error

	// GetPolicy 获取指定角色的所有策略
	GetPolicy(ctx context.Context, authorityId string) ([][]string, error)

	// SyncPolicy 从持久化存储（如数据库）重新加载所有策略到内存
	SyncPolicy(ctx context.Context) error
}

// CasbinRepository 是 ICasbinRepository 的实现，它包装了一个 Casbin Enforcer 实例
type CasbinRepository struct {
	// 使用 SyncedCachedEnforcer 可以确保在分布式环境下，当一个实例修改了策略后，
	// 其他实例可以通过 Watcher 机制（如 Redis Watcher）自动同步更新，而无需手动调用 LoadPolicy。
	// CachedEnforcer 提供了缓存，避免每次检查权限都查询数据库。
	enforcer *casbin.SyncedCachedEnforcer
}

// NewCasbinRepository 创建一个新的 CasbinRepository 实例
func NewCasbinRepository(enforcer *casbin.SyncedCachedEnforcer) ICasbinRepository {
	return &CasbinRepository{enforcer: enforcer}
}

// ClearPolicy 通过过滤策略的第0个字段（v0，即 subject/角色ID）来移除一个角色的所有权限
func (r *CasbinRepository) ClearPolicy(ctx context.Context, authorityId string) error {
	// RemoveFilteredPolicy 会删除匹配过滤器的策略，并从持久化存储中移除它们。
	// 第一个参数 0 表示按策略的第 0 个字段（v0）进行过滤。
	// 后续参数是过滤条件，这里是 authorityId。
	_, err := r.enforcer.RemoveFilteredPolicy(0, authorityId)
	return err
}

// AddPolicies 批量向 Casbin 添加新的策略规则
func (r *CasbinRepository) AddPolicies(ctx context.Context, rules [][]string) error {
	if len(rules) == 0 {
		return nil
	}
	// AddPolicies 会将规则添加到当前策略，并持久化到存储中。
	// 如果规则已存在，则不会重复添加。
	success, err := r.enforcer.AddPolicies(rules)
	if err != nil {
		return err
	}
	// 如果 success 为 false，通常意味着部分或全部规则因为已经存在而没有被添加。
	// 在我们的业务场景中，这不应视为一个硬性错误，但可以根据需要返回一个特定的错误类型。
	if !success {
		// 这个错误信息可能需要根据具体业务调整，因为 "重复" 在某些场景下是预期的。
		return errors.New("存在重复规则，部分添加失败")
	}
	return nil
}

// GetPolicy 获取指定角色的所有策略规则
func (r *CasbinRepository) GetPolicy(ctx context.Context, authorityId string) ([][]string, error) {
	// GetFilteredPolicy 从内存中的策略缓存获取匹配过滤器的策略。
	// 第一个参数 0 表示按策略的第 0 个字段（v0）进行过滤。
	return r.enforcer.GetFilteredPolicy(0, authorityId)
}

// SyncPolicy 手动触发一次从持久化存储到内存的策略全量加载
// 在未使用 Watcher 或需要强制同步时非常有用。
func (r *CasbinRepository) SyncPolicy(ctx context.Context) error {
	return r.enforcer.LoadPolicy()
}
