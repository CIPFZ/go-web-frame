package dto

type RegisterReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	NickName string `json:"nickName"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

type LoginReq struct {
	Type      string `json:"type"`                        // 登录类型
	Username  string `json:"username" binding:"required"` // 用户名
	Password  string `json:"password" binding:"required"` // 密码
	Captcha   string `json:"captcha"`                     // 验证码 (预留)
	CaptchaId string `json:"captchaId"`                   // 验证码ID (预留)
}

// SearchUserReq 用户列表查询
type SearchUserReq struct {
	PageInfo
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

// UpdateSelfInfoReq 更新个人信息请求
type UpdateSelfInfoReq struct {
	NickName string `json:"nickName" binding:"required"` // 昵称
	Bio      string `json:"bio"`                         // 个人简介
}

// UpdateUiConfigReq 更新 UI 配置请求
type UpdateUiConfigReq struct {
	// 前端传来的是一个 JSON 对象，我们用 map[string]any 接收
	// 或者直接用 json.RawMessage 接收
	Settings map[string]any `json:"settings" binding:"required"`
}

// ----------------
// --- Response ---
// ----------------

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expiresAt"`
}
