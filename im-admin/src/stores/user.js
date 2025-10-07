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
        const userInfo = await getCurrentUser()
        user.value = userInfo
      } catch (error) {
        console.error('获取用户信息失败:', error)
        logout()
      }
    }
  }
  
  const loginUser = async (credentials) => {
    try {
      const response = await login(credentials)
      token.value = response.token
      user.value = response.user
      localStorage.setItem('admin_token', response.token)
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
