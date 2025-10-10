import request from './request'

// 管理员登录
export const login = (credentials) => {
  return request.post('/auth/login', credentials)
}

// 管理员登出
export const logout = () => {
  return request.post('/auth/logout')
}

// 获取当前管理员信息
export const getCurrentUser = () => {
  return request.get('/auth/validate')
}

// 刷新令牌
export const refreshToken = () => {
  return request.post('/auth/refresh')
}
