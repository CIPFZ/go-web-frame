package system

import (
	"context"
	"errors"
	"fmt"
	"time"

	systemModel "github.com/CIPFZ/gowebframe/internal/model/system"
	systemReq "github.com/CIPFZ/gowebframe/internal/model/system/request"
	systemRes "github.com/CIPFZ/gowebframe/internal/model/system/response"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/internal/utils"
	"github.com/CIPFZ/gowebframe/internal/vars"
	"github.com/CIPFZ/gowebframe/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IUserService interface {
	Register(ctx context.Context, req systemReq.RegisterReq) (*systemModel.SysUser, error)
	Login(ctx context.Context, req systemReq.Login) (*systemRes.LoginResponse, error)
	GetUserInfo(ctx context.Context, userUUID uuid.UUID) (*systemModel.SysUser, error)
	Logout(ctx context.Context, token string) error
	SwitchAuthority(ctx context.Context, uuid uuid.UUID, authorityId uint) (*systemRes.LoginResponse, error)
	GetUserList(ctx context.Context, req systemReq.SearchUserReq) (list []systemModel.SysUser, total int64, err error)
	AddUser(ctx context.Context, req systemReq.AddUserReq) error
	UpdateUser(ctx context.Context, req systemReq.UpdateUserReq) error
	DeleteUser(ctx context.Context, id uint) error
	ResetPassword(ctx context.Context, req systemReq.ResetPasswordReq) error
}

type UserService struct {
	svcCtx *svc.ServiceContext
}

func NewUserService(ctx *svc.ServiceContext) IUserService {
	return &UserService{svcCtx: ctx}
}

// Register 用户注册实现
func (s *UserService) Register(ctx context.Context, req systemReq.RegisterReq) (*systemModel.SysUser, error) {
	// 1. 检查用户名是否已存在
	var existUser systemModel.SysUser
	// ✨ 关键：使用 WithContext(ctx) 将 TraceID 传递给 GORM
	err := s.svcCtx.DB.WithContext(ctx).
		Where("username = ?", req.Username).
		First(&existUser).Error

	if err == nil {
		return nil, errors.New("用户名已存在")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		s.svcCtx.Logger.Error("查询用户失败", zap.Error(err))
		return nil, errors.New("系统内部错误")
	}

	// 2. 密码加密
	hashPwd, err := utils.BcryptHash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %v", err)
	}

	// 3. 构建用户模型
	newUser := systemModel.SysUser{
		Username:    req.Username,
		Password:    hashPwd,
		NickName:    req.NickName,
		Phone:       req.Phone,
		Email:       req.Email,
		Avatar:      "/default_avatar.jpg",
		Status:      vars.UserActive, // 默认正常
		AuthorityID: 888,             // TODO 默认普通用户角色 (建议写在配置文件中)
		UUID:        uuid.New(),
	}

	// 4. 插入数据库
	if err := s.svcCtx.DB.WithContext(ctx).Create(&newUser).Error; err != nil {
		s.svcCtx.Logger.Error("创建用户失败", zap.Error(err))
		return nil, errors.New("注册失败，请稍后重试")
	}

	return &newUser, nil
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, req systemReq.Login) (*systemRes.LoginResponse, error) {
	// 1. 查询用户 (这里不需要 Preload 太多关联表了，因为不返回用户信息，只需验证密码)
	var user systemModel.SysUser
	err := s.svcCtx.DB.WithContext(ctx).
		Where("username = ?", req.Username).
		First(&user).Error

	// 验证用户是否存在、密码比对、状态检查逻辑保持不变
	if err != nil {
		s.svcCtx.Logger.Warn("login_failed_user_not_found", zap.String("username", req.Username))
		return nil, errors.New("用户名或密码错误")
	}

	if !utils.BcryptCheck(req.Password, user.Password) {
		return nil, errors.New("用户名或密码错误")
	}

	if user.Status == vars.UserInactive {
		return nil, errors.New("此用户已经被禁用")
	}

	// 2. 签发 Token
	token, claims, err := s.generateJwtToken(user)
	if err != nil {
		s.svcCtx.Logger.Error("login_failed_token_generate", zap.Error(err))
		return nil, errors.New("获取Token失败")
	}

	// 3. 只返回 Token 和 过期时间
	return &systemRes.LoginResponse{
		Token:     token,
		ExpiresAt: claims.RegisteredClaims.ExpiresAt.Unix() * 1000,
	}, nil
}

