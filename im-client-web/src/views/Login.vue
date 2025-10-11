<template>
  <div class="login-container">
    <el-card class="login-card">
      <template #header>
        <div class="card-header">
          <span class="logo">💬</span>
          <h2>志航密信</h2>
        </div>
      </template>
      
      <el-tabs v-model="activeTab">
        <!-- 登录标签 -->
        <el-tab-pane label="登录" name="login">
          <el-form :model="loginForm" :rules="loginRules" ref="loginFormRef">
            <el-form-item prop="phone">
              <el-input
                v-model="loginForm.phone"
                placeholder="请输入手机号"
                prefix-icon="Phone"
                size="large"
              />
            </el-form-item>
            
            <el-form-item prop="password">
              <el-input
                v-model="loginForm.password"
                type="password"
                placeholder="请输入密码"
                prefix-icon="Lock"
                size="large"
                show-password
                @keyup.enter="handleLogin"
              />
            </el-form-item>
            
            <el-form-item>
              <el-button
                type="primary"
                :loading="loading"
                @click="handleLogin"
                size="large"
                class="login-button"
              >
                登录
              </el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
        
        <!-- 注册标签 -->
        <el-tab-pane label="注册" name="register">
          <el-form :model="registerForm" :rules="registerRules" ref="registerFormRef">
            <el-form-item prop="phone">
              <el-input
                v-model="registerForm.phone"
                placeholder="请输入手机号"
                prefix-icon="Phone"
                size="large"
              />
            </el-form-item>
            
            <el-form-item prop="password">
              <el-input
                v-model="registerForm.password"
                type="password"
                placeholder="请输入密码（至少6位）"
                prefix-icon="Lock"
                size="large"
                show-password
              />
            </el-form-item>
            
            <el-form-item prop="nickname">
              <el-input
                v-model="registerForm.nickname"
                placeholder="请输入昵称（可选）"
                prefix-icon="User"
                size="large"
              />
            </el-form-item>
            
            <el-form-item>
              <el-button
                type="primary"
                :loading="loading"
                @click="handleRegister"
                size="large"
                class="login-button"
              >
                注册
              </el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'

const router = useRouter()
const userStore = useUserStore()

const activeTab = ref('login')
const loading = ref(false)

// 登录表单
const loginForm = ref({
  phone: '',
  password: '',
})

const loginRules = {
  phone: [{ required: true, message: '请输入手机号', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

const loginFormRef = ref(null)

// 注册表单
const registerForm = ref({
  phone: '',
  password: '',
  nickname: '',
})

const registerRules = {
  phone: [{ required: true, message: '请输入手机号', trigger: 'blur' }],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' },
  ],
}

const registerFormRef = ref(null)

// 处理登录
const handleLogin = async () => {
  const formRef = loginFormRef.value
  if (!formRef) return
  
  await formRef.validate(async (valid) => {
    if (valid) {
      loading.value = true
      const result = await userStore.login(loginForm.value.phone, loginForm.value.password)
      loading.value = false
      
      if (result.success) {
        ElMessage.success('登录成功')
        router.push('/')
      } else {
        ElMessage.error(result.message || '登录失败')
      }
    }
  })
}

// 处理注册
const handleRegister = async () => {
  const formRef = registerFormRef.value
  if (!formRef) return
  
  await formRef.validate(async (valid) => {
    if (valid) {
      loading.value = true
      const result = await userStore.register(
        registerForm.value.phone,
        registerForm.value.password,
        registerForm.value.nickname
      )
      loading.value = false
      
      if (result.success) {
        ElMessage.success('注册成功')
        router.push('/')
      } else {
        ElMessage.error(result.message || '注册失败')
      }
    }
  })
}
</script>

<style scoped>
.login-container {
  width: 100%;
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 400px;
  border-radius: 10px;
}

.card-header {
  text-align: center;
}

.logo {
  font-size: 48px;
  display: block;
  margin-bottom: 10px;
}

.card-header h2 {
  margin: 0;
  color: #333;
}

.login-button {
  width: 100%;
}
</style>

