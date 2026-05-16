package model

import "time"

type User struct {
	ID        int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:32;not null" json:"username"`
	Password  string    `gorm:"size:128;not null" json:"-"`
	Nickname  string    `gorm:"size:64" json:"nickname"`
	Avatar    string    `gorm:"size:255" json:"avatar"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
