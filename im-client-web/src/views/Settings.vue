<template>
  <div class="settings-container">
    <el-card>
      <template #header>
        <h3>设置</h3>
      </template>
      
      <el-form label-width="100px">
        <el-divider content-position="left">个人信息</el-divider>
        
        <el-form-item label="头像">
          <el-avatar :size="80" :src="userStore.user?.avatar">
            {{ userStore.user?.nickname?.charAt(0) || 'U' }}
          </el-avatar>
        </el-form-item>
        
        <el-form-item label="昵称">
          <el-input v-model="userInfo.nickname" disabled />
        </el-form-item>
        
        <el-form-item label="手机号">
          <el-input v-model="userInfo.phone" disabled />
        </el-form-item>
        
        <el-form-item label="用户名">
          <el-input v-model="userInfo.username" disabled />
        </el-form-item>
        
        <el-divider content-position="left">通知设置</el-divider>
        
        <el-form-item label="消息通知">
          <el-switch v-model="settings.messageNotification" />
        </el-form-item>
        
        <el-form-item label="声音提示">
          <el-switch v-model="settings.soundNotification" />
        </el-form-item>
        
        <el-divider content-position="left">隐私设置</el-divider>
        
        <el-form-item label="在线状态">
          <el-switch v-model="settings.showOnlineStatus" />
        </el-form-item>
        
        <el-form-item label="已读回执">
          <el-switch v-model="settings.readReceipt" />
        </el-form-item>
        
        <el-form-item>
          <el-button type="danger" @click="handleLogout">退出登录</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

const userInfo = computed(() => ({
  nickname: userStore.user?.nickname || '',
  phone: userStore.user?.phone || '',
  username: userStore.user?.username || '',
}))

const settings = ref({
  messageNotification: true,
  soundNotification: true,
  showOnlineStatus: true,
  readReceipt: true,
})

const handleLogout = async () => {
  await userStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.settings-container {
  padding: 20px;
  height: 100vh;
  background: #f5f5f5;
}
</style>

