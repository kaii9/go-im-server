<template>
  <el-dialog v-model="visible" title="转发消息" width="400px">
    <p class="fd-preview">{{ previewText }}</p>
    <el-divider />
    <el-input v-model="searchKeyword" placeholder="搜索会话" size="small" clearable @input="filterTargets" />
    <div class="fd-list">
      <div
        v-for="t in filteredTargets"
        :key="t.key"
        class="fd-item"
        @click="forwardTo(t)"
      >
        <el-avatar :size="32" :src="t.avatar" />
        <div class="fd-info">
          <span class="fd-name">{{ t.name }}</span>
          <span class="fd-type">{{ t.target_type === 1 ? '好友' : '群组' }}</span>
        </div>
      </div>
      <div v-if="filteredTargets.length === 0" class="empty-text">无可转发对象</div>
    </div>
  </el-dialog>
</template>

<script setup>
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useChatStore } from '../../stores/chat'
import { useFriendStore } from '../../stores/friend'

const chatStore = useChatStore()
const friendStore = useFriendStore()

const visible = defineModel('visible', { type: Boolean, default: false })
const props = defineProps({ message: { type: Object, default: null } })
const searchKeyword = ref('')

const previewText = computed(() => {
  if (!props.message) return ''
  return props.message.content_type === 2 ? '[图片]' : props.message.content
})

const targets = computed(() => {
  const list = []
  // 好友
  for (const f of friendStore.friendList) {
    list.push({
      key: `f_${f.friend_id}`,
      target_type: 1,
      target_id: f.friend_id,
      name: f.friend?.nickname || '用户',
      avatar: f.friend?.avatar || '',
    })
  }
  // 群组
  for (const g of friendStore.myGroups) {
    list.push({
      key: `g_${g.group_id}`,
      target_type: 2,
      target_id: g.group_id,
      name: g.group?.name || g.name,
      avatar: g.group?.avatar || g.avatar || '',
    })
  }
  return list
})

const filteredTargets = computed(() => {
  const kw = searchKeyword.value.toLowerCase()
  if (!kw) return targets.value
  return targets.value.filter((t) => t.name.toLowerCase().includes(kw))
})

function filterTargets() {
  // reactivity handles it
}

function forwardTo(target) {
  if (!props.message) return
  chatStore.sendMessageTo(target.target_type, target.target_id, props.message.content, props.message.content_type)
  visible.value = false
  ElMessage.success('已转发')
}
</script>

<style scoped>
.fd-preview { font-size: 14px; color: #555; padding: 8px; background: #f5f5f5; border-radius: 4px; }
.fd-list { max-height: 300px; overflow-y: auto; margin-top: 8px; }
.fd-item { display: flex; align-items: center; gap: 10px; padding: 10px 8px; border-radius: 6px; cursor: pointer; }
.fd-item:hover { background: #e8f0fe; }
.fd-info { flex: 1; min-width: 0; display: flex; flex-direction: column; }
.fd-name { font-size: 14px; font-weight: 500; }
.fd-type { font-size: 11px; color: #999; }
.empty-text { text-align: center; color: #999; padding: 40px 0; font-size: 13px; }
</style>
