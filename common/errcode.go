package common

const (
	// 用户模块 1xxx
	ErrUserNotFound    = 1001
	ErrPasswordWrong   = 1002
	ErrUsernameExists  = 1003
	ErrUserNotLoggedIn = 1004

	// 好友模块 2xxx
	ErrAlreadyApplied = 2001
	ErrNotFriend      = 2002
	ErrAlreadyFriend  = 2003
	ErrCantAddSelf    = 2004

	// 消息模块 3xxx
	ErrMessageSendFailed = 3001
	ErrInvalidFileType   = 3002
	ErrFileTooLarge      = 3003

	// 群组模块 4xxx
	ErrGroupNotFound  = 4001
	ErrAlreadyInGroup = 4002
	ErrNotInGroup     = 4003

	// 通用 5xxx
	ErrInvalidParam = 5000
	ErrUnauthorized = 5001

	// 系统 9xxx
	ErrInternal = 9999
)

var errMsg = map[int]string{
	ErrUserNotFound:      "用户不存在",
	ErrPasswordWrong:     "密码错误",
	ErrUsernameExists:    "用户名已存在",
	ErrUserNotLoggedIn:   "用户未登录",
	ErrAlreadyApplied:    "已发送过好友申请",
	ErrNotFriend:         "不是好友",
	ErrAlreadyFriend:     "已经是好友",
	ErrCantAddSelf:       "不能添加自己为好友",
	ErrMessageSendFailed: "消息发送失败",
	ErrInvalidFileType:   "不支持的文件类型",
	ErrFileTooLarge:      "文件大小超出限制",
	ErrGroupNotFound:     "群组不存在",
	ErrAlreadyInGroup:    "已在群组中",
	ErrNotInGroup:        "不在群组中",
	ErrInvalidParam:      "参数错误",
	ErrUnauthorized:      "未授权",
	ErrInternal:          "服务器内部错误",
}

func GetErrMsg(code int) string {
	if msg, ok := errMsg[code]; ok {
		return msg
	}
	return "未知错误"
}
