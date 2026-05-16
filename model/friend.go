package model

import "time"

type FriendApplication struct {
	ID         int64      `gorm:"primaryKey;autoIncrement:false" json:"id"`
	FromUserID int64      `gorm:"index;not null" json:"from_user_id"`
	ToUserID   int64      `gorm:"index;not null" json:"to_user_id"`
	Status     int8       `gorm:"default:0;not null" json:"status"`
	Reason     string     `gorm:"size:255" json:"reason"`
	HandledAt  *time.Time `json:"handled_at"`
	CreatedAt  time.Time  `json:"created_at"`

	FromUser *User `gorm:"foreignKey:FromUserID" json:"from_user,omitempty"`
	ToUser   *User `gorm:"foreignKey:ToUserID" json:"to_user,omitempty"`
}

type Friend struct {
	ID        int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	UserID    int64     `gorm:"uniqueIndex:idx_user_friend;not null" json:"user_id"`
	FriendID  int64     `gorm:"uniqueIndex:idx_user_friend;not null" json:"friend_id"`
	CreatedAt time.Time `json:"created_at"`

	Friend *User `gorm:"foreignKey:FriendID" json:"friend,omitempty"`
}
