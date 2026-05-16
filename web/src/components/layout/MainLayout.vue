<template>
  <div class="main-layout">
    <!-- 顶部栏 -->
    <header class="header">
      <div class="header-left">
        <el-avatar :size="36" :src="userStore.userInfo?.avatar" />
        <span class="header-name">{{ userStore.userInfo?.nickname }}</span>
      </div>
      <div class="header-right">
        <el-button text @click="showUserDialog = true">个人信息</el-button>
        <el-button text @click="handleLogout">退出</el-button>
      </div>
    </header>

    <div class="body">
      <!-- 左侧面板 -->
      <div class="sidebar">
        <el-tabs v-model="activeTab" class="sidebar-tabs">
          <el-tab-pane label="会话" name="chat" />
          <el-tab-pane label="好友" name="friend" />
          <el-tab-pane label="群组" name="group" />
        </el-tabs>

        <div class="sidebar-content">
          <SessionList v-if="activeTab === 'chat'" />
          <FriendView v-else-if="activeTab === 'friend'" />
          <GroupView v-else-if="activeTab === 'group'" />
        </div>
      </div>

      <!-- 中间聊天区 -->
      <div class="chat-area">
        <template v-if="chatStore.activeConv">
          <ChatArea />
        </template>
        <template v-else>
          <div class="empty-chat">
            <el-icon :size="64" color="#ccc"><ChatDotSquare /></el-icon>
            <p>选择一个会话开始聊天</p>
          </div>
        </template>
      </div>

      <!-- 右侧面板 -->
      <div v-if="rightPanel.visible" class="right-panel">
        <component :is="rightPanel.component" v-bind="rightPanel.props" @close="rightPanel.visible = false" />
      </div>
    </div>

    <!-- 个人信息弹窗 -->
    <el-dialog v-model="showUserDialog" title="个人信息" width="400px">
      <el-form :model="profileForm" label-width="60px">
        <el-form-item label="昵称">
          <el-input v-model="profileForm.nickname" />
        </el-form-item>
        <el-form-item label="头像">
          <el-upload :http-request="handleAvatarUpload" :show-file-list="false">
            <el-button size="small">上传头像</el-button>
          </el-upload>
        </el-form-item>
        <el-form-item label="用户名">
          <el-input :model-value="userStore.userInfo?.username" disabled />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showUserDialog = false">取消</el-button>
        <el-button type="primary" @click="handleUpdateProfile">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, provide, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '../../stores/user'
import { useChatStore } from '../../stores/chat'
import { useFriendStore } from '../../stores/friend'
import { connect, onMessage } from '../../ws'
import { uploadImage } from '../../api/message'
import { updateUser } from '../../api/user'

import SessionList from '../session/SessionList.vue'
import FriendView from '../friend/FriendView.vue'
import GroupView from '../group/GroupView.vue'
import ChatArea from '../chat/ChatArea.vue'

const router = useRouter()
const userStore = useUserStore()
const chatStore = useChatStore()
const friendStore = useFriendStore()

const activeTab = ref('chat')
const showUserDialog = ref(false)
const profileForm = reactive({ nickname: '', avatar: '' })

const rightPanel = reactive({
  visible: false,
  component: null,
  props: {},
})

function openRightPanel(comp, props = {}) {
  rightPanel.component = comp
  rightPanel.props = { ...props, close: () => { rightPanel.visible = false } }
  rightPanel.visible = true
}

function closeRightPanel() {
  rightPanel.visible = false
}

provide('openRightPanel', openRightPanel)
defineExpose({ openRightPanel, closeRightPanel })

function handleLogout() {
  userStore.logout()
  router.push('/login')
}

async function handleAvatarUpload(options) {
  try {
    const res = await uploadImage(options.file)
    profileForm.avatar = res.data.url
    ElMessage.success('头像上传成功')
  } catch {
    // handled
  }
}

async function handleUpdateProfile() {
  try {
    await updateUser({ nickname: profileForm.nickname, avatar: profileForm.avatar })
    await userStore.fetchUserInfo()
    ElMessage.success('保存成功')
    showUserDialog.value = false
  } catch {
    // handled
  }
}

// WS 消息处理
let unsub = null
onMounted(async () => {
  if (userStore.isLoggedIn) {
    connect(userStore.token)
    await Promise.all([
      userStore.fetchUserInfo(),
      chatStore.fetchConversations(),
      friendStore.fetchFriends(),
      friendStore.fetchApplications(),
      friendStore.fetchMyGroups(),
    ])
    profileForm.nickname = userStore.userInfo?.nickname || ''
    profileForm.avatar = userStore.userInfo?.avatar || ''
  }

  unsub = onMessage((msg) => {
    if (msg.type === 1 || msg.type === 2) {
      chatStore.appendMessage(msg)
      // 重新加载会话列表
      chatStore.fetchConversations()
    } else if (msg.type === 3) {
      friendStore.fetchApplications()
    }
  })
})

onUnmounted(() => {
  if (unsub) unsub()
})
</script>

<style scoped>
.main-layout {
  height: 100vh;
  display: flex;
  flex-direction: column;
}
.header {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  background: #fff;
  border-bottom: 1px solid #e0e0e0;
  flex-shrink: 0;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}
.header-name {
  font-weight: 600;
  font-size: 16px;
}
.header-right {
  display: flex;
  gap: 8px;
}
.body {
  flex: 1;
  display: flex;
  overflow: hidden;
}
.sidebar {
  width: 300px;
  border-right: 1px solid #e0e0e0;
  display: flex;
  flex-direction: column;
  background: #fafafa;
  flex-shrink: 0;
}
.sidebar-tabs {
  padding: 0 12px;
}
.sidebar-content {
  flex: 1;
  overflow-y: auto;
}
.chat-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #f5f5f5;
}
.empty-chat {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  color: #999;
  gap: 12px;
}
.right-panel {
  width: 320px;
  border-left: 1px solid #e0e0e0;
  background: #fff;
  overflow-y: auto;
  flex-shrink: 0;
}
</style>
