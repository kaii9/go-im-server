import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as msgApi from '../api/message'
import { send as wsSend } from '../ws'
import { useUserStore } from './user'

export const useChatStore = defineStore('chat', () => {
  const userStore = useUserStore()
  const conversations = ref([])
  const activeConv = ref(null) // { target_type, target_id, target_name, target_avatar }
  const messages = ref([])
  const totalMsg = ref(0)
  const currentPage = ref(1)

  async function fetchConversations() {
    const res = await msgApi.getConversations()
    conversations.value = res.data || []
  }

  function setActiveConv(conv) {
    activeConv.value = conv
    messages.value = []
    totalMsg.value = 0
    currentPage.value = 1
    if (conv) fetchHistory(1)
  }

  async function fetchHistory(page) {
    if (!activeConv.value) return
    const p = page || currentPage.value
    const res = await msgApi.getHistory({
      target_type: activeConv.value.target_type,
      target_id: activeConv.value.target_id,
      page: p,
      page_size: 20,
    })
    if (p === 1) {
      messages.value = res.data.list || []
    } else {
      messages.value = [...(res.data.list || []), ...messages.value]
    }
    totalMsg.value = res.data.total || 0
    currentPage.value = p
  }

  function loadMore() {
    if (messages.value.length < totalMsg.value) {
      fetchHistory(currentPage.value + 1)
    }
  }

  function sendMessage(content, contentType) {
    if (!activeConv.value || !content) return
    wsSend({
      type: activeConv.value.target_type === 1 ? 1 : 2,
      to: activeConv.value.target_id,
      target_type: activeConv.value.target_type,
      content_type: contentType || 1,
      content: content,
    })
  }

  function sendMessageTo(targetType, targetId, content, contentType) {
    if (!content) return
    wsSend({
      type: targetType === 1 ? 1 : 2,
      to: targetId,
      target_type: targetType,
      content_type: contentType || 1,
      content: content,
    })
  }

  // 被 WS 消息推送到时调用
  function appendMessage(msg) {
    if (!activeConv.value || msg.target_type !== activeConv.value.target_type) return

    // 计算对方 ID：群聊用 msg.to，单聊需判断 from/to 哪个是"对方"
    let peerId
    if (msg.target_type === 1) {
      peerId = msg.from === userStore.userInfo?.id ? msg.to : msg.from
    } else {
      peerId = msg.to
    }

    if (peerId === activeConv.value.target_id) {
      const senderAvatar = msg.from === userStore.userInfo?.id ? userStore.userInfo?.avatar : ''
      messages.value.push({
        id: Date.now(),
        sender_id: msg.from,
        sender: { avatar: senderAvatar },
        target_type: msg.target_type,
        target_id: msg.target_type === 1 ? msg.from : msg.to,
        content_type: msg.content_type,
        content: msg.content,
        created_at: new Date(msg.timestamp * 1000).toISOString(),
      })
    }

    // 更新会话列表中的最后消息
    const conv = conversations.value.find(
      (c) => c.target_type === msg.target_type && c.target_id === peerId,
    )
    if (conv) {
      conv.last_message = msg.content_type === 2 ? '[图片]' : msg.content
      conv.unread_count = (conv.unread_count || 0) + 1
    }
  }

  function clearUnread(targetType, targetId) {
    const conv = conversations.value.find(
      (c) => c.target_type === targetType && c.target_id === targetId,
    )
    if (conv) conv.unread_count = 0
  }

  return {
    conversations,
    activeConv,
    messages,
    totalMsg,
    currentPage,
    fetchConversations,
    setActiveConv,
    fetchHistory,
    loadMore,
    sendMessage,
    sendMessageTo,
    appendMessage,
    clearUnread,
  }
})
