package errorcode

const (
	// MSAPFailed 服务器出现无法确定类型的自动捕获的错误
	MSAPFailed int32 = 1000 + iota
	// MSAPUIDEmpty 传入的UID为空
	MSAPUIDEmpty
	// MSAPGroupIDEmpty 传入的GroupID为空
	MSAPGroupIDEmpty
	// MSAPContentEmpty 传入的内容为空
	MSAPContentEmpty
	// MSAPMessageTypeUnknown 传入的消息类型未知
	MSAPMessageTypeUnknown
	// MSAPMsgIDEmpty 传入的MsgID为空
	MSAPMsgIDEmpty
	// MSAPContentTooLong 传入的MsgID为空
	MSAPContentTooLong
	// MSAPMessageNotFound 消息未找到
	MSAPMessageNotFound
	// MSAPPermissionDenied 权限不足
	MSAPPermissionDenied
)
