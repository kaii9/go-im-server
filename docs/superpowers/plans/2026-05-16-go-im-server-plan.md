# Go IM Server Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a complete Go IM backend with user auth, friend management, group chat, WebSocket real-time messaging, offline message delivery, and image upload.

**Architecture:** Gin HTTP server + gorilla/websocket for real-time communication. GORM manages MySQL models with AutoMigrate. In-memory hub manages online user connection pool. Messages persist to DB and forward through goroutine channels. Layered: main → router → middleware → controller → service → model → db.

**Tech Stack:** Go 1.25, Gin, gorilla/websocket, GORM, MySQL, golang-jwt/v5, viper, bcrypt, zap logger

---

### Task 1: Initialize Go module and dependencies

**Files:**
- Modify: `go.mod`
- Create: `go.sum`

- [ ] **Step 1: Add all dependencies and run go mod tidy**

```bash
cd /Users/kai/Downloads/go_project/go-im-server && \
go get github.com/gin-gonic/gin \
       github.com/gorilla/websocket \
       github.com/golang-jwt/jwt/v5 \
       gorm.io/gorm \
       gorm.io/driver/mysql \
       github.com/spf13/viper \
       golang.org/x/crypto \
       go.uber.org/zap && \
go mod tidy
```

- [ ] **Step 2: Verify go.mod has all dependencies**

```bash
cd /Users/kai/Downloads/go_project/go-im-server && go build ./... 2>&1 | head -5
```

Expected: `go build` succeeds (no source files yet, but module should be valid).

---

### Task 2: Config loading

**Files:**
- Create: `config/config.go`
- Create: `config.yaml`

- [ ] **Step 1: Write config struct and viper loader**

Create `config/config.go`:

```go
package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Upload   UploadConfig   `mapstructure:"upload"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

type UploadConfig struct {
	Path string `mapstructure:"path"`
}

var AppConfig *Config

func Load() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.SetDefault("server.port", 8080)
	viper.SetDefault("upload.path", "./uploads")
	viper.SetDefault("jwt.secret", "change-this-secret-in-production")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	AppConfig = &Config{}
	if err := viper.Unmarshal(AppConfig); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	if err := os.MkdirAll(AppConfig.Upload.Path, 0755); err != nil {
		return fmt.Errorf("创建上传目录失败: %w", err)
	}

	return nil
}
```

- [ ] **Step 2: Write config.yaml**

Create `config.yaml`:

```yaml
server:
  port: 8080

database:
  dsn: "root:password@tcp(127.0.0.1:3306)/im_server?charset=utf8mb4&parseTime=True&loc=Local"

jwt:
  secret: "your-jwt-secret-key-change-me"

upload:
  path: "./uploads"
```

- [ ] **Step 3: Verify compilation**

```bash
cd /Users/kai/Downloads/go_project/go-im-server && go build ./config/...
```

Expected: builds successfully.

---

### Task 3: Common package — response, error codes, snowflake ID

**Files:**
- Create: `common/response.go`
- Create: `common/errcode.go`
- Create: `common/snowflake.go`

- [ ] **Step 1: Write unified response helpers**

Create `common/response.go`:

```go
package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Msg: "success", Data: data})
}

func Error(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{Code: code, Msg: msg})
}

func ErrorWithStatus(c *gin.Context, httpStatus int, code int, msg string) {
	c.JSON(httpStatus, Response{Code: code, Msg: msg})
}
```

- [ ] **Step 2: Write error codes**

Create `common/errcode.go`:

```go
package common

const (
	// 用户模块 1xxx
	ErrUserNotFound      = 1001
	ErrPasswordWrong     = 1002
	ErrUsernameExists    = 1003
	ErrUserNotLoggedIn   = 1004

	// 好友模块 2xxx
	ErrAlreadyApplied    = 2001
	ErrNotFriend         = 2002
	ErrAlreadyFriend     = 2003
	ErrCantAddSelf       = 2004

	// 消息模块 3xxx
	ErrMessageSendFailed = 3001
	ErrInvalidFileType   = 3002
	ErrFileTooLarge      = 3003

	// 群组模块 4xxx
	ErrGroupNotFound     = 4001
	ErrAlreadyInGroup    = 4002
	ErrNotInGroup        = 4003

	// 通用 5xxx
	ErrInvalidParam      = 5000
	ErrUnauthorized      = 5001

	// 系统 9xxx
	ErrInternal          = 9999
)

