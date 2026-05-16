package service

import (
	"errors"
	"time"

	"go-im-server/common"
	"go-im-server/db"
	"go-im-server/model"

	"gorm.io/gorm"
)

type ApplyFriendReq struct {
	ToUserID int64  `json:"to_user_id" binding:"required"`
	Reason   string `json:"reason"`
}

type HandleFriendReq struct {
	ApplicationID int64 `json:"application_id" binding:"required"`
	Agree         bool  `json:"agree"`
}

func ApplyFriend(fromUID int64, req *ApplyFriendReq) error {
	if fromUID == req.ToUserID {
		return errors.New(common.GetErrMsg(common.ErrCantAddSelf))
	}

	if _, err := GetUserByID(req.ToUserID); err != nil {
		return err
	}

	var friend model.Friend
	if err := db.DB.Where("user_id = ? AND friend_id = ?", fromUID, req.ToUserID).First(&friend).Error; err == nil {
		return errors.New(common.GetErrMsg(common.ErrAlreadyFriend))
	}

	var exist model.FriendApplication
	if err := db.DB.Where("from_user_id = ? AND to_user_id = ? AND status = 0", fromUID, req.ToUserID).
		First(&exist).Error; err == nil {
		return errors.New(common.GetErrMsg(common.ErrAlreadyApplied))
	}

	app := model.FriendApplication{
		ID:         common.GenID(),
		FromUserID: fromUID,
		ToUserID:   req.ToUserID,
		Status:     0,
		Reason:     req.Reason,
		CreatedAt:  time.Now(),
	}
	return db.DB.Create(&app).Error
}

func HandleFriend(handlerUID int64, req *HandleFriendReq) error {
	var app model.FriendApplication
	if err := db.DB.Where("id = ? AND to_user_id = ? AND status = 0", req.ApplicationID, handlerUID).
		First(&app).Error; err != nil {
		return errors.New("申请不存在或已处理")
	}

	now := time.Now()
	if req.Agree {
		app.Status = 1
		app.HandledAt = &now

		return db.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Save(&app).Error; err != nil {
				return err
			}
			f1 := model.Friend{ID: common.GenID(), UserID: app.FromUserID, FriendID: app.ToUserID, CreatedAt: now}
			f2 := model.Friend{ID: common.GenID(), UserID: app.ToUserID, FriendID: app.FromUserID, CreatedAt: now}
			if err := tx.Create(&f1).Error; err != nil {
				return err
			}
			return tx.Create(&f2).Error
		})
	}

	app.Status = 2
	app.HandledAt = &now
	return db.DB.Save(&app).Error
}

type ApplicationInfo struct {
	ID         int64       `json:"id"`
	FromUserID int64       `json:"from_user_id"`
	ToUserID   int64       `json:"to_user_id"`
	Status     int8        `json:"status"`
	Reason     string      `json:"reason"`
	CreatedAt  time.Time   `json:"created_at"`
	FromUser   *model.User `json:"from_user,omitempty"`
	ToUser     *model.User `json:"to_user,omitempty"`
}

func FriendApplications(uid int64, typ string) ([]ApplicationInfo, error) {
	var apps []model.FriendApplication
	query := db.DB.Model(&model.FriendApplication{})
	if typ == "sent" {
		query = query.Where("from_user_id = ?", uid).Preload("ToUser")
	} else {
		query = query.Where("to_user_id = ?", uid).Preload("FromUser")
	}

	if err := query.Order("created_at DESC").Find(&apps).Error; err != nil {
		return nil, err
	}

	result := make([]ApplicationInfo, len(apps))
	for i, a := range apps {
		result[i] = ApplicationInfo{
			ID: a.ID, FromUserID: a.FromUserID, ToUserID: a.ToUserID,
			Status: a.Status, Reason: a.Reason, CreatedAt: a.CreatedAt,
			FromUser: a.FromUser, ToUser: a.ToUser,
		}
	}
	return result, nil
}

func FriendList(uid int64) ([]model.Friend, error) {
	var friends []model.Friend
	if err := db.DB.Where("user_id = ?", uid).Preload("Friend").Find(&friends).Error; err != nil {
		return nil, err
	}
	return friends, nil
}

func DeleteFriend(uid, friendID int64) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND friend_id = ?", uid, friendID).Delete(&model.Friend{}).Error; err != nil {
			return err
		}
		return tx.Where("user_id = ? AND friend_id = ?", friendID, uid).Delete(&model.Friend{}).Error
	})
}

func IsFriend(uidA, uidB int64) bool {
	var count int64
	db.DB.Model(&model.Friend{}).Where("user_id = ? AND friend_id = ?", uidA, uidB).Count(&count)
	return count > 0
}
