import request from './request'

// 管理员登录
export const login = (credentials) => {
  return request.post('/api/auth/login', credentials)
}

// 管理员登出
export const logout = () => {
  return request.post('/api/auth/logout')
}

// 获取当前管理员信息
export const getCurrentUser = () => {
  return request.get('/api/auth/validate')
}

// 刷新令牌
export const refreshToken = () => {
  return request.post('/api/auth/refresh')
}
