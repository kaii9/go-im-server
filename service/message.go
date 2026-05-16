package service

import (
	"time"

	"go-im-server/common"
	"go-im-server/db"
	"go-im-server/model"
)

type MessageHistoryReq struct {
	UserID     int64 `json:"-"`
	TargetType int8  `form:"target_type" binding:"required,oneof=1 2"`
	TargetID   int64 `form:"target_id" binding:"required"`
	Page       int   `form:"page" binding:"min=1"`
	PageSize   int   `form:"page_size" binding:"min=1,max=50"`
}

type MessageHistoryResp struct {
	List     []model.Message `json:"list"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

type ConversationInfo struct {
	ID           int64     `json:"id"`
	TargetType   int8      `json:"target_type"`
	TargetID     int64     `json:"target_id"`
	LastMessage  string    `json:"last_message"`
	UnreadCount  int       `json:"unread_count"`
	UpdatedAt    time.Time `json:"updated_at"`
	TargetName   string    `json:"target_name"`
	TargetAvatar string    `json:"target_avatar"`
}

func SaveMessage(msg *model.Message) error {
	msg.ID = common.GenID()
	if err := db.DB.Create(msg).Error; err != nil {
		return err
	}

	upsertConversation(msg.SenderID, msg.TargetType, msg.TargetID, msg.Content, msg.ContentType, 0, msg.CreatedAt)

	if msg.TargetType == 1 {
		upsertConversation(msg.TargetID, msg.TargetType, msg.SenderID, msg.Content, msg.ContentType, 1, msg.CreatedAt)
	}

	return nil
}

func upsertConversation(userID int64, targetType int8, targetID int64, lastMsg string, contentType int8, unreadInc int, now time.Time) {
	var existing model.Conversation
	err := db.DB.Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		First(&existing).Error

	if err != nil {
		db.DB.Create(&model.Conversation{
			ID:          common.GenID(),
			UserID:      userID,
			TargetType:  targetType,
			TargetID:    targetID,
			LastMessage: truncateContent(lastMsg, contentType),
			UnreadCount: unreadInc,
			UpdatedAt:   now,
		})
	} else {
		updates := map[string]interface{}{
			"last_message": truncateContent(lastMsg, contentType),
			"updated_at":   now,
		}
		if unreadInc > 0 {
			updates["unread_count"] = existing.UnreadCount + unreadInc
		}
		db.DB.Model(&existing).Updates(updates)
	}
}

func truncateContent(s string, contentType int8) string {
	if contentType == 2 {
		return "[图片]"
	}
	runes := []rune(s)
	if len(runes) > 50 {
		return string(runes[:50]) + "..."
	}
	return s
}

func MessageHistory(req *MessageHistoryReq) (*MessageHistoryResp, error) {
	var total int64
	var msgs []model.Message

	query := db.DB.Model(&model.Message{})
	if req.TargetType == 1 {
		// 单聊：查询双向消息
		query = query.Where("target_type = 1 AND ((sender_id = ? AND target_id = ?) OR (sender_id = ? AND target_id = ?))",
			req.UserID, req.TargetID, req.TargetID, req.UserID)
	} else {
		query = query.Where("target_type = ? AND target_id = ?", req.TargetType, req.TargetID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if err := query.Preload("Sender").Order("created_at DESC").
		Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&msgs).Error; err != nil {
		return nil, err
	}

	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	return &MessageHistoryResp{
		List: msgs, Total: total, Page: req.Page, PageSize: req.PageSize,
	}, nil
}

func Conversations(userID int64) ([]ConversationInfo, error) {
	var convs []model.Conversation
	if err := db.DB.Where("user_id = ?", userID).Order("updated_at DESC").Find(&convs).Error; err != nil {
		return nil, err
	}

	result := make([]ConversationInfo, 0, len(convs))
	for _, conv := range convs {
		info := ConversationInfo{
			ID: conv.ID, TargetType: conv.TargetType, TargetID: conv.TargetID,
			LastMessage: conv.LastMessage, UnreadCount: conv.UnreadCount, UpdatedAt: conv.UpdatedAt,
		}
		if conv.TargetType == 1 {
			if user, err := GetUserByID(conv.TargetID); err == nil {
				info.TargetName = user.Nickname
				info.TargetAvatar = user.Avatar
			}
		} else {
			if group, err := GroupInfo(conv.TargetID); err == nil {
				info.TargetName = group.Name
				info.TargetAvatar = group.Avatar
			}
		}
		result = append(result, info)
	}
	return result, nil
}

func ClearUnread(userID int64, targetType int8, targetID int64) error {
	return db.DB.Model(&model.Conversation{}).
		Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		Update("unread_count", 0).Error
}

type SearchMessageReq struct {
	Keyword  string `form:"keyword" binding:"required"`
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=50"`
}

func SearchMessage(userID int64, req *SearchMessageReq) (*MessageHistoryResp, error) {
	var total int64
	var msgs []model.Message

	keyword := "%" + req.Keyword + "%"
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	// 搜索用户参与的单聊和群聊消息
	query := db.DB.Model(&model.Message{}).
		Where("content LIKE ?", keyword).
		Where("(target_type = 1 AND (sender_id = ? OR target_id = ?)) OR (target_type = 2 AND target_id IN (SELECT group_id FROM group_members WHERE user_id = ?))",
			userID, userID, userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if err := query.Preload("Sender").Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&msgs).Error; err != nil {
		return nil, err
	}

	return &MessageHistoryResp{
		List: msgs, Total: total, Page: page, PageSize: pageSize,
	}, nil
}

func UpsertGroupConversation(memberIDs []int64, groupID int64, content string, contentType int8, now time.Time) {
	for _, mid := range memberIDs {
		upsertConversation(mid, 2, groupID, content, contentType, 1, now)
	}
}
