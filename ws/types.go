package ws

const (
	TypeSingleMsg = 1 // 单聊消息
	TypeGroupMsg  = 2 // 群聊消息
	TypeSysNotify = 3 // 系统通知（好友申请结果、入群通知等）
)

const (
	TargetTypeSingle = 1 // 单聊
	TargetTypeGroup  = 2 // 群聊
)

const (
	ContentTypeText  = 1 // 文本
	ContentTypeImage = 2 // 图片
)

type Message struct {
	Type        int8   `json:"type"`
	From        int64  `json:"from"`
	To          int64  `json:"to"`
	TargetType  int8   `json:"target_type"`
	ContentType int8   `json:"content_type"`
	Content     string `json:"content"`
	Timestamp   int64  `json:"timestamp"`
}
