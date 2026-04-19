package errcode

var (
	PluginForbidden             = NewError(3001, "插件模块无权访问")
	PluginStatusInvalid         = NewError(3002, "当前发布单状态不允许执行该操作")
	PluginWorkOrderAlreadyClaim = NewError(3003, "工单已被其他审核员认领")
	PluginWorkOrderNotClaimed   = NewError(3004, "当前工单尚未被认领")
	PluginResetReasonRequired   = NewError(3005, "重置原因不能为空")
	PluginReviewCommentRequired = NewError(3006, "审核意见不能为空")
	PluginProductInvalid        = NewError(3007, "兼容产品信息无效")
)
