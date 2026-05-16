<template>
  <div class="chat-area">
    <div class="chat-header">
      <span class="chat-title">{{ chatStore.activeConv?.target_name || '聊天' }}</span>
      <div>
        <el-button text size="small" @click="showSearch">搜索</el-button>
        <el-button text size="small" @click="showInfo">详情</el-button>
      </div>
    </div>
    <MessageList @forward="openForward" />
    <ChatInput />

    <SearchDialog v-model:visible="searchVisible" />
    <ForwardDialog v-model:visible="forwardVisible" :message="forwardMsg" />
  </div>
</template>

<script setup>
import { ref, inject } from 'vue'
import MessageList from './MessageList.vue'
import ChatInput from './ChatInput.vue'
import ChatDetail from './ChatDetail.vue'
import SearchDialog from './SearchDialog.vue'
import ForwardDialog from './ForwardDialog.vue'
import { useChatStore } from '../../stores/chat'

const chatStore = useChatStore()
const openRightPanel = inject('openRightPanel')

function showInfo() {
  openRightPanel(ChatDetail)
}

const searchVisible = ref(false)
function showSearch() {
  searchVisible.value = true
}

const forwardVisible = ref(false)
const forwardMsg = ref(null)
function openForward(msg) {
  forwardMsg.value = msg
  forwardVisible.value = true
}
</script>

<style scoped>
.chat-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100%;
}
.chat-header {
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  border-bottom: 1px solid #e0e0e0;
  background: #fff;
  flex-shrink: 0;
}
.chat-title {
  font-size: 16px;
  font-weight: 600;
}
</style>