var errMsg = map[int]string{
	ErrUserNotFound:    "用户不存在",
	ErrPasswordWrong:   "密码错误",
	ErrUsernameExists:  "用户名已存在",
	ErrUserNotLoggedIn: "用户未登录",
	ErrAlreadyApplied:  "已发送过好友申请",
	ErrNotFriend:       "不是好友",
	ErrAlreadyFriend:   "已经是好友",
	ErrCantAddSelf:     "不能添加自己为好友",
	ErrMessageSendFailed: "消息发送失败",
	ErrInvalidFileType: "不支持的文件类型",
	ErrFileTooLarge:    "文件大小超出限制",
	ErrGroupNotFound:   "群组不存在",
	ErrAlreadyInGroup:  "已在群组中",
	ErrNotInGroup:      "不在群组中",
	ErrInvalidParam:    "参数错误",
	ErrUnauthorized:    "未授权",
	ErrInternal:        "服务器内部错误",
}

func GetErrMsg(code int) string {
	if msg, ok := errMsg[code]; ok {
		return msg
	}
	return "未知错误"
}
```

- [ ] **Step 3: Write snowflake ID generator**

Create `common/snowflake.go`:

```go
package common

import (
	"sync"
	"time"
)

const (
	epoch          = int64(1704038400000) // 2024-01-01 00:00:00 UTC
	workerBits     = 5
	maxWorker      = -1 ^ (-1 << workerBits)
	sequenceBits   = 12
	sequenceMask   = -1 ^ (-1 << sequenceBits)
	workerShift    = sequenceBits
	timestampShift = sequenceBits + workerBits
)

type Snowflake struct {
	mu        sync.Mutex
	workerID  int64
	sequence  int64
	lastStamp int64
}

func NewSnowflake(workerID int64) *Snowflake {
	if workerID < 0 || workerID > maxWorker {
		workerID = 0
	}
	return &Snowflake{workerID: workerID}
}

func (s *Snowflake) NextID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()
	if now < s.lastStamp {
		now = s.lastStamp
	}

	if now == s.lastStamp {
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			for now <= s.lastStamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		s.sequence = 0
	}

	s.lastStamp = now
	return ((now - epoch) << timestampShift) | (s.workerID << workerShift) | s.sequence
}

var defaultSF = NewSnowflake(1)

func GenID() int64 {
	return defaultSF.NextID()
}
```

- [ ] **Step 4: Verify compilation**

```bash
cd /Users/kai/Downloads/go_project/go-im-server && go build ./common/...
```

Expected: builds successfully.

---

### Task 4: Database initialization

**Files:**
- Create: `db/mysql.go`

- [ ] **Step 1: Write DB init with AutoMigrate**

Create `db/mysql.go`:

```go
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
```

- [ ] **Step 2: Verify compilation**

```bash
cd /Users/kai/Downloads/go_project/go-im-server && go build ./db/...
```

Expected: fails — models not defined yet. This is expected; models come next.

---

### Task 5: User, Friend, Group, Message, Conversation models

**Files:**
- Create: `model/user.go`
- Create: `model/friend.go`
- Create: `model/group.go`
- Create: `model/message.go`
- Create: `model/conversation.go`

- [ ] **Step 1: Write User model**

Create `model/user.go`:

```go
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
```

- [ ] **Step 2: Write Friend models**

Create `model/friend.go`:

```go
package model

import "time"

type FriendApplication struct {
	ID         int64      `gorm:"primaryKey;autoIncrement:false" json:"id"`
	FromUserID int64      `gorm:"index;not null" json:"from_user_id"`
	ToUserID   int64      `gorm:"index;not null" json:"to_user_id"`
	Status     int8       `gorm:"default:0;not null" json:"status"` // 0-待处理 1-已同意 2-已拒绝
	Reason     string     `gorm:"size:255" json:"reason"`
	HandledAt  *time.Time `json:"handled_at"`
	CreatedAt  time.Time  `json:"created_at"`

	FromUser *User `gorm:"foreignKey:FromUserID" json:"from_user,omitempty"`
	ToUser   *User `gorm:"foreignKey:ToUserID" json:"to_user,omitempty"`
}

type Friend struct {
	ID        int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	UserID    int64     `gorm:"uniqueIndex:idx_user_friend;not null" json:"user_id"`
	FriendID  int64     `gorm:"uniqueIndex:idx_user_friend;not null" json:"friend_id"`
	CreatedAt time.Time `json:"created_at"`

	Friend *User `gorm:"foreignKey:FriendID" json:"friend,omitempty"`
}
```

- [ ] **Step 3: Write Group models**

Create `model/group.go`:

```go
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
	Role     int8      `gorm:"default:0;not null" json:"role"` // 0-成员 1-群主
	JoinedAt time.Time `json:"joined_at"`

	User  *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Group *Group `gorm:"foreignKey:GroupID" json:"group,omitempty"`
}
```

- [ ] **Step 4: Write Message model**

Create `model/message.go`:

```go
package model

import "time"

