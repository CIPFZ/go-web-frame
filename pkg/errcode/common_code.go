package errcode

var (
	Success       = NewError(0, "成功")
	ServerError   = NewError(1000, "服务内部错误")
	InvalidParams = NewError(1001, "入参错误")
	NotFound      = NewError(1002, "找不到资源")
	Unauthorized  = NewError(1003, "未授权，请登录")
	AssessDenied  = NewError(1004, "权限不足")

	// 通用 CRUD 错误 (2000 ~ 2999)
	CreateFailed       = NewError(2001, "创建失败")
	UpdateFailed       = NewError(2002, "更新失败")
	DeleteFailed       = NewError(2003, "删除失败")
	GetListFailed      = NewError(2004, "获取列表失败")
	GetDetailFailed    = NewError(2005, "获取详情失败")
	FileUploadFailed   = NewError(2006, "文件上传失败")
	NoUploadFileFailed = NewError(2007, "请选择要上传的文件")
)
