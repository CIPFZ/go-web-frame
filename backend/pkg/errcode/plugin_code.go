package errcode

var (
	PluginForbidden             = NewError(3001, "plugin access forbidden")
	PluginStatusInvalid         = NewError(3002, "invalid plugin release status transition")
	PluginWorkOrderAlreadyClaim = NewError(3003, "work order has already been claimed")
	PluginWorkOrderNotClaimed   = NewError(3004, "work order has not been claimed")
	PluginResetReasonRequired   = NewError(3005, "reset reason is required")
	PluginReviewCommentRequired = NewError(3006, "review comment is required")
	PluginProductInvalid        = NewError(3007, "invalid compatible target")
	PluginVersionInvalid        = NewError(3008, "invalid plugin version")
	PluginVersionDuplicate      = NewError(3009, "duplicate plugin version")
	PluginCompatibilityRequired = NewError(3010, "compatibility is required")
	PluginReleaseNotEditable    = NewError(3011, "release content is not editable in current status")
)
