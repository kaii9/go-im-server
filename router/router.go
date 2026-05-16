package router

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"go-im-server/controller"
	"go-im-server/middleware"
	"go-im-server/ws"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(middleware.CORS())
	r.Use(middleware.Recovery())

	api := r.Group("/api")
	{
		api.POST("/user/register", controller.Register)
		api.POST("/user/login", controller.Login)
	}

	r.Static("/uploads", "./uploads")
	r.GET("/ws", ws.HandleWebSocket)

	auth := api.Group("")
	auth.Use(middleware.JWTAuth())
	{
		auth.GET("/user/info", controller.GetUserInfo)
		auth.PUT("/user/update", controller.UpdateUser)
		auth.GET("/user/search", controller.SearchUser)

		auth.POST("/friend/apply", controller.ApplyFriend)
		auth.POST("/friend/handle", controller.HandleFriend)
		auth.GET("/friend/applications", controller.FriendApplications)
		auth.GET("/friend/list", controller.FriendList)
		auth.DELETE("/friend/delete", controller.DeleteFriend)

		auth.POST("/group/create", controller.CreateGroup)
		auth.POST("/group/join", controller.JoinGroup)
		auth.POST("/group/leave", controller.LeaveGroup)
		auth.POST("/group/invite", controller.InviteMember)
		auth.GET("/group/info", controller.GroupInfo)
		auth.GET("/group/members", controller.GroupMembers)
		auth.GET("/group/mine", controller.MyGroups)

		auth.GET("/message/search", controller.SearchMessage)
		auth.GET("/message/history", controller.MessageHistory)
		auth.GET("/message/conversations", controller.Conversations)
		auth.POST("/message/upload", controller.UploadImage)
	}

	// Serve frontend static files (production build)
	serveFrontend(r)

	return r
}

func serveFrontend(r *gin.Engine) {
	dist := "./web/dist"
	if _, err := os.Stat(dist); os.IsNotExist(err) {
		return
	}

	r.Static("/assets", filepath.Join(dist, "/assets"))
	r.StaticFile("/favicon.ico", filepath.Join(dist, "/favicon.ico"))

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// API 和 WS 路径不处理
		if strings.HasPrefix(path, "/api") || path == "/ws" {
			c.Status(http.StatusNotFound)
			return
		}
		// SPA fallback: 返回 index.html
		c.File(filepath.Join(dist, "/index.html"))
	})
}
