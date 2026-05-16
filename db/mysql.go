package db

import (
	"go-im-server/config"
	"go-im-server/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() error {
	var err error
	DB, err = gorm.Open(mysql.Open(config.AppConfig.Database.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	return DB.AutoMigrate(
		&model.User{},
		&model.FriendApplication{},
		&model.Friend{},
		&model.Group{},
		&model.GroupMember{},
		&model.Message{},
		&model.Conversation{},
	)
}
