<template>
  <div class="friend-view">
    <el-tabs v-model="tab" class="ftabs">
      <el-tab-pane label="好友" name="list" />
      <el-tab-pane label="搜索" name="search" />
      <el-tab-pane label="申请" name="app" />
    </el-tabs>

    <div class="fv-content">
      <!-- 好友列表 -->
      <template v-if="tab === 'list'">
        <div v-if="friendStore.friendList.length === 0" class="empty-text">暂无好友</div>
        <div
          v-for="f in friendStore.friendList"
          :key="f.id"
          class="fv-item"
          style="cursor:pointer"
          @click="startChat(f)"
        >
          <el-avatar :size="36" :src="f.friend?.avatar" />
          <div class="fv-info">
            <span class="fv-name">{{ f.friend?.nickname || '用户' }}</span>
            <span class="fv-remark">{{ f.remark || '' }}</span>
          </div>
          <el-button text size="small" type="danger" @click.stop="handleDelete(f.friend_id)">删除</el-button>
        </div>
        <!-- 已发送的申请 -->
        <el-divider v-if="friendStore.sentApps.length" content-position="left">已发送的申请</el-divider>
        <div
          v-for="app in friendStore.sentApps"
          :key="app.id"
          class="fv-item fv-app-item"
        >
          <el-avatar :size="36" />
          <div class="fv-info">
            <span class="fv-name">用户 {{ app.to_user_id }}</span>
            <el-tag :type="app.status === 0 ? 'warning' : app.status === 1 ? 'success' : 'danger'" size="small">
              {{ app.status === 0 ? '待处理' : app.status === 1 ? '已通过' : '已拒绝' }}
            </el-tag>
          </div>
          <span class="fv-reason">备注：{{ app.reason }}</span>
        </div>
      </template>

      <!-- 搜索用户 -->
      <template v-if="tab === 'search'">
        <div class="search-box">
          <el-input v-model="keyword" placeholder="搜索用户名" size="small" clearable @keydown.enter="doSearch" />
          <el-button size="small" type="primary" @click="doSearch">搜索</el-button>
        </div>
        <div v-if="searchResult.length" class="search-results">
          <div v-for="u in searchResult" :key="u.id" class="fv-item">
            <el-avatar :size="36" :src="u.avatar" />
            <div class="fv-info">
              <span class="fv-name">{{ u.nickname || u.username }}</span>
              <span class="fv-remark">@{{ u.username }}</span>
            </div>
            <el-button size="small" @click="showApplyDialog(u)">加好友</el-button>
          </div>
        </div>
        <div v-if="searched && !searchResult.length" class="empty-text">未找到用户</div>
      </template>

      <!-- 好友申请 -->
      <template v-if="tab === 'app'">
        <div v-if="friendStore.receivedApps.length === 0" class="empty-text">暂无申请</div>
        <div
          v-for="app in friendStore.receivedApps"
          :key="app.id"
          class="fv-item fv-app-item"
        >
          <el-avatar :size="36" />
          <div class="fv-info">
            <span class="fv-name">用户 {{ app.from_user_id }}</span>
            <span class="fv-reason">{{ app.reason || '无备注' }}</span>
          </div>
          <div v-if="app.status === 0" class="app-actions">
            <el-button size="small" type="primary" @click="handleApp(app.id, true)">同意</el-button>
            <el-button size="small" @click="handleApp(app.id, false)">拒绝</el-button>
          </div>
          <el-tag v-else :type="app.status === 1 ? 'success' : 'danger'" size="small">
            {{ app.status === 1 ? '已同意' : '已拒绝' }}
          </el-tag>
        </div>
      </template>
    </div>

    <!-- 加好友弹窗 -->
    <el-dialog v-model="applyDialog" title="添加好友" width="360px">
      <p>发送好友申请给 <strong>{{ applyTarget?.nickname || applyTarget?.username }}</strong></p>
      <el-input v-model="applyReason" placeholder="添加备注（可选）" />
      <template #footer>
        <el-button @click="applyDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmApply">发送</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useFriendStore } from '../../stores/friend'
import { useChatStore } from '../../stores/chat'
import { searchUser } from '../../api/user'

const friendStore = useFriendStore()
const chatStore = useChatStore()

const tab = ref('list')
const keyword = ref('')
const searchResult = ref([])
const searched = ref(false)

// 搜索用户
async function doSearch() {
  const kw = keyword.value.trim()
  if (!kw) { searchResult.value = []; searched.value = false; return }
  try {
    const res = await searchUser(kw)
    searchResult.value = res.data || []
  } catch {
    searchResult.value = []
  }
  searched.value = true
}

// 申请好友
const applyDialog = ref(false)
const applyTarget = ref(null)
const applyReason = ref('')

function showApplyDialog(user) {
  applyTarget.value = user
  applyReason.value = ''
  applyDialog.value = true
}

async function confirmApply() {
  try {
    await friendStore.applyFriend(applyTarget.value.id, applyReason.value)
    ElMessage.success('申请已发送')
    applyDialog.value = false
  } catch {
    // handled by interceptor
  }
}

// 处理申请
async function handleApp(appId, agree) {
  try {
    await friendStore.handleApplication(appId, agree)
    ElMessage.success(agree ? '已同意' : '已拒绝')
  } catch {
    // handled
  }
}

// 删除好友
async function handleDelete(friendId) {
  try {
    await ElMessageBox.confirm('确定删除该好友？', '提示', { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' })
    await friendStore.deleteFriend(friendId)
    ElMessage.success('已删除')
  } catch {
    // cancelled or error
  }
}

// 开始聊天
function startChat(f) {
  chatStore.setActiveConv({
    target_type: 1,
    target_id: f.friend_id,
    target_name: f.friend?.nickname || '用户',
    target_avatar: f.friend?.avatar || '',
  })
}
</script>

<style scoped>
.friend-view { height: 100%; display: flex; flex-direction: column; }
.ftabs { padding: 0 12px; }
.fv-content { flex: 1; overflow-y: auto; padding: 0 12px 12px; }
.fv-item {
  display: flex; align-items: center; gap: 10px;
  padding: 10px 8px; border-radius: 6px; cursor: default;
}
.fv-item:hover { background: #e8f0fe; }
.fv-info { flex: 1; min-width: 0; display: flex; flex-direction: column; }
.fv-name { font-size: 14px; font-weight: 500; }
.fv-remark { font-size: 12px; color: #888; }
.fv-reason { font-size: 12px; color: #555; }
.fv-app-item { flex-wrap: wrap; }
.app-actions { display: flex; gap: 4px; }
.empty-text { text-align: center; color: #999; padding: 40px 0; font-size: 13px; }
.search-box { display: flex; gap: 8px; margin-bottom: 12px; }
.search-results { display: flex; flex-direction: column; }
</style>
