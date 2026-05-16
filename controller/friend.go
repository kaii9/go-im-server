package controller

import (
	"strconv"

	"go-im-server/common"
	"go-im-server/service"

	"github.com/gin-gonic/gin"
)

func ApplyFriend(c *gin.Context) {
	var req service.ApplyFriendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.ErrInvalidParam, common.GetErrMsg(common.ErrInvalidParam))
		return
	}

	if err := service.ApplyFriend(getUID(c), &req); err != nil {
		code := common.ErrInternal
		switch err.Error() {
		case common.GetErrMsg(common.ErrCantAddSelf):
			code = common.ErrCantAddSelf
		case common.GetErrMsg(common.ErrAlreadyFriend):
			code = common.ErrAlreadyFriend
		case common.GetErrMsg(common.ErrAlreadyApplied):
			code = common.ErrAlreadyApplied
		}
		common.Error(c, code, err.Error())
		return
	}
	common.Success(c, nil)
}

func HandleFriend(c *gin.Context) {
	var req service.HandleFriendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.ErrInvalidParam, common.GetErrMsg(common.ErrInvalidParam))
		return
	}

	if err := service.HandleFriend(getUID(c), &req); err != nil {
		common.Error(c, common.ErrInternal, err.Error())
		return
	}
	common.Success(c, nil)
}

func FriendApplications(c *gin.Context) {
	typ := c.DefaultQuery("type", "received")
	apps, err := service.FriendApplications(getUID(c), typ)
	if err != nil {
		common.Error(c, common.ErrInternal, err.Error())
		return
	}
	common.Success(c, apps)
}

func FriendList(c *gin.Context) {
	friends, err := service.FriendList(getUID(c))
	if err != nil {
		common.Error(c, common.ErrInternal, err.Error())
		return
	}
	common.Success(c, friends)
}

func DeleteFriend(c *gin.Context) {
	friendID, err := strconv.ParseInt(c.Query("friend_id"), 10, 64)
	if err != nil {
		common.Error(c, common.ErrInvalidParam, common.GetErrMsg(common.ErrInvalidParam))
		return
	}

	if err := service.DeleteFriend(getUID(c), friendID); err != nil {
		common.Error(c, common.ErrInternal, err.Error())
		return
	}
	common.Success(c, nil)
}
