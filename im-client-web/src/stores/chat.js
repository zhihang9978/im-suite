import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/api/client'
import { connectWebSocket, sendMessage as wsSendMessage } from '@/utils/websocket'

export const useChatStore = defineStore('chat', () => {
  const messages = ref([])
  const contacts = ref([])
  const currentChat = ref(null)
  const ws = ref(null)

  // 初始化WebSocket
  const initWebSocket = () => {
    const token = localStorage.getItem('token')
    if (token) {
      ws.value = connectWebSocket(token, handleWSMessage)
    }
  }

  // 处理WebSocket消息
  const handleWSMessage = (message) => {
    if (message.type === 'message') {
      messages.value.push(message.data)
    } else if (message.type === 'typing') {
      // 处理打字提示
    }
  }

  // 发送消息
  const sendMessage = async (content, receiverId) => {
    try {
      const response = await api.post('/api/messages/send', {
        receiver_id: receiverId,
        content: content,
        message_type: 'text'
      })
      
      if (response.data.success) {
        messages.value.push(response.data.data)
        return { success: true }
      }
      return { success: false, message: response.data.error }
    } catch (error) {
      return { success: false, message: error.message }
    }
  }

  // 获取消息列表
  const getMessages = async (receiverId) => {
    try {
      const response = await api.get(`/api/messages?receiver_id=${receiverId}&limit=50`)
      if (response.data.success) {
        messages.value = response.data.data || []
        return { success: true }
      }
      return { success: false }
    } catch (error) {
      console.error('获取消息失败:', error)
      return { success: false }
    }
  }

  // 获取联系人列表
  const getContacts = async () => {
    try {
      const response = await api.get('/api/users/friends')
      if (response.data.success) {
        contacts.value = response.data.data || []
        return { success: true }
      }
      return { success: false }
    } catch (error) {
      console.error('获取联系人失败:', error)
      return { success: false }
    }
  }

  // 切换当前聊天
  const switchChat = (contact) => {
    currentChat.value = contact
    getMessages(contact.id)
  }

  return {
    messages,
    contacts,
    currentChat,
    initWebSocket,
    sendMessage,
    getMessages,
    getContacts,
    switchChat,
  }
})

