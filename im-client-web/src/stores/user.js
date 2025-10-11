import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '@/api/client'

export const useUserStore = defineStore('user', () => {
  const user = ref(null)
  const token = ref(localStorage.getItem('token') || '')
  const isAuthenticated = ref(!!token.value)

  // 登录
  const login = async (phone, password) => {
    try {
      const response = await api.post('/api/auth/login', { phone, password })
      if (response.data.success) {
        token.value = response.data.data.token
        user.value = response.data.data.user
        localStorage.setItem('token', token.value)
        isAuthenticated.value = true
        return { success: true }
      }
      return { success: false, message: response.data.error }
    } catch (error) {
      return { success: false, message: error.message }
    }
  }

  // 注册
  const register = async (phone, password, nickname) => {
    try {
      const response = await api.post('/api/auth/register', { phone, password, nickname })
      if (response.data.success) {
        token.value = response.data.data.token
        user.value = response.data.data.user
        localStorage.setItem('token', token.value)
        isAuthenticated.value = true
        return { success: true }
      }
      return { success: false, message: response.data.error }
    } catch (error) {
      return { success: false, message: error.message }
    }
  }

  // 登出
  const logout = async () => {
    try {
      await api.post('/api/auth/logout')
    } catch (error) {
      console.error('登出失败:', error)
    } finally {
      token.value = ''
      user.value = null
      isAuthenticated.value = false
      localStorage.removeItem('token')
    }
  }

  // 检查认证状态
  const checkAuth = async () => {
    try {
      const response = await api.get('/api/users/me')
      if (response.data.success) {
        user.value = response.data.data
        isAuthenticated.value = true
      } else {
        logout()
      }
    } catch (error) {
      logout()
    }
  }

  return {
    user,
    token,
    isAuthenticated,
    login,
    register,
    logout,
    checkAuth,
  }
})

