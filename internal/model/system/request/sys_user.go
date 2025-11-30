package request

import "github.com/CIPFZ/gowebframe/internal/model/common/request"

type RegisterReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	NickName string `json:"nickName"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

type Login struct {
	Type      string `json:"type"`                        // 登录类型
	Username  string `json:"username" binding:"required"` // 用户名
	Password  string `json:"password" binding:"required"` // 密码
	Captcha   string `json:"captcha"`                     // 验证码 (预留)
	CaptchaId string `json:"captchaId"`                   // 验证码ID (预留)
}

// SearchUserReq 用户列表查询
type SearchUserReq struct {
	request.PageInfo
	Username string `json:"username"`
	NickName string `json:"nickName"`
	Phone    string `json:"phone"`
}

// AddUserReq 新增用户
type AddUserReq struct {
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
	NickName     string `json:"nickName"`
	AuthorityIds []uint `json:"authorityIds"` // 选择的角色
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Status       int    `json:"status"` // 1正常 2冻结
}

// UpdateUserReq 更新用户 (不包含密码)
type UpdateUserReq struct {
	ID           uint   `json:"id" binding:"required"`
	NickName     string `json:"nickName"`
	AuthorityIds []uint `json:"authorityIds"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Status       int    `json:"status"`
}

// ResetPasswordReq 重置密码
type ResetPasswordReq struct {
	ID       uint   `json:"id" binding:"required"`
	Password string `json:"password" binding:"required"`
}