// generateJwtToken 内部辅助函数
func (s *UserService) generateJwtToken(user systemModel.SysUser) (string, systemReq.CustomClaims, error) {
	// 构造 Claims
	claims := s.svcCtx.JWT.CreateClaims(systemReq.BaseClaims{
		UUID:        user.UUID,
		UserID:      user.ID,
		NickName:    user.NickName,
		Username:    user.Username,
		AuthorityId: user.AuthorityID,
	})

	token, err := s.svcCtx.JWT.CreateToken(claims)
	return token, claims, err
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(ctx context.Context, userUUID uuid.UUID) (*systemModel.SysUser, error) {
	var user systemModel.SysUser

	// 关联查询 Authority (当前角色) 和 Authorities (所有角色)
	err := s.svcCtx.DB.WithContext(ctx).
		Where("uuid = ?", userUUID).
		Preload("Authority").
		Preload("Authorities").
		First(&user).Error

	if err != nil {
		return nil, err
	}

	// 安全起见，清空密码
	user.Password = ""
	return &user, nil
}

// Logout 用户登出实现
func (s *UserService) Logout(ctx context.Context, token string) error {
	j := s.svcCtx.JWT

	// 1. 解析 Token 以获取过期时间
	// 我们不关心 ParseToken 是否报错（例如过期），因为如果它无效，登出目的已经达到了
	claims, err := j.ParseToken(token)
	if err != nil {
		// 如果 Token 已经无效（格式错误或已过期），直接返回成功即可
		return nil
	}

	// 2. 计算剩余有效期
	duration := claims.ExpiresAt.Sub(time.Now())
	if duration <= 0 {
		return nil // 已经过期，不需要加入黑名单
	}

	// 3. 加入 Redis 黑名单
	// 调用我们在 pkg/utils/jwt.go 中实现的 SetBlacklist
	if err := j.SetBlacklist(ctx, token, duration); err != nil {
		s.svcCtx.Logger.Error("logout_blacklist_failed", zap.Error(err))
		return err
	}

	return nil
}

// GetUserList 分页获取用户列表
func (s *UserService) GetUserList(ctx context.Context, req systemReq.SearchUserReq) (list []systemModel.SysUser, total int64, err error) {
	limit := req.PageSize
	offset := req.PageSize * (req.Page - 1)
	db := s.svcCtx.DB.WithContext(ctx).Model(&systemModel.SysUser{})

	// 动态条件
	if req.Username != "" {
		db = db.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.NickName != "" {
		db = db.Where("nick_name LIKE ?", "%"+req.NickName+"%")
	}
	if req.Phone != "" {
		db = db.Where("phone LIKE ?", "%"+req.Phone+"%")
	}

	if err = db.Count(&total).Error; err != nil {
		return
	}

	// 关联查询 Authority (主角色) 和 Authorities (多角色)
	err = db.Limit(limit).Offset(offset).
		Preload("Authority").
		Preload("Authorities").
		Order("id desc").
		Find(&list).Error

	return list, total, err
}

// AddUser 新增用户
func (s *UserService) AddUser(ctx context.Context, req systemReq.AddUserReq) error {
	// 1. 查重
	var existed systemModel.SysUser
	if !errors.Is(s.svcCtx.DB.WithContext(ctx).Where("username = ?", req.Username).First(&existed).Error, gorm.ErrRecordNotFound) {
		return errors.New("用户名已存在")
	}

	// 1. 检查角色列表
	if len(req.AuthorityIds) == 0 {
		return errors.New("请至少选择一个角色")
	}

	// 2. 加密密码
	hashPwd, err := utils.BcryptHash(req.Password)
	if err != nil {
		return err
	}

	// 3. 构建用户
	user := systemModel.SysUser{
		UUID:        uuid.New(),
		Username:    req.Username,
		Password:    hashPwd,
		NickName:    req.NickName,
		Avatar:      "/default_avatar.jpg",
		AuthorityID: req.AuthorityIds[0],
		Phone:       req.Phone,
		Email:       req.Email,
		Status:      req.Status,
	}
	if user.Status == 0 {
		user.Status = 1
	} // 默认正常

	// 4. 处理多角色
	var auths []systemModel.SysAuthority
	for _, id := range req.AuthorityIds {
		auths = append(auths, systemModel.SysAuthority{AuthorityId: id})
	}
	user.Authorities = auths

	return s.svcCtx.DB.WithContext(ctx).Create(&user).Error
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(ctx context.Context, req systemReq.UpdateUserReq) error {
	var user systemModel.SysUser
	if err := s.svcCtx.DB.WithContext(ctx).First(&user, req.ID).Error; err != nil {
		return errors.New("用户不存在")
	}

	if len(req.AuthorityIds) <= 0 {
		return errors.New("请至少选择一个角色")
	}

	return s.svcCtx.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// 1. 检查 "当前角色" 是否还在新分配的列表中
		// 如果不在，强制重置为新列表的第一个
		isOldAuthValid := false
		authorityId := user.AuthorityID
		for _, id := range req.AuthorityIds {
			if user.AuthorityID == id {
				isOldAuthValid = true
				break
			}
		}
		if !isOldAuthValid {
			// 原来的当前角色被删掉了，重置为新的第一个
			authorityId = req.AuthorityIds[0]
		}
		// 2. 更新基础信息
		updMap := map[string]interface{}{
			"nick_name":    req.NickName,
			"authority_id": authorityId,
			"phone":        req.Phone,
			"email":        req.Email,
			"status":       req.Status,
		}

		// 2. 更新主表
		if err := tx.Model(&user).Updates(updMap).Error; err != nil {
			return err
		}
		// 3. 更新角色关联
		var auths []systemModel.SysAuthority
		for _, id := range req.AuthorityIds {
			auths = append(auths, systemModel.SysAuthority{AuthorityId: id})
		}
		if err := tx.Model(&user).Association("Authorities").Replace(auths); err != nil {
			return err
		}

		return nil
	})
}

// SwitchAuthority 切换角色
func (s *UserService) SwitchAuthority(ctx context.Context, uuid uuid.UUID, authorityId uint) (*systemRes.LoginResponse, error) {
	var user systemModel.SysUser
	// 1. 查询用户及其拥有的所有角色
	err := s.svcCtx.DB.WithContext(ctx).
		Preload("Authorities").
		Where("uuid = ?", uuid).
		First(&user).Error
	if err != nil {
		return nil, err
	}

	// 2. 校验：用户是否真的拥有目标角色？
	hasAuth := false
	for _, auth := range user.Authorities {
		if auth.AuthorityId == authorityId {
			hasAuth = true
			break
		}
	}
	if !hasAuth {
		return nil, errors.New("您未拥有该角色权限")
	}

	// 3. 更新数据库中的 "当前角色"
	if err := s.svcCtx.DB.WithContext(ctx).Model(&user).Update("authority_id", authorityId).Error; err != nil {
		return nil, err
	}

	// 4. 更新内存对象以便签发 Token
	user.AuthorityID = authorityId

	// 5. 签发新 Token (因为 Token 里包含 AuthorityId，切换角色必须换 Token)
	token, claims, err := s.generateJwtToken(user)
	if err != nil {
		return nil, err
	}

	return &systemRes.LoginResponse{
		Token:     token,
		ExpiresAt: claims.RegisteredClaims.ExpiresAt.Unix() * 1000,
	}, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, id uint) error {
	log := logger.GetLogger(ctx)
	// 软删除，GORM 会自动处理
	//return s.svcCtx.DB.WithContext(ctx).Delete(&systemModel.SysUser{}, id).Error
	return s.svcCtx.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 清理用户与角色的关联 (硬删除)
		// 必须先删除这个，否则删除角色时会报错 "有用户正在使用"
		if err := tx.Table("sys_user_authorities").
			Where("user_id = ?", id).
			Delete(nil).Error; err != nil {
			log.Error("delete_user_authority_failed", zap.Error(err))
			return err
		}

		// 2. 删除用户自身 (软删除)
		if err := tx.Delete(&systemModel.SysUser{}, id).Error; err != nil {
			log.Error("delete_user_failed", zap.Error(err))
			return err
		}

		return nil
	})
}

// ResetPassword 重置密码
func (s *UserService) ResetPassword(ctx context.Context, req systemReq.ResetPasswordReq) error {
	hashPwd, err := utils.BcryptHash(req.Password)
	if err != nil {
		return err
	}
	return s.svcCtx.DB.WithContext(ctx).Model(&systemModel.SysUser{}).
		Where("id = ?", req.ID).
		Update("password", hashPwd).Error
}
