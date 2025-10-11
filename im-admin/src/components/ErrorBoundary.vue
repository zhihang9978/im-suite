<template>
  <div v-if="hasError" class="error-boundary">
    <div class="error-content">
      <el-result icon="error" title="页面出现错误">
        <template #sub-title>
          <p>抱歉，页面遇到了问题</p>
          <p v-if="errorMessage" class="error-message">{{ errorMessage }}</p>
        </template>
        <template #extra>
          <el-button type="primary" @click="handleReload">重新加载</el-button>
          <el-button @click="handleGoHome">返回首页</el-button>
        </template>
      </el-result>
    </div>
  </div>
  <slot v-else></slot>
</template>

<script setup>
import { ref, onErrorCaptured } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

const router = useRouter()
const hasError = ref(false)
const errorMessage = ref('')

// 捕获子组件错误
onErrorCaptured((err, instance, info) => {
  console.error('ErrorBoundary caught error:', err)
  console.error('Error info:', info)
  
  hasError.value = true
  errorMessage.value = err.message || '未知错误'
  
  // 上报错误到监控系统（如Sentry）
  if (window.Sentry) {
    window.Sentry.captureException(err, {
      contexts: {
        vue: {
          componentName: instance?.$options?.name || 'Unknown',
          errorInfo: info
        }
      }
    })
  }
  
  // 阻止错误继续传播
  return false
})

const handleReload = () => {
  hasError.value = false
  errorMessage.value = ''
  window.location.reload()
}

const handleGoHome = () => {
  hasError.value = false
  errorMessage.value = ''
  router.push('/')
}
</script>

<style scoped>
.error-boundary {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: #f5f7fa;
}

.error-content {
  background: white;
  padding: 40px;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.error-message {
  margin-top: 10px;
  color: #909399;
  font-size: 14px;
}
</style>

