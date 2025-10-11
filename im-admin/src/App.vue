<template>
  <ErrorBoundary>
    <div id="app">
      <router-view />
    </div>
  </ErrorBoundary>
</template>

<script setup>
import { onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import ErrorBoundary from '@/components/ErrorBoundary.vue'

const userStore = useUserStore()

onMounted(() => {
  // 初始化用户信息
  userStore.initUser()
})
</script>

<style lang="scss">
#app {
  height: 100vh;
  width: 100vw;
}
</style>
