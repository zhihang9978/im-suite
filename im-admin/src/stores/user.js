import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login, logout, getCurrentUser } from '@/api/auth'

export const useUserStore = defineStore('user', () => {
  const user = ref(null)
  const token = ref(localStorage.getItem('admin_token') || '')
  
  const isLoggedIn = computed(() => !!token.value)
  
  const initUser = async () => {
    if (token.value) {
      try {
        const response = await getCurrentUser()
        // 后端返回的数据结构可能是 { user: {...} } 或直接是用户对象
        user.value = response.user || response
      } catch (error) {
        console.error('获取用户信息失败:', error)
        logoutUser()
      }
    }
  }
  
  const loginUser = async (credentials) => {
    try {
      const response = await login(credentials)
      // 后端返回的是 access_token 和 refresh_token，不是 token
      const accessToken = response.access_token || response.token
      token.value = accessToken
      user.value = response.user
      localStorage.setItem('admin_token', accessToken)
      // 也保存 refresh_token
      if (response.refresh_token) {
        localStorage.setItem('refresh_token', response.refresh_token)
      }
      return response
    } catch (error) {
      throw error
    }
  }
  
  const logoutUser = async () => {
    try {
      await logout()
    } catch (error) {
      console.error('登出失败:', error)
    } finally {
      token.value = ''
      user.value = null
      localStorage.removeItem('admin_token')
    }
  }
  
  return {
    user,
    token,
    isLoggedIn,
    initUser,
    loginUser,
    logoutUser
  }
})
