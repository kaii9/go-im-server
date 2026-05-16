# IM 前端架构设计

## 项目概述

Vue 3 单页应用，为 Go IM 后端提供完整的前端界面。桌面端布局，仿微信 PC 版交互风格。

### 功能范围

| 模块 | 要点 |
|------|------|
| 用户 | 注册、登录、退出、个人信息查看/编辑 |
| 好友 | 搜索用户、发起申请、同意/拒绝、好友列表、删除好友 |
| 单聊 | WebSocket 实时文本/图片消息、历史翻页 |
| 群聊 | 创建群、加入群、退出群、群聊消息、成员列表 |
| 会话 | 最近会话列表、未读标记、最后一条消息摘要 |
| 实时 | 在线状态、消息即时推送、离线消息上线拉取 |

### 不包含

- 移动端适配（本期只做桌面端）
- 文件/语音/视频消息
- 消息已读回执
- 群审批、踢人、群公告

---

## 技术栈

| 层 | 选型 | 说明 |
|----|------|------|
| 框架 | Vue 3.4 + Vite 5 | `<script setup>` 语法 |
| UI | Element Plus | 组件库，按需导入 |
| 状态 | Pinia | 用户/会话/好友 store |
| 路由 | Vue Router 4 | 登录/注册/主页三路由 |
| HTTP | axios | 统一拦截 token 和错误 |
| WS | 原生 WebSocket | 单例连接管理 + 心跳 |
| 样式 | Scoped CSS | 少量全局变量 |

---

## 目录结构

```
web/
├── index.html
├── package.json
├── vite.config.js
├── src/
│   ├── main.js
│   ├── App.vue
│   ├── api/                    # HTTP 请求封装
│   │   ├── request.js          # axios 实例 + 拦截器
│   │   ├── user.js
│   │   ├── friend.js
│   │   ├── group.js
│   │   └── message.js
│   ├── ws/
│   │   └── index.js            # WebSocket 单例 + 消息分发
│   ├── stores/
│   │   ├── user.js             # 用户信息 + token + 登录状态
│   │   ├── chat.js             # 会话列表 + 当前会话 + 消息
│   │   └── friend.js           # 好友列表 + 申请列表
│   ├── router/
│   │   └── index.js
│   ├── views/
│   │   ├── Login.vue
│   │   ├── Register.vue
│   │   └── Home.vue            # 主布局：三栏
│   ├── components/
│   │   ├── layout/
│   │   │   └── MainLayout.vue   # 三栏容器
│   │   ├── session/
│   │   │   ├── SessionList.vue  # 会话列表
│   │   │   └── SessionItem.vue  # 单条会话
│   │   ├── chat/
│   │   │   ├── ChatArea.vue     # 聊天区域容器
│   │   │   ├── MessageList.vue  # 消息历史
│   │   │   ├── MessageItem.vue  # 单条消息（文本/图片）
│   │   │   └── ChatInput.vue    # 输入框 + 发送 + 图片上传
│   │   ├── friend/
│   │   │   ├── FriendList.vue   # 好友列表侧栏
│   │   │   ├── FriendSearch.vue # 搜索用户弹窗
│   │   │   └── FriendApply.vue  # 好友申请列表
│   │   └── group/
│   │       ├── CreateGroup.vue  # 创建群弹窗
│   │       └── GroupInfo.vue    # 群信息 + 成员列表
│   └── utils/
│       └── index.js            # 时间格式化等工具
```

---

## 页面与路由

| 路径 | 组件 | 说明 | 守卫 |
|------|------|------|------|
| `/login` | Login.vue | 登录 | 已登录跳转主页 |
| `/register` | Register.vue | 注册 | 已登录跳转主页 |
| `/` | Home.vue | 主页面（三栏） | 未登录跳转登录 |

---

## 布局方案

