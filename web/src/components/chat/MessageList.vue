<template>
  <div ref="listRef" class="msg-list" @scroll="onScroll">
    <div v-if="chatStore.messages.length < chatStore.totalMsg" class="load-more">
      <el-button text size="small" @click="chatStore.loadMore()">加载更多</el-button>
    </div>
    <div v-for="msg in chatStore.messages" :key="msg.id" class="msg-item" :class="{ self: msg.sender_id === uid }">
      <div class="msg-avatar">
        <el-avatar :size="36" :src="msg.sender?.avatar" />
      </div>
      <div class="msg-body">
        <div class="msg-content">
          <img v-if="msg.content_type === 2" :src="msg.content" class="msg-img" @click="preview(msg.content)" />
          <span v-else>{{ msg.content }}</span>
        </div>
        <div class="msg-meta">
          <span class="msg-time">{{ formatT(msg.created_at) }}</span>
          <el-button text size="small" class="msg-forward-btn" @click="$emit('forward', msg)">转发</el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick, watch } from 'vue'
import { useChatStore } from '../../stores/chat'
import { useUserStore } from '../../stores/user'

const emit = defineEmits(['forward'])
const chatStore = useChatStore()
const userStore = useUserStore()
const uid = userStore.userInfo?.id
const listRef = ref(null)

onMounted(() => scrollToBottom())

watch(() => chatStore.messages.length, () => nextTick(scrollToBottom))

function scrollToBottom() {
  if (listRef.value) {
    listRef.value.scrollTop = listRef.value.scrollHeight
  }
}

function onScroll() {
  const el = listRef.value
  if (el && el.scrollTop === 0) {
    chatStore.loadMore()
  }
}

function preview(url) {
  // Element Plus image viewer
  const div = document.createElement('div')
  document.body.appendChild(div)
  // 简单实现：在新窗口打开
  window.open(url, '_blank')
}

function formatT(t) {
  if (!t) return ''
  const d = new Date(t)
  return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}
</script>

<style scoped>
.msg-list {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}
.load-more {
  text-align: center;
  padding: 8px;
}
.msg-item {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
}
.msg-item.self {
  flex-direction: row-reverse;
}
.msg-body {
  max-width: 60%;
}
.msg-content {
  background: #fff;
  padding: 8px 14px;
  border-radius: 8px;
  font-size: 14px;
  line-height: 1.5;
  word-break: break-word;
}
.self .msg-content {
  background: #95ec69;
}
.msg-img {
  max-width: 200px;
  max-height: 200px;
  border-radius: 4px;
  cursor: pointer;
}
.msg-time {
  font-size: 11px;
  color: #999;
  margin-top: 4px;
}
.self .msg-time {
  text-align: right;
}
.msg-meta {
  display: flex; align-items: center; gap: 8px; margin-top: 4px;
}
.msg-forward-btn {
  opacity: 0; transition: opacity 0.15s; font-size: 11px;
}
.msg-item:hover .msg-forward-btn {
  opacity: 1;
}
</style>
