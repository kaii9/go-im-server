package model

import "time"

type Group struct {
	ID          int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	Name        string    `gorm:"size:64;not null" json:"name"`
	Avatar      string    `gorm:"size:255" json:"avatar"`
	OwnerID     int64     `gorm:"not null" json:"owner_id"`
	MemberCount int       `gorm:"default:1" json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Owner *User `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
}

type GroupMember struct {
	ID       int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	GroupID  int64     `gorm:"uniqueIndex:idx_group_user;not null" json:"group_id"`
	UserID   int64     `gorm:"uniqueIndex:idx_group_user;not null" json:"user_id"`
	Role     int8      `gorm:"default:0;not null" json:"role"`
	JoinedAt time.Time `json:"joined_at"`

	User  *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Group *Group `gorm:"foreignKey:GroupID" json:"group,omitempty"`
}