type Message struct {
	ID          int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	SenderID    int64     `gorm:"index;not null" json:"sender_id"`
	TargetType  int8      `gorm:"not null" json:"target_type"`  // 1-单聊 2-群聊
	TargetID    int64     `gorm:"not null" json:"target_id"`
	ContentType int8      `gorm:"not null" json:"content_type"` // 1-文本 2-图片
	Content     string    `gorm:"type:text" json:"content"`
	CreatedAt   time.Time `gorm:"index:idx_target_msg,priority:3" json:"created_at"`

	Sender *User `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
}
```

- [ ] **Step 5: Write Conversation model**

Create `model/conversation.go`:

```go
package model

import "time"

type Conversation struct {
	ID          int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	UserID      int64     `gorm:"uniqueIndex:idx_user_target,priority:1;not null" json:"user_id"`
	TargetType  int8      `gorm:"uniqueIndex:idx_user_target,priority:2;not null" json:"target_type"`
	TargetID    int64     `gorm:"uniqueIndex:idx_user_target,priority:3;not null" json:"target_id"`
	LastMessage string    `gorm:"size:255" json:"last_message"`
	UnreadCount int       `gorm:"default:0" json:"unread_count"`
	UpdatedAt   time.Time `json:"updated_at"`
}
```

- [ ] **Step 6: Verify compilation**

```bash
cd /Users/kai/Downloads/go_project/go-im-server && go build ./model/...
```

Expected: builds successfully.

---

### Task 6: Middleware — JWT, CORS, Recovery

**Files:**
- Create: `middleware/auth.go`
- Create: `middleware/cors.go`
- Create: `middleware/recovery.go`

- [ ] **Step 1: Write JWT auth middleware**

Create `middleware/auth.go`:

```go
package middleware

import (
	"net/http"
	"strings"

	"go-im-server/common"
	"go-im-server/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			common.ErrorWithStatus(c, http.StatusUnauthorized, common.ErrUnauthorized, common.GetErrMsg(common.ErrUnauthorized))
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(config.AppConfig.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			common.ErrorWithStatus(c, http.StatusUnauthorized, common.ErrUnauthorized, common.GetErrMsg(common.ErrUnauthorized))
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			common.ErrorWithStatus(c, http.StatusUnauthorized, common.ErrUnauthorized, common.GetErrMsg(common.ErrUnauthorized))
			c.Abort()
			return
		}

		uidFloat, ok := claims["uid"].(float64)
		if !ok {
			common.ErrorWithStatus(c, http.StatusUnauthorized, common.ErrUnauthorized, common.GetErrMsg(common.ErrUnauthorized))
			c.Abort()
			return
		}

		c.Set("uid", int64(uidFloat))
		c.Next()
	}
}

func ParseToken(tokenStr string) (int64, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWT.Secret), nil
	})
	if err != nil || !token.Valid {
		return 0, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, err
	}
	uidFloat, ok := claims["uid"].(float64)
	if !ok {
		return 0, err
	}
	return int64(uidFloat), nil
}
```

- [ ] **Step 2: Write CORS middleware**

Create `middleware/cors.go`:

```go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
```

- [ ] **Step 3: Write Recovery middleware**

Create `middleware/recovery.go`:

```go
package middleware

import (
	"go-im-server/common"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Logger, _ = zap.NewProduction()

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				Logger.Error("panic recovered", zap.Any("error", err))
				common.Error(c, common.ErrInternal, common.GetErrMsg(common.ErrInternal))
				c.Abort()
			}
		}()
		c.Next()
	}
}
```

- [ ] **Step 4: Verify compilation**

```bash
cd /Users/kai/Downloads/go_project/go-im-server && go build ./middleware/...
```

Expected: builds successfully.

---

### Task 7: Router setup

**Files:**
- Create: `router/router.go`

- [ ] **Step 1: Write router with all route groups**

Create `router/router.go`:

```go
package router

import (
	"go-im-server/controller"
	"go-im-server/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(middleware.CORS())
	r.Use(middleware.Recovery())

	// 无需鉴权的路由
	api := r.Group("/api")
	{
		api.POST("/user/register", controller.Register)
		api.POST("/user/login", controller.Login)
	}

	// 静态文件
	r.Static("/uploads", "./uploads")

	// 需要 JWT 鉴权的路由
	auth := api.Group("")
	auth.Use(middleware.JWTAuth())
	{
		// 用户
		auth.GET("/user/info", controller.GetUserInfo)
		auth.PUT("/user/update", controller.UpdateUser)
		auth.GET("/user/search", controller.SearchUser)

		// 好友
		auth.POST("/friend/apply", controller.ApplyFriend)
		auth.POST("/friend/handle", controller.HandleFriend)
		auth.GET("/friend/applications", controller.FriendApplications)
		auth.GET("/friend/list", controller.FriendList)
		auth.DELETE("/friend/delete", controller.DeleteFriend)

		// 群组
		auth.POST("/group/create", controller.CreateGroup)
		auth.POST("/group/join", controller.JoinGroup)
		auth.POST("/group/leave", controller.LeaveGroup)
		auth.GET("/group/info", controller.GroupInfo)
		auth.GET("/group/members", controller.GroupMembers)
		auth.GET("/group/mine", controller.MyGroups)

		// 消息
		auth.GET("/message/history", controller.MessageHistory)
		auth.GET("/message/conversations", controller.Conversations)
		auth.POST("/message/upload", controller.UploadImage)
	}

	return r
}
```

- [ ] **Step 2: Verify**

```bash
cd /Users/kai/Downloads/go_project/go-im-server && go build ./router/...
```

Expected: fails — controllers not defined yet. This is expected.

---

### Task 8: User service

**Files:**
- Create: `service/user.go`

- [ ] **Step 1: Write user service**

Create `service/user.go`:

```go
package service

import (
	"errors"
	"time"

	"go-im-server/common"
	"go-im-server/config"
	"go-im-server/db"
	"go-im-server/model"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Nickname string `json:"nickname"`
}

type UserLoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserLoginResp struct {
	Token string      `json:"token"`
	User  model.User  `json:"user"`
}

type UserUpdateReq struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

func Register(req *UserRegisterReq) error {
	var exist model.User
	if err := db.DB.Where("username = ?", req.Username).First(&exist).Error; err == nil {
		return errors.New(common.GetErrMsg(common.ErrUsernameExists))
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := model.User{
		ID:        common.GenID(),
		Username:  req.Username,
		Password:  string(hashed),
		Nickname:  req.Nickname,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if user.Nickname == "" {
		user.Nickname = req.Username
	}

	return db.DB.Create(&user).Error
}

func Login(req *UserLoginReq) (*UserLoginResp, error) {
	var user model.User
	if err := db.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(common.GetErrMsg(common.ErrUserNotFound))
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New(common.GetErrMsg(common.ErrPasswordWrong))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": user.ID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenStr, err := token.SignedString([]byte(config.AppConfig.JWT.Secret))
	if err != nil {
		return nil, err
	}

	return &UserLoginResp{Token: tokenStr, User: user}, nil
}

func GetUserByID(uid int64) (*model.User, error) {
	var user model.User
	if err := db.DB.First(&user, uid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(common.GetErrMsg(common.ErrUserNotFound))
		}
		return nil, err
	}
	return &user, nil
}

func UpdateUser(uid int64, req *UserUpdateReq) error {
	updates := map[string]interface{}{}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if len(updates) == 0 {
		return errors.New(common.GetErrMsg(common.ErrInvalidParam))
	}
	return db.DB.Model(&model.User{}).Where("id = ?", uid).Updates(updates).Error
}

func SearchUsers(keyword string) ([]model.User, error) {
	var users []model.User
	if err := db.DB.Where("username LIKE ? OR nickname LIKE ?", "%"+keyword+"%", "%"+keyword+"%").
		Select("id, username, nickname, avatar").Limit(20).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd /Users/kai/Downloads/go_project/go-im-server && go build ./service/...
```

Expected: builds successfully.

---

### Task 9: Friend service

**Files:**
- Create: `service/friend.go`

- [ ] **Step 1: Write friend service**

Create `service/friend.go`:

```go
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
	Agree        bool  `json:"agree"`
}

func ApplyFriend(fromUID int64, req *ApplyFriendReq) error {
	if fromUID == req.ToUserID {
		return errors.New(common.GetErrMsg(common.ErrCantAddSelf))
	}

	// 检查目标用户是否存在
	if _, err := GetUserByID(req.ToUserID); err != nil {
		return err
	}

	// 检查是否已是好友
	var friend model.Friend
	if err := db.DB.Where("user_id = ? AND friend_id = ?", fromUID, req.ToUserID).First(&friend).Error; err == nil {
		return errors.New(common.GetErrMsg(common.ErrAlreadyFriend))
	}

	// 检查是否有待处理的申请
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

		// 双向插入 friends 表（事务）
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
```

- [ ] **Step 2: Verify compilation**

```bash
cd /Users/kai/Downloads/go_project/go-im-server && go build ./service/...
```

Expected: builds successfully.

---

### Task 10: Group service

**Files:**
- Create: `service/group.go`

- [ ] **Step 1: Write group service**

Create `service/group.go`:

```go
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
			Role:     1, // 群主
			JoinedAt: now,
		}
		return tx.Create(&member).Error
	})

	return &group, err
}

func JoinGroup(userID, groupID int64) error {
	// 检查群是否存在
	var group model.Group
	if err := db.DB.First(&group, groupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(common.GetErrMsg(common.ErrGroupNotFound))
		}
		return err
	}

	// 检查是否已加入
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
```

- [ ] **Step 2: Verify compilation**

```bash
cd /Users/kai/Downloads/go_project/go-im-server && go build ./service/...
```

Expected: builds successfully.

---

### Task 11: Message service

**Files:**
- Create: `service/message.go`
- Create: `service/uploader.go`

- [ ] **Step 1: Write message service**

Create `service/message.go`:

```go
package service

import (
	"time"

	"go-im-server/common"
	"go-im-server/db"
	"go-im-server/model"
)

type MessageHistoryReq struct {
	TargetType int8  `form:"target_type" binding:"required,oneof=1 2"`
	TargetID   int64 `form:"target_id" binding:"required"`
	Page       int   `form:"page" binding:"min=1"`
	PageSize   int   `form:"page_size" binding:"min=1,max=50"`
}

type MessageHistoryResp struct {
	List       []model.Message `json:"list"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
}

