package service

import (
	"errors"
	"time"

	"go-im-server/common"
	"go-im-server/db"
	"go-im-server/model"

	"gorm.io/gorm"
)

type CreateGroupReq struct {
	Name string `json:"name" binding:"required,max=64"`
}

type GroupInfoResp struct {
	model.Group
	Members []model.GroupMember `json:"members"`
}

func CreateGroup(ownerID int64, req *CreateGroupReq) (*model.Group, error) {
	now := time.Now()
	group := model.Group{
		ID:          common.GenID(),
		Name:        req.Name,
		OwnerID:     ownerID,
		MemberCount: 1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&group).Error; err != nil {
			return err
		}
		member := model.GroupMember{
			ID:       common.GenID(),
			GroupID:  group.ID,
			UserID:   ownerID,
			Role:     1,
			JoinedAt: now,
		}
		return tx.Create(&member).Error
	})

	return &group, err
}

func JoinGroup(userID, groupID int64) error {
	var group model.Group
	if err := db.DB.First(&group, groupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(common.GetErrMsg(common.ErrGroupNotFound))
		}
		return err
	}

	var exist model.GroupMember
	if err := db.DB.Where("group_id = ? AND user_id = ?", groupID, userID).First(&exist).Error; err == nil {
		return errors.New(common.GetErrMsg(common.ErrAlreadyInGroup))
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		member := model.GroupMember{
			ID:       common.GenID(),
			GroupID:  groupID,
			UserID:   userID,
			Role:     0,
			JoinedAt: time.Now(),
		}
		if err := tx.Create(&member).Error; err != nil {
			return err
		}
		return tx.Model(&group).UpdateColumn("member_count", gorm.Expr("member_count + 1")).Error
	})
}

func LeaveGroup(userID, groupID int64) error {
	var member model.GroupMember
	if err := db.DB.Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error; err != nil {
		return errors.New(common.GetErrMsg(common.ErrNotInGroup))
	}

	if member.Role == 1 {
		return errors.New("群主不能退出群")
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&member).Error; err != nil {
			return err
		}
		return tx.Model(&model.Group{}).Where("id = ?", groupID).
			UpdateColumn("member_count", gorm.Expr("member_count - 1")).Error
	})
}

func GroupInfo(groupID int64) (*model.Group, error) {
	var group model.Group
	if err := db.DB.Preload("Owner").First(&group, groupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(common.GetErrMsg(common.ErrGroupNotFound))
		}
		return nil, err
	}
	return &group, nil
}

func GroupMembers(groupID int64) ([]model.GroupMember, error) {
	var members []model.GroupMember
	if err := db.DB.Where("group_id = ?", groupID).Preload("User").Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

func MyGroups(userID int64) ([]model.GroupMember, error) {
	var members []model.GroupMember
	if err := db.DB.Where("user_id = ?", userID).Preload("Group").Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

func IsGroupMember(userID, groupID int64) bool {
	var count int64
	db.DB.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, userID).Count(&count)
	return count > 0
}

func InviteMember(inviterID, groupID, inviteeID int64) error {
	if !IsGroupMember(inviterID, groupID) {
		return errors.New("您不是群成员")
	}

	var group model.Group
	if err := db.DB.First(&group, groupID).Error; err != nil {
		return errors.New(common.GetErrMsg(common.ErrGroupNotFound))
	}

	var exist model.GroupMember
	if err := db.DB.Where("group_id = ? AND user_id = ?", groupID, inviteeID).First(&exist).Error; err == nil {
		return errors.New(common.GetErrMsg(common.ErrAlreadyInGroup))
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		member := model.GroupMember{
			ID:       common.GenID(),
			GroupID:  groupID,
			UserID:   inviteeID,
			Role:     0,
			JoinedAt: time.Now(),
		}
		if err := tx.Create(&member).Error; err != nil {
			return err
		}
		return tx.Model(&group).UpdateColumn("member_count", gorm.Expr("member_count + 1")).Error
	})
}

func GetGroupMemberIDs(groupID int64, excludeUID int64) ([]int64, error) {
	var members []model.GroupMember
	if err := db.DB.Where("group_id = ? AND user_id != ?", groupID, excludeUID).
		Select("user_id").Find(&members).Error; err != nil {
		return nil, err
	}
	ids := make([]int64, len(members))
	for i, m := range members {
		ids[i] = m.UserID
	}
	return ids, nil
}
