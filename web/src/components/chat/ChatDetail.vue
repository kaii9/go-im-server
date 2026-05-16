<template>
  <div class="chat-detail">
    <div class="cd-header">
      <span>详情</span>
      <el-button text size="small" @click="$emit('close')">关闭</el-button>
    </div>
    <div class="cd-body">
      <div class="cd-profile">
        <el-avatar :size="64" :src="activeConv?.target_avatar" />
        <span class="cd-name">{{ activeConv?.target_name }}</span>
        <el-tag :type="activeConv?.target_type === 1 ? '' : 'success'" size="small">
          {{ activeConv?.target_type === 1 ? '好友' : '群组' }}
        </el-tag>
      </div>

      <template v-if="activeConv?.target_type === 2">
        <el-divider />
        <div class="cd-section-title">群成员（{{ members.length }}）</div>
        <div v-for="m in members" :key="m.id" class="cd-member">
          <el-avatar :size="32" :src="m.user?.avatar" />
          <span>{{ m.user?.nickname || '用户' }}</span>
          <el-tag v-if="m.role === 1" size="small" type="warning">群主</el-tag>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useChatStore } from '../../stores/chat'
import { getGroupMembers } from '../../api/group'

const chatStore = useChatStore()
const activeConv = chatStore.activeConv
const members = ref([])

watch(() => activeConv.value, async (conv) => {
  members.value = []
  if (conv?.target_type === 2) {
    try {
      const res = await getGroupMembers(conv.target_id)
      members.value = res.data || []
    } catch { /* ignore */ }
  }
}, { immediate: true })
</script>

<style scoped>
.chat-detail { padding: 16px; }
.cd-header {
  display: flex; justify-content: space-between; align-items: center;
  padding-bottom: 16px; border-bottom: 1px solid #e0e0e0;
  font-weight: 600;
}
.cd-body { display: flex; flex-direction: column; gap: 12px; }
.cd-profile {
  display: flex; flex-direction: column; align-items: center; gap: 8px; padding-top: 16px;
}
.cd-name { font-size: 16px; font-weight: 500; }
.cd-section-title { font-size: 13px; color: #666; font-weight: 500; }
.cd-member {
  display: flex; align-items: center; gap: 10px; padding: 6px 0;
}
</style>
