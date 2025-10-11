<template>
  <div id="app">
    <router-view />
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { useRoute } from 'vue-router'

const userStore = useUserStore()
const route = useRoute()

onMounted(() => {
  // 只在非登录页面初始化用户信息 (important-comment)
  // Only initialize user info on non-login pages (important-comment)
  if (route.path !== '/login') {
    userStore.initUser()
  }
})
</script>

<style lang="scss">
#app {
  height: 100vh;
  width: 100vw;
}
</style>
