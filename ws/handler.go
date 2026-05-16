package ws

import (
	"encoding/json"
	"time"

	"go-im-server/model"
	"go-im-server/service"

	"go.uber.org/zap"
)

var handlerLog, _ = zap.NewProduction()

func HandleMessage(client *Client, raw []byte) {
	var msg Message
	if err := json.Unmarshal(raw, &msg); err != nil {
		handlerLog.Error("解析WS消息失败", zap.Error(err))
		return
	}

	msg.From = client.UserID
	msg.Timestamp = time.Now().Unix()

	now := time.Now()

	msgModel := &model.Message{
		SenderID:    client.UserID,
		TargetType:  msg.TargetType,
		TargetID:    msg.To,
		ContentType: msg.ContentType,
		Content:     msg.Content,
		CreatedAt:   now,
	}

	switch msg.Type {
	case TypeSingleMsg:
		if !service.IsFriend(client.UserID, msg.To) {
			SendJSONToUser(client.UserID, map[string]interface{}{
				"type": TypeSysNotify, "content": "对方不是您的好友",
			})
			return
		}

		if err := service.SaveMessage(msgModel); err != nil {
			handlerLog.Error("保存消息失败", zap.Error(err))
			return
		}

		if DefaultHub.IsOnline(msg.To) {
			DefaultHub.SendToUser(msg.To, msg)
		}

		DefaultHub.SendToUser(client.UserID, msg)

	case TypeGroupMsg:
		if !service.IsGroupMember(client.UserID, msg.To) {
			return
		}

		if err := service.SaveMessage(msgModel); err != nil {
			handlerLog.Error("保存群消息失败", zap.Error(err))
			return
		}

		memberIDs, err := service.GetGroupMemberIDs(msg.To, client.UserID)
		if err != nil {
			return
		}
		service.UpsertGroupConversation(memberIDs, msg.To, msg.Content, msg.ContentType, now)

		for _, mid := range memberIDs {
			if DefaultHub.IsOnline(mid) {
				DefaultHub.SendToUser(mid, msg)
			}
		}

		DefaultHub.SendToUser(client.UserID, msg)

	default:
	}
}
