import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import router from '@/router'

// 创建 axios 实例
const request = axios.create({
  baseURL: '/api',
  timeout: 10000
})

// 请求拦截器
request.interceptors.request.use(
  config => {
    const userStore = useUserStore()
    if (userStore.token) {
      config.headers.Authorization = `Bearer ${userStore.token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    const { response, request, message } = error
    
    // 处理HTTP响应错误
    if (response) {
      // 服务器返回了错误状态码
      switch (response.status) {
        case 401:
          // 不显示错误消息和不调用logout，让调用方处理
          // Don't show error message or call logout, let the caller handle it
          // 这避免了与stores/user.js中的错误处理重复
          // This avoids duplication with error handling in stores/user.js
          break
        case 403:
          ElMessage.error('没有权限访问该资源')
          break
        case 404:
          ElMessage.error('请求的资源不存在')
          break
        case 429:
          ElMessage.error('请求过于频繁，请稍后再试')
          break
        case 500:
          ElMessage.error('服务器内部错误，请稍后重试')
          console.error('Server error:', response.data)
          break
        case 502:
        case 503:
        case 504:
          ElMessage.error('服务暂时不可用，请稍后重试')
          break
        default:
          ElMessage.error(response.data?.error || response.data?.message || '请求失败')
      }
    } else if (request) {
      // 请求已发送但没有收到响应
      ElMessage.error('服务器无响应，请检查网络连接')
      console.error('No response received:', request)
    } else {
      // 请求配置错误
      ElMessage.error(`请求配置错误: ${message}`)
      console.error('Request setup error:', message)
    }
    
    return Promise.reject(error)
  }
)

export default request
