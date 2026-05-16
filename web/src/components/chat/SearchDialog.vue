<template>
  <el-dialog v-model="visible" title="搜索消息" width="500px">
    <el-input
      v-model="keyword"
      placeholder="搜索关键词"
      clearable
      @keydown.enter="doSearch"
    >
      <template #append>
        <el-button @click="doSearch">搜索</el-button>
      </template>
    </el-input>

    <div v-if="searched" class="sr-results">
      <div v-if="results.length === 0" class="empty-text">未找到相关消息</div>
      <div v-for="msg in results" :key="msg.id" class="sr-item" @click="goToMsg(msg)">
        <div class="sr-sender">
          <el-avatar :size="24" :src="msg.sender?.avatar" />
          <span>{{ msg.sender?.nickname || '用户' }}</span>
          <span class="sr-time">{{ formatTime(msg.created_at) }}</span>
        </div>
        <div class="sr-content">{{ msg.content_type === 2 ? '[图片]' : msg.content }}</div>
      </div>
      <div v-if="hasMore" class="load-more">
        <el-button text size="small" @click="loadMore">加载更多</el-button>
      </div>
    </div>
  </el-dialog>
</template>

<script setup>
import { ref, watch } from 'vue'
import { searchMessage } from '../../api/message'
import { useChatStore } from '../../stores/chat'

const chatStore = useChatStore()
const visible = defineModel('visible', { type: Boolean, default: false })
const keyword = ref('')
const results = ref([])
const searched = ref(false)
const page = ref(1)
const hasMore = ref(false)
const total = ref(0)

async function doSearch() {
  const kw = keyword.value.trim()
  if (!kw) return
  page.value = 1
  await fetchResults()
}

async function fetchResults() {
  try {
    const res = await searchMessage({ keyword: keyword.value.trim(), page: page.value, page_size: 20 })
    const data = res.data || {}
    const list = data.list || []
    if (page.value === 1) {
      results.value = list
    } else {
      results.value = [...results.value, ...list]
    }
    total.value = data.total || 0
    hasMore.value = results.value.length < total.value
    searched.value = true
  } catch {
    results.value = []
  }
}

async function loadMore() {
  page.value++
  await fetchResults()
}

function goToMsg(msg) {
  // 切换到对应会话
  chatStore.setActiveConv({
    target_type: msg.target_type,
    target_id: msg.target_type === 1 ? (msg.sender_id === chatStore.userId ? msg.target_id : msg.sender_id) : msg.target_id,
    target_name: msg.sender?.nickname || '聊天',
    target_avatar: msg.sender?.avatar || '',
  })
  visible.value = false
}

function formatTime(t) {
  if (!t) return ''
  return new Date(t).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
.sr-results { margin-top: 12px; max-height: 400px; overflow-y: auto; }
.sr-item { padding: 10px; border-radius: 6px; cursor: pointer; }
.sr-item:hover { background: #e8f0fe; }
.sr-sender { display: flex; align-items: center; gap: 6px; font-size: 13px; }
.sr-time { margin-left: auto; color: #999; font-size: 11px; }
.sr-content { margin-top: 4px; font-size: 14px; color: #333; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.empty-text { text-align: center; color: #999; padding: 40px 0; font-size: 13px; }
.load-more { text-align: center; padding: 8px; }
</style>
