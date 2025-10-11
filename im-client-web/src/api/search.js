import api from './client'

// 搜索用户（通过手机号）
export const searchUserByPhone = async (phone) => {
  try {
    // 由于后端可能没有专门的搜索接口，我们可以尝试注册来检查用户是否存在
    // 或者使用一个模拟的搜索逻辑
    
    // 方案1: 如果后端有搜索接口
    // const response = await api.get(`/api/users/search?phone=${phone}`)
    
    // 方案2: 简化方案 - 直接尝试获取用户信息（需要后端支持）
    // 这里我们假设后端有一个公开的用户查询接口
    const response = await api.get(`/api/users/by-phone/${phone}`)
    
    if (response.data.success) {
      return { success: true, data: response.data.data }
    }
    return { success: false, message: '用户不存在' }
  } catch (error) {
    // 如果没有搜索接口，返回模拟数据（开发阶段）
    return { success: false, message: '搜索功能暂未开放' }
  }
}

// 搜索用户（通过用户名）
export const searchUserByUsername = async (username) => {
  try {
    const response = await api.get(`/api/users/search?username=${username}`)
    if (response.data.success) {
      return { success: true, data: response.data.data }
    }
    return { success: false }
  } catch (error) {
    return { success: false }
  }
}

