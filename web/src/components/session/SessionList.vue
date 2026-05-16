<template>
  <div class="session-list">
    <div
      v-for="conv in chatStore.conversations"
      :key="conv.id"
      class="session-item"
      :class="{ active: isActive(conv) }"
      @click="selectConv(conv)"
    >
      <el-avatar :size="40" :src="conv.target_avatar" />
      <div class="session-info">
        <div class="session-top">
          <span class="session-name">{{ conv.target_name || '未知' }}</span>
          <span class="session-time">{{ formatTime(conv.updated_at) }}</span>
        </div>
        <div class="session-bottom">
          <span class="session-last">{{ conv.last_message }}</span>
          <el-badge v-if="conv.unread_count > 0" :value="conv.unread_count" :max="99" class="badge" />
        </div>
      </div>
    </div>
    <div v-if="chatStore.conversations.length === 0" class="empty-text">暂无会话</div>
  </div>
</template>

<script setup>
import { useChatStore } from '../../stores/chat'
import { useFriendStore } from '../../stores/friend'
import * as groupApi from '../../api/group'

const chatStore = useChatStore()
const friendStore = useFriendStore()

function isActive(conv) {
  const a = chatStore.activeConv
  return a && a.target_type === conv.target_type && a.target_id === conv.target_id
}

async function selectConv(conv) {
  // 填充名称和头像
  const info = { ...conv }
  if (!info.target_name) {
    if (conv.target_type === 1) {
      const friend = friendStore.friendList.find((f) => f.friend && f.friend_id === conv.target_id)
      info.target_name = friend?.friend?.nickname || '用户'
      info.target_avatar = friend?.friend?.avatar || ''
    } else {
      try {
        const res = await groupApi.getGroupInfo(conv.target_id)
        info.target_name = res.data?.name || '群组'
        info.target_avatar = res.data?.avatar || ''
      } catch { info.target_name = '群组' }
    }
  }
  chatStore.clearUnread(conv.target_type, conv.target_id)
  chatStore.setActiveConv(info)
}

function formatTime(t) {
  if (!t) return ''
  const d = new Date(t)
  const now = new Date()
  const pad = (n) => String(n).padStart(2, '0')
  if (d.toDateString() === now.toDateString()) {
    return `${pad(d.getHours())}:${pad(d.getMinutes())}`
  }
  return `${pad(d.getMonth() + 1)}/${pad(d.getDate())}`
}
</script>

<style scoped>
.session-list { padding: 0; }
.session-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  cursor: pointer;
  transition: background 0.15s;
}
.session-item:hover { background: #e8f0fe; }
.session-item.active { background: #d4e4fc; }
.session-info {
  flex: 1;
  min-width: 0;
}
.session-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.session-name { font-size: 14px; font-weight: 500; }
.session-time { font-size: 11px; color: #999; }
.session-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 2px;
}
.session-last { font-size: 12px; color: #888; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 160px; }
.badge { margin-left: 4px; }
.empty-text { text-align: center; color: #999; padding: 40px 0; }
</style>