type ConversationInfo struct {
	ID          int64     `json:"id"`
	TargetType  int8      `json:"target_type"`
	TargetID    int64     `json:"target_id"`
	LastMessage string    `json:"last_message"`
	UnreadCount int       `json:"unread_count"`
	UpdatedAt   time.Time `json:"updated_at"`
	TargetName  string    `json:"target_name"`
	TargetAvatar string   `json:"target_avatar"`
}

func SaveMessage(msg *model.Message) error {
	msg.ID = common.GenID()
	if err := db.DB.Create(msg).Error; err != nil {
		return err
	}

	// 更新自己发送方的 conversation
	upsertConversation(msg.SenderID, msg.TargetType, msg.TargetID, msg.Content, msg.ContentType, 0, msg.CreatedAt)

	if msg.TargetType == 1 {
		// 单聊：更新对方 conversation，未读数+1
		upsertConversation(msg.TargetID, msg.TargetType, msg.SenderID, msg.Content, msg.ContentType, 1, msg.CreatedAt)
	}

	return nil
}

func upsertConversation(userID int64, targetType int8, targetID int64, lastMsg string, contentType int8, unreadInc int, now time.Time) {
	var existing model.Conversation
	err := db.DB.Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		First(&existing).Error

	if err != nil {
		// 新建会话
		db.DB.Create(&model.Conversation{
			ID:          common.GenID(),
			UserID:      userID,
			TargetType:  targetType,
			TargetID:    targetID,
			LastMessage: truncateContent(lastMsg, contentType),
			UnreadCount: unreadInc,
			UpdatedAt:   now,
		})
	} else {
		// 更新已有会话
		updates := map[string]interface{}{
			"last_message": truncateContent(lastMsg, contentType),
			"updated_at":   now,
		}
		if unreadInc > 0 {
			updates["unread_count"] = existing.UnreadCount + unreadInc
		}
		db.DB.Model(&existing).Updates(updates)
	}
}

