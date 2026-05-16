# Go IM Server

基于 Gin + WebSocket + GORM + MySQL 的即时通讯系统，支持单聊、群聊、好友管理，附带完整 Vue 3 前端。

## 技术栈

| 层 | 技术 |
| --- | --- |
| 后端框架 | Gin |
| 实时通讯 | gorilla/websocket |
| ORM | GORM |
| 数据库 | MySQL |
| 鉴权 | JWT (golang-jwt/v5) |
| 配置管理 | Viper |
| 日志 | zap |
| ID 生成 | Snowflake |
| 前端 | Vue 3 + Vite 5 |
| UI 组件 | Element Plus |
| 状态管理 | Pinia |
| HTTP 客户端 | axios |

## 项目结构

```
go-im-server/
├── main.go                 # 入口：启动 HTTP/WebSocket 服务
├── config.yaml             # 配置文件
├── common/                 # 常量、响应格式、工具函数
│   ├── errcode.go          # 错误码定义
│   ├── response.go         # 统一 JSON 响应
│   └── snowflake.go        # Snowflake ID 生成器
├── config/                 # 配置加载
├── db/                     # MySQL 初始化 + AutoMigrate
├── middleware/             # CORS、JWT 鉴权、Panic 恢复
├── model/                  # 数据模型（User, Friend, Group, Message, Conversation）
├── controller/             # 请求处理层
├── service/                # 业务逻辑层
├── router/                 # 路由注册
├── ws/                     # WebSocket 连接管理、消息路由
│   ├── hub.go              # 连接池（用户ID → Client）
│   ├── client.go           # 读写协程 + 心跳
│   ├── handler.go          # 消息分发校验
│   └── upgrade.go          # HTTP → WS 升级
├── web/                    # Vue 3 前端
│   └── src/
│       ├── api/            # API 封装
│       ├── stores/         # Pinia 状态
│       ├── components/     # UI 组件
│       ├── views/          # 页面视图
│       ├── router/         # 路由配置
│       ├── ws/             # WebSocket 客户端
│       └── App.vue
└── docs/                   # 设计文档
```

## 快速开始

### 前置条件

- Go 1.21+
- MySQL 8.0+
- Node.js 20+
- npm

### 配置

编辑 `config.yaml`：

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

### 启动后端

```bash
# 初始化数据库（需要先创建 im_server 数据库）
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS im_server DEFAULT CHARSET utf8mb4;"

# 启动服务（自动建表）
go mod tidy
go run main.go
```

服务启动在 `http://localhost:8080`，WebSocket 端点 `ws://localhost:8080/ws`。

### 启动前端

```bash
cd web
npm install
npm run dev
```

前端运行在 `http://localhost:5173`，API 和 WebSocket 通过 Vite proxy 转发到后端。

### 构建部署

```bash
# 后端编译
go build -o im-server .

# 前端构建
cd web && npm run build
# 产物在 web/dist/，可部署到 Nginx 或嵌入 Gin
```

## API 文档

所有接口统一响应格式：

```json
{
  "code": 0,
  "msg": "success",
  "data": { ... }
}
```

### 用户

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | /api/user/register | 注册（username, password, nickname） |
| POST | /api/user/login | 登录，返回 token |
| GET | /api/user/info | 获取个人信息（需鉴权） |
| PUT | /api/user/update | 更新资料（nickname, avatar） |
| GET | /api/user/search | 搜索用户（keyword） |

### 好友

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | /api/friend/apply | 发送好友申请（to_user_id, reason） |
| POST | /api/friend/handle | 处理申请（application_id, agree） |
| GET | /api/friend/applications | 申请列表（type: received/sent） |
| GET | /api/friend/list | 好友列表 |
| DELETE | /api/friend/delete | 删除好友 |

### 群组

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | /api/group/create | 创建群组（name） |
| POST | /api/group/join | 加入群组（group_id） |
| POST | /api/group/leave | 退出群组（group_id） |
| GET | /api/group/info | 群组信息（id） |
| GET | /api/group/members | 群成员列表（id） |
| GET | /api/group/mine | 我的群组列表 |

### 消息

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | /api/message/history | 历史消息（target_type, target_id, page, page_size） |
| GET | /api/message/conversations | 会话列表 |
| POST | /api/message/upload | 上传图片（multipart/form-data） |

## WebSocket 协议

### 连接

`ws://localhost:8080/ws?token={jwt_token}`

### 消息格式

```json
{
  "type": 1,
  "from": 123456,
  "to": 789012,
  "target_type": 1,
  "content_type": 1,
  "content": "hello",
  "timestamp": 1700000000
}
```

| 字段 | 说明 |
| --- | --- |
| type | 1-单聊 2-群聊 3-系统通知 |
| target_type | 1-用户 2-群组 |
| content_type | 1-文本 2-图片 |
| from | 发送者 ID（服务端填充） |
| timestamp | 时间戳（服务端填充） |

## 数据模型

- **User** — 用户（snowflake ID, bcrypt 密码）
- **Friend** — 好友关系（双向记录）
- **FriendApplication** — 好友申请（pending → approved/rejected）
- **Group** — 群组（含 member_count 冗余）
- **GroupMember** — 群成员（role: 0-成员 1-群主）
- **Message** — 消息（含 content_type 区分文本/图片）
- **Conversation** — 会话（冗余 last_message + unread_count）

## 前端功能

- 用户注册 / 登录 / 个人信息编辑
- 头像上传（本地存储）
- 实时单聊、群聊
- 文本消息、图片消息
- 历史消息翻页加载
- 会话列表 + 未读计数
- 好友搜索 / 添加 / 同意或拒绝 / 删除
- 群组创建 / 加入 / 退出
- WebSocket 自动重连

## 错误码

| 范围 | 模块 |
| --- | --- |
| 1xxx | 用户模块 |
| 2xxx | 好友模块 |
| 3xxx | 消息模块 |
| 4xxx | 群组模块 |
| 5xxx | 通用错误 |
| 9xxx | 系统错误 |
