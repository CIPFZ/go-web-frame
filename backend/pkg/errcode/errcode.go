package errcode

import "fmt"

// Error 结构体，实现了 go 的 error 接口
type Error struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("code: %d, msg: %s", e.Code, e.Msg)
}

func NewError(code int, msg string) *Error {
	return &Error{
		Code: code,
		Msg:  msg,
	}
}

// WithDetails 允许在基础错误上追加详细信息（例如：数据库的具体报错）
// 返回一个新的 Error 对象，不污染全局变量
func (e *Error) WithDetails(details string) *Error {
	return &Error{
		Code: e.Code,
		Msg:  fmt.Sprintf("%s: %s", e.Msg, details),
	}
}
