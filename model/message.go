package model

import "time"

type Message struct {
	ID          int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	SenderID    int64     `gorm:"index;not null" json:"sender_id"`
	TargetType  int8      `gorm:"not null" json:"target_type"`
	TargetID    int64     `gorm:"not null" json:"target_id"`
	ContentType int8      `gorm:"not null" json:"content_type"`
	Content     string    `gorm:"type:text" json:"content"`
	CreatedAt   time.Time `gorm:"index:idx_target_msg,priority:3" json:"created_at"`

	Sender *User `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
}
