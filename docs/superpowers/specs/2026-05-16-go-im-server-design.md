# Go IM Server 架构设计

## 项目概述

Go 后端即时通讯系统，基于 Gin + gorilla/websocket + GORM + MySQL + JWT 实现。

### 功能范围

| 模块 | 要点 |
|------|------|
| 用户 | 注册、登录、JWT 鉴权、个人信息更新、用户搜索 |
| 好友 | 双向确认：A 申请 → B 同意/拒绝 / 好友列表 / 删除好友 |
| 单聊 | 文本 + 图片消息，需好友关系 |
| 群组 | 创建群、自由加入、退出群、群聊消息、查看成员 |
| 消息 | 历史分页查询、最近会话列表（含未读数） |
| 离线 | 消息存入 DB，上线自动推送未读 |
| 图片 | 本地磁盘存储，预留云存储接口 |

### 不包含

- 好友审批通知推送到客户端（需 APNs/FCM）
- 消息已读/未读状态（本期只计未读数，不做已读回执）
- 群主审批入群、踢人、群公告
- 视频/语音/文件消息

---

## 架构方案

单机单体：HTTP 和 WebSocket 同进程部署。内存 map 管理在线用户连接池，goroutine + channel 做消息转发。

- 优点：开发快、部署简单、无外部依赖
- 缺点：单机瓶颈
- 演进方向：用户量增长后引入 Redis Pub/Sub 过渡到分布式

---

## 目录结构

```
go-im-server/
├── main.go                     # 入口：启动 HTTP/WebSocket
├── config/                     # 配置文件结构体、viper 加载
│   └── config.go
├── config.yaml                 # YAML 配置（dsn, jwt_secret, port）
├── router/                     # 路由注册 + 中间件绑定
│   └── router.go
├── middleware/                  # JWT、CORS、全局异常恢复
│   ├── auth.go
│   ├── cors.go
│   └── recovery.go
├── controller/                 # 请求参数校验、调用 service、统一响应
│   ├── user.go
│   ├── friend.go
│   ├── message.go
│   └── group.go
├── service/                    # 业务逻辑
│   ├── user.go
│   ├── friend.go
│   ├── message.go
│   └── group.go
├── model/                      # GORM 数据模型
│   ├── user.go
│   ├── friend.go
│   ├── group.go
│   ├── message.go
│   └── conversation.go
├── db/                         # MySQL/GORM 初始化 + AutoMigrate
│   └── mysql.go
├── common/                     # 统一响应、错误码、雪花 ID、工具函数
│   ├── response.go
│   ├── errcode.go
│   └── snowflake.go
├── ws/                         # WebSocket 管理
│   ├── hub.go                  # 连接池、消息路由
│   ├── client.go               # 单连接读写、心跳
│   └── types.go                # 消息结构体定义
├── uploads/                    # 图片上传目录（gitignore）
├── go.mod
└── go.sum
```

---

## 数据模型（7 张表）

### users
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED PK | 雪花算法生成 |
| username | VARCHAR(32) UNIQUE NOT NULL | 登录名 |
| password | VARCHAR(128) NOT NULL | bcrypt 加密 |
| nickname | VARCHAR(64) | 昵称 |
| avatar | VARCHAR(255) | 头像 URL |
| created_at | DATETIME | |
| updated_at | DATETIME | |

### friend_applications
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED PK | |
| from_user_id | BIGINT UNSIGNED | 发起方 |
| to_user_id | BIGINT UNSIGNED | 接收方 |
| status | TINYINT | 0-待处理 1-已同意 2-已拒绝 |
| reason | VARCHAR(255) | 申请附言 |
| handled_at | DATETIME | 处理时间 |
| created_at | DATETIME | |

### friends
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED PK | |
| user_id | BIGINT UNSIGNED | 用户 |
| friend_id | BIGINT UNSIGNED | 好友 |
| created_at | DATETIME | |

> 一条好友关系存两条记录 (A, B) 和 (B, A)，方便双向查询。

### groups
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED PK | |
| name | VARCHAR(64) NOT NULL | 群名称 |
| avatar | VARCHAR(255) | 群头像 |
| owner_id | BIGINT UNSIGNED | 群主 |
| member_count | INT DEFAULT 1 | 成员数冗余 |
| created_at | DATETIME | |
| updated_at | DATETIME | |

### group_members
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED PK | |
| group_id | BIGINT UNSIGNED | |
| user_id | BIGINT UNSIGNED | |
| role | TINYINT | 0-成员 1-群主 |
| joined_at | DATETIME | |

### messages
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED PK | |
| sender_id | BIGINT UNSIGNED | 发送者 |
| target_type | TINYINT | 1-单聊 2-群聊 |
| target_id | BIGINT UNSIGNED | 接收者 ID 或群 ID |
| content_type | TINYINT | 1-文本 2-图片 |
| content | TEXT | 消息体 |
| created_at | DATETIME | |

索引：`(target_type, target_id, created_at)` 用于历史查询分页。

### conversations
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED PK | |
| user_id | BIGINT UNSIGNED | 所属用户 |
| target_type | TINYINT | 1-单聊 2-群聊 |
| target_id | BIGINT UNSIGNED | 对方 ID 或群 ID |
| last_message | VARCHAR(255) | 最后一条消息摘要 |
| unread_count | INT DEFAULT 0 | 未读数 |
| updated_at | DATETIME | 最后消息时间 |

唯一索引：`(user_id, target_type, target_id)`，用于 upsert。

---

## 接口清单

### 用户模块

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| POST | `/api/user/register` | 注册 | 否 |
| POST | `/api/user/login` | 登录 | 否 |
| GET | `/api/user/info` | 当前用户信息 | JWT |
| PUT | `/api/user/update` | 更新昵称/头像 | JWT |
| GET | `/api/user/search?keyword=` | 搜索用户 | JWT |

