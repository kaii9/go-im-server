package controller

import (
	"strconv"

	"go-im-server/common"
	"go-im-server/service"

	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	var req service.UserRegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.ErrInvalidParam, common.GetErrMsg(common.ErrInvalidParam))
		return
	}

	if err := service.Register(&req); err != nil {
		if err.Error() == common.GetErrMsg(common.ErrUsernameExists) {
			common.Error(c, common.ErrUsernameExists, err.Error())
			return
		}
		common.Error(c, common.ErrInternal, err.Error())
		return
	}

	common.Success(c, nil)
}

func Login(c *gin.Context) {
	var req service.UserLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.ErrInvalidParam, common.GetErrMsg(common.ErrInvalidParam))
		return
	}

	resp, err := service.Login(&req)
	if err != nil {
		code := common.ErrInternal
		if err.Error() == common.GetErrMsg(common.ErrUserNotFound) {
			code = common.ErrUserNotFound
		} else if err.Error() == common.GetErrMsg(common.ErrPasswordWrong) {
			code = common.ErrPasswordWrong
		}
		common.Error(c, code, err.Error())
		return
	}

	common.Success(c, resp)
}

func GetUserInfo(c *gin.Context) {
	uid := c.GetInt64("uid")
	user, err := service.GetUserByID(uid)
	if err != nil {
		common.Error(c, common.ErrUserNotFound, err.Error())
		return
	}
	common.Success(c, user)
}

func UpdateUser(c *gin.Context) {
	uid := c.GetInt64("uid")
	var req service.UserUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.ErrInvalidParam, common.GetErrMsg(common.ErrInvalidParam))
		return
	}

	if err := service.UpdateUser(uid, &req); err != nil {
		common.Error(c, common.ErrInternal, err.Error())
		return
	}
	common.Success(c, nil)
}

func SearchUser(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		common.Error(c, common.ErrInvalidParam, common.GetErrMsg(common.ErrInvalidParam))
		return
	}

	users, err := service.SearchUsers(keyword)
	if err != nil {
		common.Error(c, common.ErrInternal, err.Error())
		return
	}
	common.Success(c, users)
}

func getUID(c *gin.Context) int64 {
	return c.GetInt64("uid")
}

func getQueryInt64(c *gin.Context, key string) (int64, error) {
	return strconv.ParseInt(c.Query(key), 10, 64)
}
