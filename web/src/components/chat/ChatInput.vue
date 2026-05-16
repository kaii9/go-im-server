<template>
  <div class="chat-input">
    <div class="toolbar">
      <el-upload
        :show-file-list="false"
        :http-request="handleUpload"
        accept="image/jpeg,image/png,image/gif"
      >
        <el-button text size="small">
          <el-icon><Picture /></el-icon> 图片
        </el-button>
      </el-upload>
    </div>
    <el-input
      v-model="text"
      type="textarea"
      :rows="3"
      placeholder="输入消息，Enter 发送"
      @keydown.enter.prevent="send"
      resize="none"
    />
    <div class="send-bar">
      <el-button type="primary" size="small" @click="send">发送</el-button>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useChatStore } from '../../stores/chat'
import { uploadImage } from '../../api/message'

const chatStore = useChatStore()
const text = ref('')

function send() {
  const content = text.value.trim()
  if (!content) return
  chatStore.sendMessage(content, 1)
  text.value = ''
}

async function handleUpload(options) {
  try {
    const res = await uploadImage(options.file)
    chatStore.sendMessage(res.data.url, 2)
  } catch {
    ElMessage.error('图片上传失败')
  }
}
</script>

<style scoped>
.chat-input {
  border-top: 1px solid #e0e0e0;
  padding: 8px 16px;
  background: #fff;
  flex-shrink: 0;
}
.toolbar {
  margin-bottom: 4px;
}
.send-bar {
  display: flex;
  justify-content: flex-end;
  margin-top: 6px;
}
</style>
