package controller

import (
	"strconv"

	"go-im-server/common"
	"go-im-server/service"

	"github.com/gin-gonic/gin"
)

func MessageHistory(c *gin.Context) {
	targetType, err := strconv.Atoi(c.Query("target_type"))
	if err != nil {
		common.Error(c, common.ErrInvalidParam, common.GetErrMsg(common.ErrInvalidParam))
		return
	}

	targetID, err := strconv.ParseInt(c.Query("target_id"), 10, 64)
	if err != nil {
		common.Error(c, common.ErrInvalidParam, common.GetErrMsg(common.ErrInvalidParam))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	req := &service.MessageHistoryReq{
		UserID:     getUID(c),
		TargetType: int8(targetType),
		TargetID:   targetID,
		Page:       page,
		PageSize:   pageSize,
	}

	resp, err := service.MessageHistory(req)
	if err != nil {
		common.Error(c, common.ErrInternal, err.Error())
		return
	}

	service.ClearUnread(getUID(c), int8(targetType), targetID)

	common.Success(c, resp)
}

func Conversations(c *gin.Context) {
	convs, err := service.Conversations(getUID(c))
	if err != nil {
		common.Error(c, common.ErrInternal, err.Error())
		return
	}
	common.Success(c, convs)
}

func UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		common.Error(c, common.ErrInvalidParam, "请选择文件")
		return
	}
	defer file.Close()

	url, err := service.DefaultUploader.Upload(file, header)
	if err != nil {
		code := common.ErrInternal
		switch err.Error() {
		case common.GetErrMsg(common.ErrInvalidFileType):
			code = common.ErrInvalidFileType
		case common.GetErrMsg(common.ErrFileTooLarge):
			code = common.ErrFileTooLarge
		}
		common.Error(c, code, err.Error())
		return
	}

	common.Success(c, gin.H{"url": url})
}

func SearchMessage(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	req := &service.SearchMessageReq{
		Keyword:  c.Query("keyword"),
		Page:     page,
		PageSize: pageSize,
	}
	if req.Keyword == "" {
		common.Error(c, common.ErrInvalidParam, "请输入搜索关键词")
		return
	}

	resp, err := service.SearchMessage(getUID(c), req)
	if err != nil {
		common.Error(c, common.ErrInternal, err.Error())
		return
	}
	common.Success(c, resp)
}