```
┌─────────────────────────────────────────────────┐
│  status  bar: 头像 + 昵称          [退出]       │
├──────────┬──────────────────┬───────────────────┤
│ 会话列表  │   聊天区域        │  侧边栏            │
│          │                  │  (好友/群/搜索)    │
│ [好友]   │  ◀ 消息历史 ▶    │                    │
│ [群聊]   │                  │                   │
│ ───────  │                  │                   │
│ 会话1 🔴 │                  │                   │
│ 会话2    │  ────────        │                   │
│ 会话3 🔴 │  [输入框]        │                   │
│          │  [发送📎📷]      │                   │
├──────────┴──────────────────┴───────────────────┤
│  状态栏                                           │
└─────────────────────────────────────────────────┘
```

- **左栏**（300px）：Tab 切换「会话」「好友」「群组」，会话列表带未读红点
- **中栏**（flex-grow）：消息历史（滚动到底）+ 输入区域
- **右栏**（320px，可选）：好友详情/群详情/成员列表/申请列表，点击展开

---

## 组件树

```
App.vue
 └─ <router-view>
     ├─ Login.vue
     ├─ Register.vue
     └─ Home.vue
         └─ MainLayout.vue
             ├─ HeaderBar.vue                # 顶部：头像 + 昵称 + 退出
             ├─ SidePanel.vue                # 左栏容器
             │   ├─ TabButtons (会话/好友/群)
             │   ├─ SessionList.vue          # tab=0
             │   │   └─ SessionItem.vue × N
             │   ├─ FriendList.vue           # tab=1
             │   └─ GroupView.vue            # tab=2
             │       └─ GroupItem.vue × N
             ├─ ChatArea.vue                 # 中栏
             │   ├─ ChatHeader.vue           # 会话标题
             │   ├─ MessageList.vue
             │   │   └─ MessageItem.vue × N
             │   └─ ChatInput.vue
             └─ RightPanel.vue               # 右栏（可选）
                 ├─ FriendInfo.vue
                 ├─ GroupInfo.vue
                 ├─ FriendApply.vue          # 申请列表弹窗
                 └─ FriendSearch.vue         # 搜索用户弹窗
```

---

## 数据流

### Pinia Store

**userStore**
```
token, userInfo, isLoggedIn
actions: login(), register(), logout(), fetchUserInfo(), updateProfile()
```

**chatStore**
```
conversations[], activeConversation, messages[]
actions: fetchConversations(), setActiveConv(), sendMessage(), fetchHistory()
         appendMessage()        // WS 收到消息时调用
```

**friendStore**
```
friendList[], applications[] (sent + received)
actions: fetchFriends(), fetchApplications(), applyFriend(), handleApplication()
```

### WebSocket 消息流

```
connect(token) → 建立连接
onmessage → 解析 type:
  1 → chatStore.appendMessage() + 更新会话最后消息
  2 → chatStore.appendMessage() + 更新会话
  3 → friendStore.fetchApplications() (重新拉取申请列表)
onclose → 5 秒后自动重连
```

### HTTP → Pinia 同步

所有 API 成功响应后同步更新对应 store。
axios 拦截器自动注入 `Authorization: Bearer <token>`，401 时自动跳转登录。

---

## API 对接

所有接口路径以 `/api` 开头，baseURL 通过 vite proxy 代理到 `http://localhost:8080`。

```js
// vite.config.js
server: {
  proxy: {
    '/api': 'http://localhost:8080',
    '/uploads': 'http://localhost:8080',
    '/ws': { target: 'ws://localhost:8080', ws: true }
  }
}
```

---

## 组件规范

- 每个 `.vue` 文件使用 `<script setup>` 语法
- 样式使用 `<style scoped>`
- 全局变量定义在 `assets/variables.css`
- 消息组件根据 `content_type` 渲染不同展示（1=文本气泡，2=图片块）
- 图片消息支持点击放大预览（Element Plus el-image）

---

## 依赖清单

```json
{
  "dependencies": {
    "vue": "^3.4",
    "vue-router": "^4",
    "pinia": "^2",
    "element-plus": "^2",
    "axios": "^1"
  },
  "devDependencies": {
    "vite": "^5",
    "@vitejs/plugin-vue": "^5"
  }
}
```
