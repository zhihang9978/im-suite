// S+级前端性能优化工具

import { ref, computed } from 'vue'

/**
 * 防抖函数（搜索输入优化）
 * @param {Function} func 要执行的函数
 * @param {number} wait 等待时间（毫秒）
 * @returns {Function} 防抖后的函数
 */
export function debounce(func, wait = 300) {
  let timeout
  return function executedFunction(...args) {
    const later = () => {
      clearTimeout(timeout)
      func(...args)
    }
    clearTimeout(timeout)
    timeout = setTimeout(later, wait)
  }
}

/**
 * 节流函数（滚动事件优化）
 * @param {Function} func 要执行的函数
 * @param {number} limit 时间限制（毫秒）
 * @returns {Function} 节流后的函数
 */
export function throttle(func, limit = 300) {
  let inThrottle
  return function executedFunction(...args) {
    if (!inThrottle) {
      func(...args)
      inThrottle = true
      setTimeout(() => (inThrottle = false), limit)
    }
  }
}

/**
 * 乐观更新（立即UI反馈）
 * @param {Function} optimisticUpdate UI立即更新函数
 * @param {Function} apiCall API调用函数
 * @param {Function} rollback 回滚函数（API失败时）
 */
export async function optimisticUpdate(optimisticUpdate, apiCall, rollback) {
  // 1. 立即更新UI
  optimisticUpdate()
  
  try {
    // 2. 调用API
    await apiCall()
  } catch (error) {
    // 3. API失败，回滚UI
    rollback()
    throw error
  }
}

/**
 * 图片懒加载
 * @param {string} src 图片URL
 * @param {string} placeholder 占位图
 * @returns {Object} Vue组合式API对象
 */
export function useLazyImage(src, placeholder = '/placeholder.png') {
  const imageSrc = ref(placeholder)
  const loaded = ref(false)
  
  const observer = new IntersectionObserver((entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        const img = new Image()
        img.src = src
        img.onload = () => {
          imageSrc.value = src
          loaded.value = true
        }
        observer.disconnect()
      }
    })
  })
  
  return { imageSrc, loaded, observer }
}

/**
 * 虚拟滚动辅助函数
 * @param {Array} items 所有数据
 * @param {number} itemHeight 每项高度
 * @param {number} visibleCount 可见数量
 * @returns {Object} 虚拟滚动数据
 */
export function useVirtualScroll(items, itemHeight = 50, visibleCount = 20) {
  const scrollTop = ref(0)
  
  const startIndex = computed(() => {
    return Math.floor(scrollTop.value / itemHeight)
  })
  
  const endIndex = computed(() => {
    return startIndex.value + visibleCount
  })
  
  const visibleItems = computed(() => {
    return items.value.slice(startIndex.value, endIndex.value)
  })
  
  const totalHeight = computed(() => {
    return items.value.length * itemHeight
  })
  
  const offsetY = computed(() => {
    return startIndex.value * itemHeight
  })
  
  const handleScroll = (event) => {
    scrollTop.value = event.target.scrollTop
  }
  
  return {
    visibleItems,
    totalHeight,
    offsetY,
    handleScroll
  }
}

/**
 * 网络状态监控
 * @returns {Object} 网络状态对象
 */
export function useNetworkStatus() {
  const isOnline = ref(navigator.onLine)
  const effectiveType = ref(null)
  
  // 监听网络状态变化
  window.addEventListener('online', () => {
    isOnline.value = true
  })
  
  window.addEventListener('offline', () => {
    isOnline.value = false
  })
  
  // 监听网络类型变化（如果支持）
  if ('connection' in navigator) {
    const connection = navigator.connection || navigator.mozConnection || navigator.webkitConnection
    effectiveType.value = connection.effectiveType
    
    connection.addEventListener('change', () => {
      effectiveType.value = connection.effectiveType
    })
  }
  
  return {
    isOnline,
    effectiveType,
    isSlow: computed(() => effectiveType.value === 'slow-2g' || effectiveType.value === '2g')
  }
}

/**
 * 请求去重（防止重复提交）
 */
const pendingRequests = new Map()

export function requestDeduplication(key) {
  if (pendingRequests.has(key)) {
    return pendingRequests.get(key)
  }
  
  return {
    start: (promise) => {
      pendingRequests.set(key, promise)
      promise.finally(() => {
        pendingRequests.delete(key)
      })
      return promise
    }
  }
}

/**
 * 性能监控
 */
export function measurePerformance(name, fn) {
  const start = performance.now()
  const result = fn()
  
  if (result instanceof Promise) {
    return result.finally(() => {
      const duration = performance.now() - start
      console.log(`[Performance] ${name}: ${duration.toFixed(2)}ms`)
      
      // 慢操作告警（>1秒）
      if (duration > 1000) {
        console.warn(`[Performance Warning] ${name}耗时过长: ${duration.toFixed(2)}ms`)
      }
    })
  } else {
    const duration = performance.now() - start
    console.log(`[Performance] ${name}: ${duration.toFixed(2)}ms`)
    return result
  }
}

