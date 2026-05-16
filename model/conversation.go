package model

import "time"

type Conversation struct {
	ID          int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	UserID      int64     `gorm:"uniqueIndex:idx_user_target,priority:1;not null" json:"user_id"`
	TargetType  int8      `gorm:"uniqueIndex:idx_user_target,priority:2;not null" json:"target_type"`
	TargetID    int64     `gorm:"uniqueIndex:idx_user_target,priority:3;not null" json:"target_id"`
	LastMessage string    `gorm:"size:255" json:"last_message"`
	UnreadCount int       `gorm:"default:0" json:"unread_count"`
	UpdatedAt   time.Time `json:"updated_at"`
}