func truncateContent(s string, contentType int8) string {
	if contentType == 2 {
		return "[图片]"
	}
	runes := []rune(s)
	if len(runes) > 50 {
		return string(runes[:50]) + "..."
	}
	return s
}

func MessageHistory(req *MessageHistoryReq) (*MessageHistoryResp, error) {
	var total int64
	var msgs []model.Message

	query := db.DB.Model(&model.Message{}).
		Where("target_type = ? AND target_id = ?", req.TargetType, req.TargetID)

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if err := query.Preload("Sender").Order("created_at DESC").
		Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&msgs).Error; err != nil {
		return nil, err
	}

	// 反转顺序，让最早的在前面
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	return &MessageHistoryResp{
		List: msgs, Total: total, Page: req.Page, PageSize: req.PageSize,
	}, nil
}

func Conversations(userID int64) ([]ConversationInfo, error) {
	var convs []model.Conversation
	if err := db.DB.Where("user_id = ?", userID).Order("updated_at DESC").Find(&convs).Error; err != nil {
		return nil, err
	}

	result := make([]ConversationInfo, 0, len(convs))
	for _, conv := range convs {
		info := ConversationInfo{
			ID: conv.ID, TargetType: conv.TargetType, TargetID: conv.TargetID,
			LastMessage: conv.LastMessage, UnreadCount: conv.UnreadCount, UpdatedAt: conv.UpdatedAt,
		}
		if conv.TargetType == 1 {
			if user, err := GetUserByID(conv.TargetID); err == nil {
				info.TargetName = user.Nickname
				info.TargetAvatar = user.Avatar
			}
		} else {
			if group, err := GroupInfo(conv.TargetID); err == nil {
				info.TargetName = group.Name
				info.TargetAvatar = group.Avatar
			}
		}
		result = append(result, info)
	}
	return result, nil
}

func ClearUnread(userID int64, targetType int8, targetID int64) error {
	return db.DB.Model(&model.Conversation{}).
		Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		Update("unread_count", 0).Error
}

