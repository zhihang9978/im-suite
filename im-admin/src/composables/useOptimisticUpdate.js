// S+级用户体验优化：乐观更新

import { ElMessage } from 'element-plus'

/**
 * 乐观更新组合式API
 * 立即更新UI，然后调用API，失败时回滚
 * 
 * @example
 * const { execute } = useOptimisticUpdate()
 * 
 * await execute({
 *   optimistic: () => {
 *     users.value = users.value.filter(u => u.id !== userId)
 *   },
 *   api: () => request.delete(`/users/${userId}`),
 *   rollback: () => {
 *     users.value.push(deletedUser)
 *   },
 *   successMessage: '删除成功',
 *   errorMessage: '删除失败'
 * })
 */
export function useOptimisticUpdate() {
  const execute = async ({
    optimistic,
    api,
    rollback,
    successMessage = '操作成功',
    errorMessage = '操作失败'
  }) => {
    // 1. 立即执行乐观更新
    const optimisticResult = optimistic()
    
    try {
      // 2. 调用API
      const result = await api()
      
      // 3. API成功，显示成功提示
      if (successMessage) {
        ElMessage.success(successMessage)
      }
      
      return result
    } catch (error) {
      // 4. API失败，回滚UI
      if (rollback) {
        rollback(optimisticResult)
      }
      
      // 5. 显示错误提示
      if (errorMessage) {
        ElMessage.error(errorMessage + ': ' + (error.response?.data?.error || error.message))
      }
      
      throw error
    }
  }
  
  return { execute }
}

/**
 * 批量操作优化
 * 将多个API调用合并为一次批量请求
 */
export function useBatchOperation() {
  const queue = []
  let timer = null
  
  const add = (operation) => {
    queue.push(operation)
    
    // 100ms内的操作合并为一次批量请求
    if (timer) {
      clearTimeout(timer)
    }
    
    timer = setTimeout(async () => {
      const batch = [...queue]
      queue.length = 0
      
      try {
        // 批量执行
        await Promise.all(batch.map(op => op()))
        ElMessage.success(`批量操作成功（${batch.length}项）`)
      } catch (error) {
        ElMessage.error('批量操作部分失败')
      }
    }, 100)
  }
  
  return { add }
}

