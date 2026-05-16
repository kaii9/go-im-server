<template>
  <div class="group-view">
    <div class="gv-header">
      <span class="gv-title">我的群组（{{ friendStore.myGroups.length }}）</span>
      <el-button size="small" type="primary" @click="showCreate = true">创建群组</el-button>
    </div>

    <div class="gv-content">
      <div v-if="friendStore.myGroups.length === 0" class="empty-text">暂无群组</div>
      <div
        v-for="g in friendStore.myGroups"
        :key="g.id"
        class="gv-item"
        :class="{ active: isActiveConv(g) }"
        @click="selectGroup(g)"
      >
        <el-avatar :size="40" :src="g.group?.avatar || g.avatar" />
        <div class="gv-info">
          <span class="gv-name">{{ g.group?.name || g.name }}</span>
          <span class="gv-meta">{{ g.group?.member_count || g.member_count }} 人</span>
        </div>
        <div class="gv-actions">
          <el-button text size="small" @click.stop="showInvite(g)">邀请</el-button>
          <el-button v-if="g.role !== 1" text size="small" type="danger" @click.stop="handleLeave(g.group_id)">退出</el-button>
          <el-tag v-if="g.role === 1" size="small">群主</el-tag>
        </div>
      </div>
    </div>

    <!-- 创建群组弹窗 -->
    <el-dialog v-model="showCreate" title="创建群组" width="360px">
      <el-form :model="createForm" label-width="60px">
        <el-form-item label="群名称">
          <el-input v-model="createForm.name" placeholder="输入群组名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" @click="confirmCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- 邀请好友弹窗 -->
    <el-dialog v-model="showInviteDialog" title="邀请好友" width="360px">
      <div v-if="friendStore.friendList.length === 0" class="empty-text">暂无好友可邀请</div>
      <div v-for="f in friendStore.friendList" :key="f.id" class="fv-item" @click="confirmInvite(f)">
        <el-avatar :size="32" :src="f.friend?.avatar" />
        <span>{{ f.friend?.nickname || '用户' }}</span>
      </div>
      <template #footer>
        <el-button @click="showInviteDialog = false">取消</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useFriendStore } from '../../stores/friend'
import { useChatStore } from '../../stores/chat'

const friendStore = useFriendStore()
const chatStore = useChatStore()

// 创建群组
const showCreate = ref(false)
const createForm = reactive({ name: '' })

async function confirmCreate() {
  const name = createForm.name.trim()
  if (!name) { ElMessage.warning('请输入群组名称'); return }
  try {
    await friendStore.createGroup(name)
    ElMessage.success('群组创建成功')
    showCreate.value = false
    createForm.name = ''
  } catch {
    // handled
  }
}

// 邀请好友
const showInviteDialog = ref(false)
const inviteGroup = ref(null)

function showInvite(g) {
  inviteGroup.value = g
  showInviteDialog.value = true
}

async function confirmInvite(f) {
  try {
    await friendStore.inviteGroupMember(inviteGroup.value.group_id, f.friend_id)
    ElMessage.success('邀请成功')
  } catch {
    // handled
  }
}

// 退出群组
async function handleLeave(groupId) {
  try {
    await ElMessageBox.confirm('确定退出该群组？', '提示', { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' })
    await friendStore.leaveGroup(groupId)
    ElMessage.success('已退出群组')
  } catch {
    // cancelled or error
  }
}

// 选中群组会话
function isActiveConv(g) {
  const a = chatStore.activeConv
  return a && a.target_type === 2 && a.target_id === g.group_id
}

function selectGroup(g) {
  chatStore.clearUnread(2, g.group_id)
  chatStore.setActiveConv({
    target_type: 2,
    target_id: g.group_id,
    target_name: g.group?.name || g.name,
    target_avatar: g.group?.avatar || g.avatar || '',
  })
}
</script>

<style scoped>
.group-view { height: 100%; display: flex; flex-direction: column; }
.gv-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 8px 16px; border-bottom: 1px solid #e0e0e0;
}
.gv-title { font-size: 14px; font-weight: 500; }
.gv-content { flex: 1; overflow-y: auto; }
.gv-item {
  display: flex; align-items: center; gap: 10px;
  padding: 12px 16px; cursor: pointer; transition: background 0.15s;
}
.gv-item:hover { background: #e8f0fe; }
.gv-item.active { background: #d4e4fc; }
.gv-info { flex: 1; min-width: 0; display: flex; flex-direction: column; }
.gv-name { font-size: 14px; font-weight: 500; }
.gv-meta { font-size: 12px; color: #888; }
.gv-actions { flex-shrink: 0; }
.empty-text { text-align: center; color: #999; padding: 40px 0; font-size: 13px; }
.fv-item { display: flex; align-items: center; gap: 10px; padding: 8px 4px; cursor: pointer; border-radius: 4px; }
.fv-item:hover { background: #e8f0fe; }
</style>