// UpsertGroupConversation 更新群聊所有成员的会话记录
func UpsertGroupConversation(memberIDs []int64, groupID int64, content string, contentType int8, now time.Time) {
	for _, mid := range memberIDs {
		upsertConversation(mid, 2, groupID, content, contentType, 1, now)
	}
}
```

Create `service/uploader.go`:

```go
package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"go-im-server/common"
	"go-im-server/config"
)

type Uploader interface {
	Upload(file multipart.File, header *multipart.FileHeader) (string, error)
}

type LocalUploader struct{}

func (l *LocalUploader) Upload(file multipart.File, header *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
		return "", errors.New(common.GetErrMsg(common.ErrInvalidFileType))
	}

	if header.Size > 10*1024*1024 {
		return "", errors.New(common.GetErrMsg(common.ErrFileTooLarge))
	}

	filename := fmt.Sprintf("%d%s", common.GenID(), ext)
	dst := filepath.Join(config.AppConfig.Upload.Path, filename)

	dstFile, err := createFile(dst)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, file); err != nil {
		return "", err
	}

	return "/uploads/" + filename, nil
}

func createFile(path string) (io.WriteCloser, error) {
	return fileOpen(path)
}

// 兼容层封装，便于后续替换
var DefaultUploader Uploader = &LocalUploader{}
```

Create the helper file `service/file_helper.go`:

```go
package service

import (
	"os"
)

func fileOpen(path string) (*os.File, error) {
	return os.Create(path)
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd /Users/kai/Downloads/go_project/go-im-server && go build ./service/...
```

Expected: builds successfully.

---

### Task 12: User controller

**Files:**
- Create: `controller/user.go`

- [ ] **Step 1: Write user controller**

Create `controller/user.go`:

```go
package controller

import (
	"net/http"
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

// getUID 工具函数
func getUID(c *gin.Context) int64 {
	return c.GetInt64("uid")
}

// getQueryInt64 从 query 参数获取 int64
func getQueryInt64(c *gin.Context, key string) (int64, error) {
	return strconv.ParseInt(c.Query(key), 10, 64)
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd /Users/kai/Downloads/go_project/go-im-server && go build ./controller/...
```

Expected: fails — other controllers not defined yet. Continue to next task.

---

### Task 13: Friend controller

**Files:**
- Create: `controller/friend.go`

- [ ] **Step 1: Write friend controller**

Create `controller/friend.go`:

```go
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
```

- [ ] **Step 2: Verify compilation**

```bash
cd /Users/kai/Downloads/go_project/go-im-server && go build ./controller/...
```

Expected: fails — remaining controllers not defined. Continue.

---

### Task 14: Group controller

**Files:**
- Create: `controller/group.go`

- [ ] **Step 1: Write group controller**

Create `controller/group.go`:

```go
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
```

- [ ] **Step 2: Verify compilation**

```bash
cd /Users/kai/Downloads/go_project/go-im-server && go build ./controller/...
```

Expected: fails — message controller not defined. Continue.

---

### Task 15: Message controller

**Files:**
- Create: `controller/message.go`

- [ ] **Step 1: Write message controller**

Create `controller/message.go`:

```go
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

	// 拉取历史消息后清除未读数
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
```

- [ ] **Step 2: Verify controller compilation**

```bash
cd /Users/kai/Downloads/go_project/go-im-server && go build ./controller/...
```

Expected: builds all controllers successfully.

---

### Task 16: WebSocket — types and hub

**Files:**
- Create: `ws/types.go`
- Create: `ws/hub.go`
- Create: `ws/client.go`

- [ ] **Step 1: Write WS message types**

Create `ws/types.go`:

```go
package ws

const (
	TypeSingleMsg = 1 // 单聊消息
	TypeGroupMsg  = 2 // 群聊消息
	TypeSysNotify = 3 // 系统通知（好友申请结果、入群通知等）
)

const (
	TargetTypeSingle = 1 // 单聊
	TargetTypeGroup  = 2 // 群聊
)

const (
	ContentTypeText  = 1 // 文本
	ContentTypeImage = 2 // 图片
)

type Message struct {
	Type        int8   `json:"type"`
	From        int64  `json:"from"`
	To          int64  `json:"to"`
	TargetType  int8   `json:"target_type"`
	ContentType int8   `json:"content_type"`
	Content     string `json:"content"`
	Timestamp   int64  `json:"timestamp"`
}
```

- [ ] **Step 2: Write WS Hub**

Create `ws/hub.go`:

```go
package ws

import (
	"encoding/json"
	"sync"

	"go.uber.org/zap"
)

var hubLog, _ = zap.NewProduction()

type Hub struct {
	clients    map[int64]*Client
	mu         sync.RWMutex
	Register   chan *Client
	Unregister chan *Client
}

var DefaultHub = &Hub{
	clients:    make(map[int64]*Client),
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			h.mu.Unlock()
			h.broadcastOnlineStatus(client.UserID, true)

		case client := <-h.Unregister:
			h.mu.Lock()
			if c, ok := h.clients[client.UserID]; ok && c == client {
				delete(h.clients, client.UserID)
			}
			h.mu.Unlock()
			close(client.Send)
			h.broadcastOnlineStatus(client.UserID, false)
		}
	}
}

func (h *Hub) IsOnline(uid int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[uid]
	return ok
}

func (h *Hub) SendToUser(uid int64, msg interface{}) {
	h.mu.RLock()
	client, ok := h.clients[uid]
	h.mu.RUnlock()
	if ok {
		data, err := json.Marshal(msg)
		if err != nil {
			hubLog.Error("序列化消息失败", zap.Error(err))
			return
		}
		select {
		case client.Send <- data:
		default:
		}
	}
}

func (h *Hub) broadcastOnlineStatus(uid int64, online bool) {
	// 通知好友在线状态变更（简化实现：不发好友列表遍历，仅做记录）
}
```

- [ ] **Step 3: Verify compilation**

```bash
cd /Users/kai/Downloads/go_project/go-im-server && go build ./ws/...
```

Expected: fails — client.go not defined. Continue.

---

### Task 17: WebSocket — client

**Files:**
- Create: `ws/client.go`

- [ ] **Step 1: Write WS client with heartbeat**

Create `ws/client.go`:

```go
package ws

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 90 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 4096
)

