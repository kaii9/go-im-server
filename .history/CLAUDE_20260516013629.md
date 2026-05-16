# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

Go 后端即时通讯系统，基于 Gin + WebSocket + GORM + MySQL + JWT 实现。支持用户注册登录、单聊、群聊、消息推送。

## 架构分层

```
router → controller → service → model → db
```

- **router** — 路由注册、中间件（CORS、JWT 鉴权、全局异常恢复）
- **controller** — 请求参数校验、调用 service、统一响应格式
- **service** — 业务逻辑编排
- **model** — 数据模型定义（对应数据库表）
- **db** — 数据库初始化、连接管理

## 核心目录结构

```
go-im-server/
├── main.go                   # 入口：启动 HTTP/WebSocket 服务
├── config/                   # 配置文件与加载
├── router/                   # 路由注册
├── middleware/                # JWT、CORS、错误恢复等中间件
├── controller/               # 请求处理层
├── service/                  # 业务逻辑层
├── model/                    # 数据模型（GORM）
├── db/                       # MySQL/GORM 初始化
├── common/                   # 统一响应、常量、工具函数
├── ws/                       # WebSocket 管理（连接池、消息转发）
├── go.mod
└── config.yaml
```

## 命令

```bash
# 安装依赖
go mod tidy

# 本地启动
go run main.go

# 编译
go build -o im-server .

# 运行测试（单个包）
go test ./service/... -v

# 运行测试（所有包）
go test ./... -v

# 代码格式化
go fmt ./...

# 代码检查
go vet ./...
```

## 设计约定

- **统一响应格式**：`{"code": 0, "msg": "success", "data": {...}}`
- **WebSocket 消息结构**：`{"type": 1, "from": "uid", "to": "uid/gid", "content": "...", "timestamp": 123}`
- **错误处理**：所有 panic 由 recovery 中间件兜底，业务错误通过统一错误码返回
- **JWT**：登录接口返回 token，其他接口通过 Bearer 方式鉴权
- **数据库**：GORM AutoMigrate 自动建表，dsn 从 config.yaml 读取
