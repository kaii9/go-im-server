package controller

import (
	"strconv"

	"go-im-server/common"
	"go-im-server/service"

	"github.com/gin-gonic/gin"
)

func CreateGroup(c *gin.Context) {
	var req service.CreateGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.ErrInvalidParam, common.GetErrMsg(common.ErrInvalidParam))
		return
	}

	group, err := service.CreateGroup(getUID(c), &req)
	if err != nil {
		common.Error(c, common.ErrInternal, err.Error())
		return
	}
	common.Success(c, group)
}

func JoinGroup(c *gin.Context) {
	var req struct {
		GroupID int64 `json:"group_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.ErrInvalidParam, common.GetErrMsg(common.ErrInvalidParam))
		return
	}

	if err := service.JoinGroup(getUID(c), req.GroupID); err != nil {
		code := common.ErrInternal
		switch err.Error() {
		case common.GetErrMsg(common.ErrGroupNotFound):
			code = common.ErrGroupNotFound
		case common.GetErrMsg(common.ErrAlreadyInGroup):
			code = common.ErrAlreadyInGroup
		}
		common.Error(c, code, err.Error())
		return
	}
	common.Success(c, nil)
}

func LeaveGroup(c *gin.Context) {
	var req struct {
		GroupID int64 `json:"group_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.ErrInvalidParam, common.GetErrMsg(common.ErrInvalidParam))
		return
	}

	if err := service.LeaveGroup(getUID(c), req.GroupID); err != nil {
		common.Error(c, common.ErrInternal, err.Error())
		return
	}
	common.Success(c, nil)
}

func GroupInfo(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil {
		common.Error(c, common.ErrInvalidParam, common.GetErrMsg(common.ErrInvalidParam))
		return
	}

	group, err := service.GroupInfo(groupID)
	if err != nil {
		code := common.ErrInternal
		if err.Error() == common.GetErrMsg(common.ErrGroupNotFound) {
			code = common.ErrGroupNotFound
		}
		common.Error(c, code, err.Error())
		return
	}
	common.Success(c, group)
}

func GroupMembers(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil {
		common.Error(c, common.ErrInvalidParam, common.GetErrMsg(common.ErrInvalidParam))
		return
	}

	members, err := service.GroupMembers(groupID)
	if err != nil {
		common.Error(c, common.ErrInternal, err.Error())
		return
	}
	common.Success(c, members)
}

func MyGroups(c *gin.Context) {
	groups, err := service.MyGroups(getUID(c))
	if err != nil {
		common.Error(c, common.ErrInternal, err.Error())
		return
	}
	common.Success(c, groups)
}

func InviteMember(c *gin.Context) {
	var req struct {
		GroupID int64 `json:"group_id" binding:"required"`
		UserID  int64 `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.ErrInvalidParam, common.GetErrMsg(common.ErrInvalidParam))
		return
	}

	if err := service.InviteMember(getUID(c), req.GroupID, req.UserID); err != nil {
		code := common.ErrInternal
		switch err.Error() {
		case common.GetErrMsg(common.ErrGroupNotFound):
			code = common.ErrGroupNotFound
		case common.GetErrMsg(common.ErrAlreadyInGroup):
			code = common.ErrAlreadyInGroup
		}
		common.Error(c, code, err.Error())
		return
	}
	common.Success(c, nil)
}