type Client struct {
	UserID int64
	Conn   *websocket.Conn
	Hub    *Hub
	Send   chan []byte
}

func NewClient(userID int64, conn *websocket.Conn) *Client {
	return &Client{
		UserID: userID,
		Conn:   conn,
		Hub:    DefaultHub,
		Send:   make(chan []byte, 256),
	}
}

func (c *Client) ReadPump(msgHandler func(*Client, []byte)) {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msgBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				hubLog.Error("ws读取错误", zap.Error(err))
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			continue
		}
		msg.From = c.UserID
		msg.Timestamp = time.Now().Unix()

		msgHandler(c, msgBytes)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// 便捷方法：发送 JSON 消息
func SendJSONToUser(uid int64, msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	// 包装为 JSON 后再发送
	DefaultHub.SendToUser(uid, json.RawMessage(data))
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd /Users/kai/Downloads/go_project/go-im-server && go build ./ws/...
```

Expected: builds successfully. If there is an unused `fmt` import in client.go, remove it.

---

### Task 18: WebSocket — message handler

**Files:**
- Create: `ws/handler.go`

- [ ] **Step 1: Write message handler**

Create `ws/handler.go`:

```go
package ws

import (
	"encoding/json"
	"time"

	"go-im-server/model"
	"go-im-server/service"

	"go.uber.org/zap"
)

var handlerLog, _ = zap.NewProduction()

func HandleMessage(client *Client, raw []byte) {
	var msg Message
	if err := json.Unmarshal(raw, &msg); err != nil {
		handlerLog.Error("解析WS消息失败", zap.Error(err))
		return
	}

	msg.From = client.UserID
	msg.Timestamp = time.Now().Unix()

	now := time.Now()

	// 保存数据模型
	msgModel := &model.Message{
		SenderID:    client.UserID,
		TargetType:  msg.TargetType,
		TargetID:    msg.To,
		ContentType: msg.ContentType,
		Content:     msg.Content,
		CreatedAt:   now,
	}

	switch msg.Type {
	case TypeSingleMsg:
		// 单聊：校验好友关系
		if !service.IsFriend(client.UserID, msg.To) {
			SendJSONToUser(client.UserID, map[string]interface{}{
				"type": TypeSysNotify, "content": "对方不是您的好友",
			})
			return
		}

		if err := service.SaveMessage(msgModel); err != nil {
			handlerLog.Error("保存消息失败", zap.Error(err))
			return
		}

		// 推送给接收方
		if DefaultHub.IsOnline(msg.To) {
			DefaultHub.SendToUser(msg.To, msg)
		}

		// 给自己发 ack
		DefaultHub.SendToUser(client.UserID, msg)

	case TypeGroupMsg:
		// 群聊：校验群成员
		if !service.IsGroupMember(client.UserID, msg.To) {
			return
		}

		if err := service.SaveMessage(msgModel); err != nil {
			handlerLog.Error("保存群消息失败", zap.Error(err))
			return
		}

		// 更新所有群成员的 conversation
		memberIDs, err := service.GetGroupMemberIDs(msg.To, client.UserID)
		if err != nil {
			return
		}
		service.UpsertGroupConversation(memberIDs, msg.To, msg.Content, msg.ContentType, now)

		// 推送给在线成员
		for _, mid := range memberIDs {
			if DefaultHub.IsOnline(mid) {
				DefaultHub.SendToUser(mid, msg)
			}
		}

		// ack 自己
		DefaultHub.SendToUser(client.UserID, msg)

	default:
		// 未知消息类型，忽略
	}
}
```

---

### Task 19: WebSocket — upgrade endpoint

**Files:**
- Create: `ws/upgrade.go`

- [ ] **Step 1: Write WS upgrade handler**

Create `ws/upgrade.go`:

```go
package ws

import (
	"net/http"

	"go-im-server/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWebSocket(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 5001, "msg": "缺少 token"})
		return
	}

	uid, err := middleware.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 5001, "msg": "token 无效"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := NewClient(uid, conn)
	client.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump(HandleMessage)
}
```

- [ ] **Step 2: Add WS route to router**

Modify `router/router.go`, adding the WS endpoint. After the line `r.Static("/uploads", "./uploads")`, add:

```go
// WebSocket 长连接
r.GET("/ws", ws.HandleWebSocket)
```

Also add the import for ws:

```go
import (
	"go-im-server/controller"
	"go-im-server/middleware"
	"go-im-server/ws"

	"github.com/gin-gonic/gin"
)
```

- [ ] **Step 3: Verify compilation**

```bash
cd /Users/kai/Downloads/go_project/go-im-server && go build ./ws/...
```

Expected: builds successfully.

---

### Task 20: Main entry

**Files:**
- Create: `main.go`

- [ ] **Step 1: Write main.go**

Create `main.go`:

```go
package main

