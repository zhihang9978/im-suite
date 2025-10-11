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
        // 后端返回的数据结构可能是 { user: {...} } 或直接是用户对象 (important-comment)
        user.value = response.user || response
      } catch (error) {
        // 静默处理过期token，不打印错误日志
        // Silently handle expired tokens without console errors
        // 只清除token，不需要调用logout API
        token.value = ''
        user.value = null
        localStorage.removeItem('admin_token')
        localStorage.removeItem('refresh_token')
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
    // 先清除本地状态
    token.value = ''
    user.value = null
    localStorage.removeItem('admin_token')
    localStorage.removeItem('refresh_token')
    
    // 尝试调用后端logout API，但不在意成功与否
    try {
      await logout()
    } catch (error) {
      // 静默处理logout错误，因为本地状态已经清除
      // Silently handle logout errors since local state is already cleared
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
