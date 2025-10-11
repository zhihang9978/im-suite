import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '@/api/client'

export const useMessageStore = defineStore('message', () => {
  const unreadCount = ref(0)
  
  // 获取未读消息数
  const getUnreadCount = async () => {
    try {
      const response = await api.get('/api/messages/unread/count')
      if (response.data.success) {
        unreadCount.value = response.data.data.count || 0
      }
    } catch (error) {
      console.error('获取未读消息数失败:', error)
    }
  }
  
  // 标记消息为已读
  const markAsRead = async (messageId) => {
    try {
      const response = await api.post(`/api/messages/${messageId}/read`)
      if (response.data.success) {
        getUnreadCount() // 刷新未读数
        return { success: true }
      }
      return { success: false }
    } catch (error) {
      console.error('标记已读失败:', error)
      return { success: false }
    }
  }
  
  // 撤回消息
  const recallMessage = async (messageId, reason = '') => {
    try {
      const response = await api.post(`/api/messages/${messageId}/recall`, { reason })
      if (response.data.success) {
        return { success: true }
      }
      return { success: false, message: response.data.error }
    } catch (error) {
      return { success: false, message: error.message }
    }
  }
  
  // 删除消息
  const deleteMessage = async (messageId) => {
    try {
      const response = await api.delete(`/api/messages/${messageId}`)
      if (response.data.success) {
        return { success: true }
      }
      return { success: false }
    } catch (error) {
      return { success: false }
    }
  }
  
  return {
    unreadCount,
    getUnreadCount,
    markAsRead,
    recallMessage,
    deleteMessage,
  }
})