import (
	"fmt"

	"go-im-server/config"
	"go-im-server/db"
	"go-im-server/router"
	"go-im-server/ws"

	"go.uber.org/zap"
)

func main() {
	mainLog, _ := zap.NewProduction()
	defer mainLog.Sync()

	// 加载配置
	if err := config.Load(); err != nil {
		mainLog.Fatal("加载配置失败", zap.Error(err))
	}

	// 初始化数据库
	if err := db.Init(); err != nil {
		mainLog.Fatal("数据库连接失败", zap.Error(err))
	}
	mainLog.Info("数据库连接成功")

	// 启动 WebSocket Hub
	go ws.DefaultHub.Run()

	// 启动 HTTP 服务
	r := router.SetupRouter()
	addr := fmt.Sprintf(":%d", config.AppConfig.Server.Port)
	mainLog.Info("服务启动", zap.String("addr", addr))
	if err := r.Run(addr); err != nil {
		mainLog.Fatal("服务启动失败", zap.Error(err))
	}
}
```

- [ ] **Step 2: Build the full project**

```bash
cd /Users/kai/Downloads/go_project/go-im-server && go build ./...
```

Expected: full build success.

---

### Task 21: Fixups and final verification

**Files:**
- Modify: All files that have unused imports or type errors

- [ ] **Step 1: Fix unused imports in ws/client.go**

Read `ws/client.go` and remove unused `fmt` import if present.

- [ ] **Step 2: Remove all unused variable warnings**

Run `go vet ./...` and fix each issue.

```bash
cd /Users/kai/Downloads/go_project/go-im-server && go vet ./...
```

- [ ] **Step 3: Run go fmt on everything**

```bash
cd /Users/kai/Downloads/go_project/go-im-server && go fmt ./...
```

- [ ] **Step 4: Final build**

```bash
cd /Users/kai/Downloads/go_project/go-im-server && go build -o im-server .
```

Expected: builds without errors.

- [ ] **Step 5: Add .gitignore**

Create `.gitignore`:

```
/uploads
im-server
*.exe
.DS_Store
.history/
```

---

### Task 22: Integration test (manual)

**Files:**
- Create: `test/main.go` (optional, for quick verification)

- [ ] **Step 1: Start MySQL**

Ensure MySQL is running and the `im_server` database exists:

```sql
CREATE DATABASE IF NOT EXISTS im_server CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

- [ ] **Step 2: Start server**

```bash
cd /Users/kai/Downloads/go_project/go-im-server && go run main.go
```

Expected: "服务启动" log with port.

- [ ] **Step 3: Test registration**

```bash
curl -X POST http://localhost:8080/api/user/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"123456","nickname":"Alice"}'
```

Expected: `{"code":0,"msg":"success"}`

- [ ] **Step 4: Test login and get token**

```bash
curl -X POST http://localhost:8080/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"123456"}'
```

Expected: `{"code":0,"msg":"success","data":{"token":"...","user":{...}}}`

- [ ] **Step 5: Verify JWT auth works**

```bash
# Without token — should fail
curl http://localhost:8080/api/user/info
# With token — should succeed
curl http://localhost:8080/api/user/info -H "Authorization: Bearer <token>"
```