### 好友模块

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| POST | `/api/friend/apply` | 发起好友申请 | JWT |
| POST | `/api/friend/handle` | 同意/拒绝申请 | JWT |
| GET | `/api/friend/applications?type=sent/received` | 申请列表 | JWT |
| GET | `/api/friend/list` | 好友列表 | JWT |
| DELETE | `/api/friend/delete` | 删除好友 | JWT |

### 消息模块

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| GET | `/api/message/history?target_type=&target_id=&page=&page_size=` | 历史消息 | JWT |
| GET | `/api/message/conversations` | 会话列表 | JWT |
| POST | `/api/message/upload` | 上传图片 | JWT |
| GET | `/uploads/*filepath` | 图片访问 | 否 |

### 群组模块

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| POST | `/api/group/create` | 创建群 | JWT |
| POST | `/api/group/join` | 加入群 | JWT |
| POST | `/api/group/leave` | 退出群 | JWT |
| GET | `/api/group/info?id=` | 群详情 | JWT |
| GET | `/api/group/members?id=` | 群成员列表 | JWT |
| GET | `/api/group/mine` | 我的群列表 | JWT |

### WebSocket

| 路径 | 说明 |
|------|------|
| `ws://host/ws?token=xxx` | 连接升级，query 参数携带 token |

---

## WebSocket 消息协议

```json
{
  "type": 1,
  "from": "sender_id",
  "to": "target_id",
  "target_type": 1,
  "content_type": 1,
  "content": "消息内容",
  "timestamp": 1715932800
}
```

| type | 值 | 方向 |
|------|------|------|
| 单聊消息 | 1 | 客户端 ↔ 服务端 |
| 群聊消息 | 2 | 客户端 ↔ 服务端 |
| 系统通知 | 3 | 服务端 → 客户端（好友申请结果、入群通知等） |

---

## 核心流程

### 注册/登录

1. 注册：校验 username 唯一 → bcrypt 加密密码 → 写入 users 表 → 返回成功
2. 登录：查 users 表 → bcrypt 比对 → 生成 JWT（含 uid，24h 过期）→ 返回 token + 用户信息

### 好友申请

1. A 发起 POST `/api/friend/apply`，写入 friend_applications（status=0）
2. 若 B 在线，通过 WS type=3 推送系统通知
3. B 调用 POST `/api/friend/handle`，更新 status
4. 若同意：friends 表插入 (A,B) 和 (B,A) 两条记录；通过 WS 通知 A 结果

### 消息发送

1. 客户端通过 WS 发送消息 JSON
2. 服务端：校验 JWT → 校验好友关系（单聊）或群成员（群聊）→ 写入 messages 表 → upsert conversations 表
3. 目标在线：直接 WS 推送
4. 目标离线：更新 unread_count，消息已持久化在 messages 表

### 离线消息拉取

1. 用户 WS 连接成功后，服务端查 conversations 表，推送有未读数的会话摘要
2. 客户端按需调用 GET `/api/message/history` 拉取具体消息
3. 服务端将对应 conversation 的 unread_count 清零

### 图片上传

1. POST `/api/message/upload`（multipart/form-data），校验格式(jpg/png/gif)和大小(10MB)
2. 生成雪花 ID 文件名 → 写入 `./uploads/` → 返回 `/uploads/{filename}` 访问 URL
3. 客户端拿到 URL 后，通过 WS 消息发送（content_type=2，content=URL）

### 在线状态

- 连接池：`map[userID]*Client`，sync.RWMutex 保护
- 上线：WS 升级成功 → 注册到 map → 通过 WS 通知好友"online"
- 离线：WS 断开 → 从 map 移除 → 通知好友 "offline"
- 心跳：客户端每 30s 发送 ping，服务端 90s 无消息则断开连接

---

## 统一响应格式

```json
{"code": 0, "msg": "success", "data": {}}
```

错误码段：
- 0：成功
- 1xxx：用户模块（1001-用户不存在、1002-密码错误、1003-用户名已存在）
- 2xxx：好友模块（2001-已申请、2002-非好友）
- 3xxx：消息模块
- 4xxx：群组模块（4001-群不存在、4002-已在群中）
- 5xxx：通用（5000-参数错误、5001-未授权）
- 9xxx：系统（9999-内部错误）

---

## 图片存储策略

定义 `Uploader` interface：

```go
type Uploader interface {
    Upload(file multipart.File, header *multipart.FileHeader) (url string, err error)
}
```

本期实现 `LocalUploader`（本地磁盘），后续实现 `OSSUploader` 或 `S3Uploader` 只需实现此接口并替换注入。

---

## 依赖项

```
github.com/gin-gonic/gin
github.com/gorilla/websocket
github.com/golang-jwt/jwt/v5
gorm.io/gorm
gorm.io/driver/mysql
github.com/spf13/viper
golang.org/x/crypto
go.uber.org/zap
```

---

## 验收标准

1. 用户可注册、登录，获取 JWT token
2. JWT 鉴权中间件保护除注册/登录外的所有 HTTP 接口
3. 用户可搜索、发起好友申请、同意/拒绝、查看好友列表、删除好友
4. 好友之间可通过 WebSocket 发送文本和图片消息，实时接收
5. 非好友无法发送单聊消息
6. 用户可创建群、加入群、退出群、群内发送消息
7. 群聊消息推送给群内所有成员（发送者除外）
8. 离线消息持久化，上线后通过会话列表可见未读数，可拉取历史
9. 图片上传返回可访问 URL，图片文件存储在本地磁盘
10. 所有接口返回统一 `{code, msg, data}` 格式
11. 未授权请求返回 401
12. Panic 由全局 recovery 中间件捕获返回 500
