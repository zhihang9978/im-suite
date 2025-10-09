<template>
  <div class="two-factor-settings">
    <el-card class="settings-card">
      <template #header>
        <div class="card-header">
          <h2><el-icon><Lock /></el-icon> 双因子认证设置</h2>
          <el-tag v-if="twoFactorStatus.enabled" type="success">已启用</el-tag>
          <el-tag v-else type="info">未启用</el-tag>
        </div>
      </template>

      <!-- 2FA未启用 -->
      <div v-if="!twoFactorStatus.enabled" class="enable-section">
        <el-alert
          title="什么是双因子认证？"
          type="info"
          :closable="false"
          show-icon
        >
          双因子认证（2FA）为您的账户提供额外的安全保护。即使有人知道您的密码，没有您的手机也无法登录。
        </el-alert>

        <el-form
          ref="enableForm"
          :model="enableFormData"
          :rules="enableRules"
          label-width="120px"
          class="enable-form"
        >
          <el-form-item label="验证密码" prop="password">
            <el-input
              v-model="enableFormData.password"
              type="password"
              placeholder="请输入您的密码"
              show-password
            />
          </el-form-item>

          <el-form-item>
            <el-button type="primary" @click="handleEnable" :loading="loading">
              启用双因子认证
            </el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 2FA已启用 -->
      <div v-else class="enabled-section">
        <!-- 状态信息 -->
        <el-descriptions :column="2" border>
          <el-descriptions-item label="状态">
            <el-tag type="success">已启用</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="受信任设备">
            {{ twoFactorStatus.trusted_devices_count }} 台
          </el-descriptions-item>
          <el-descriptions-item label="剩余备用码">
            {{ twoFactorStatus.backup_codes_remaining }} 个
          </el-descriptions-item>
        </el-descriptions>

        <!-- 操作按钮 -->
        <div class="action-buttons">
          <el-button @click="showTrustedDevices">
            <el-icon><Monitor /></el-icon> 管理受信任设备
          </el-button>
          <el-button @click="handleRegenerateBackupCodes">
            <el-icon><Refresh /></el-icon> 重新生成备用码
          </el-button>
          <el-button type="danger" @click="handleDisable">
            <el-icon><Close /></el-icon> 禁用双因子认证
          </el-button>
        </div>

        <!-- 最近验证记录 -->
        <div class="recent-attempts" v-if="twoFactorStatus.recent_attempts">
          <h3>最近验证记录</h3>
          <el-table :data="twoFactorStatus.recent_attempts" style="width: 100%">
            <el-table-column prop="method" label="验证方式" width="120">
              <template #default="scope">
                <el-tag v-if="scope.row.method === 'totp'" size="small">TOTP</el-tag>
                <el-tag v-else-if="scope.row.method === 'backup_code'" size="small" type="warning">备用码</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="scope">
                <el-tag v-if="scope.row.status === 'success'" type="success" size="small">成功</el-tag>
                <el-tag v-else type="danger" size="small">失败</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="ip" label="IP地址" width="150" />
            <el-table-column prop="created_at" label="时间">
              <template #default="scope">
                {{ formatTime(scope.row.created_at) }}
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </el-card>

    <!-- 启用对话框 - 显示二维码 -->
    <el-dialog v-model="qrDialog" title="启用双因子认证" width="600px">
      <div class="qr-section">
        <el-steps :active="qrStep" align-center>
          <el-step title="扫描二维码" />
          <el-step title="输入验证码" />
          <el-step title="保存备用码" />
        </el-steps>

        <!-- 步骤1: 扫描二维码 -->
        <div v-if="qrStep === 0" class="step-content">
          <el-alert
            title="请使用验证器APP扫描二维码"
            type="info"
            :closable="false"
            show-icon
            class="mb-3"
          >
            推荐使用 Google Authenticator、Microsoft Authenticator 或 Authy
          </el-alert>
          
          <div class="qr-code">
            <img :src="qrCodeData" alt="QR Code" />
          </div>
          
          <div class="secret-key">
            <p>如果无法扫描，请手动输入密钥：</p>
            <el-input v-model="secretKey" readonly>
              <template #append>
                <el-button @click="copySecret">复制</el-button>
              </template>
            </el-input>
          </div>

          <div class="dialog-footer">
            <el-button @click="qrDialog = false">取消</el-button>
            <el-button type="primary" @click="qrStep = 1">下一步</el-button>
          </div>
        </div>

        <!-- 步骤2: 输入验证码 -->
        <div v-if="qrStep === 1" class="step-content">
          <el-form :model="verifyFormData" :rules="verifyRules" ref="verifyForm">
            <el-form-item label="验证码" prop="code">
              <el-input
                v-model="verifyFormData.code"
                placeholder="请输入6位验证码"
                maxlength="6"
                class="code-input"
              />
            </el-form-item>
          </el-form>

          <div class="dialog-footer">
            <el-button @click="qrStep = 0">上一步</el-button>
            <el-button type="primary" @click="handleVerify" :loading="loading">
              验证并启用
            </el-button>
          </div>
        </div>

        <!-- 步骤3: 保存备用码 -->
        <div v-if="qrStep === 2" class="step-content">
          <el-alert
            title="请妥善保存这些备用码"
            type="warning"
            :closable="false"
            show-icon
            class="mb-3"
          >
            当您无法访问验证器APP时，可以使用备用码登录。每个备用码只能使用一次。
          </el-alert>

          <div class="backup-codes">
            <el-tag
              v-for="(code, index) in backupCodes"
              :key="index"
              size="large"
              class="code-tag"
            >
              {{ code }}
            </el-tag>
          </div>

          <div class="dialog-footer">
            <el-button @click="downloadBackupCodes">
              <el-icon><Download /></el-icon> 下载备用码
            </el-button>
            <el-button type="primary" @click="finishSetup">
              我已保存，完成设置
            </el-button>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 禁用对话框 -->
    <el-dialog v-model="disableDialog" title="禁用双因子认证" width="500px">
      <el-alert
        title="警告"
        type="warning"
        :closable="false"
        show-icon
        class="mb-3"
      >
        禁用双因子认证将降低您账户的安全性
      </el-alert>

      <el-form :model="disableFormData" :rules="disableRules" ref="disableForm">
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="disableFormData.password"
            type="password"
            placeholder="请输入您的密码"
            show-password
          />
        </el-form-item>
        <el-form-item label="验证码" prop="code">
          <el-input
            v-model="disableFormData.code"
            placeholder="请输入6位验证码"
            maxlength="6"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="disableDialog = false">取消</el-button>
        <el-button type="danger" @click="confirmDisable" :loading="loading">
          确认禁用
        </el-button>
      </template>
    </el-dialog>

    <!-- 受信任设备对话框 -->
    <el-dialog v-model="devicesDialog" title="受信任设备管理" width="800px">
      <el-table :data="trustedDevices" style="width: 100%">
        <el-table-column prop="device_name" label="设备名称" />
        <el-table-column prop="device_type" label="类型" width="100">
          <template #default="scope">
            <el-tag size="small">{{ scope.row.device_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="ip" label="IP地址" width="150" />
        <el-table-column prop="last_used_at" label="最后使用" width="180">
          <template #default="scope">
            {{ formatTime(scope.row.last_used_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="scope">
            <el-button
              type="danger"
              size="small"
              @click="removeDevice(scope.row.device_id)"
            >
              移除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Lock, Monitor, Refresh, Close, Download } from '@element-plus/icons-vue'
import axios from 'axios'

// 状态
const loading = ref(false)
const twoFactorStatus = reactive({
  enabled: false,
  trusted_devices_count: 0,
  backup_codes_remaining: 0,
  recent_attempts: []
})

// 对话框
const qrDialog = ref(false)
const disableDialog = ref(false)
const devicesDialog = ref(false)

// QR相关
const qrStep = ref(0)
const qrCodeData = ref('')
const secretKey = ref('')
const backupCodes = ref([])

// 受信任设备
const trustedDevices = ref([])

// 表单数据
const enableFormData = reactive({
  password: ''
})

const verifyFormData = reactive({
  code: ''
})

const disableFormData = reactive({
  password: '',
  code: ''
})

// 表单验证规则
const enableRules = {
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const verifyRules = {
  code: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { len: 6, message: '验证码必须是6位', trigger: 'blur' }
  ]
}

const disableRules = {
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  code: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { len: 6, message: '验证码必须是6位', trigger: 'blur' }
  ]
}

// 加载2FA状态
const loadTwoFactorStatus = async () => {
  try {
    const response = await axios.get('/api/2fa/status')
    Object.assign(twoFactorStatus, response.data.data)
  } catch (error) {
    ElMessage.error('加载2FA状态失败')
  }
}

// 启用2FA
const handleEnable = async () => {
  loading.value = true
  try {
    const response = await axios.post('/api/2fa/enable', enableFormData)
    qrCodeData.value = response.data.data.qr_code
    secretKey.value = response.data.data.secret
    backupCodes.value = response.data.data.backup_codes
    qrDialog.value = true
    qrStep.value = 0
    ElMessage.success(response.data.message)
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '启用失败')
  } finally {
    loading.value = false
  }
}

// 验证并启用
const handleVerify = async () => {
  loading.value = true
  try {
    await axios.post('/api/2fa/verify', verifyFormData)
    qrStep.value = 2
    ElMessage.success('验证成功！')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '验证失败')
  } finally {
    loading.value = false
  }
}

// 完成设置
const finishSetup = () => {
  qrDialog.value = false
  qrStep.value = 0
  enableFormData.password = ''
  verifyFormData.code = ''
  loadTwoFactorStatus()
  ElMessage.success('双因子认证已成功启用！')
}

// 禁用2FA
const handleDisable = () => {
  disableDialog.value = true
}

const confirmDisable = async () => {
  loading.value = true
  try {
    await axios.post('/api/2fa/disable', disableFormData)
    disableDialog.value = false
    disableFormData.password = ''
    disableFormData.code = ''
    loadTwoFactorStatus()
    ElMessage.success('双因子认证已禁用')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '禁用失败')
  } finally {
    loading.value = false
  }
}

// 重新生成备用码
const handleRegenerateBackupCodes = async () => {
  try {
    const { value } = await ElMessageBox.prompt('请输入密码和验证码', '重新生成备用码', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputPattern: /.+/,
      inputErrorMessage: '请输入完整信息'
    })
    
    // 这里需要一个更完整的表单，简化处理
    ElMessage.info('请实现完整的备用码重新生成流程')
  } catch {
    // 用户取消
  }
}

// 显示受信任设备
const showTrustedDevices = async () => {
  try {
    const response = await axios.get('/api/2fa/trusted-devices')
    trustedDevices.value = response.data.devices
    devicesDialog.value = true
  } catch (error) {
    ElMessage.error('加载设备列表失败')
  }
}

// 移除设备
const removeDevice = async (deviceId) => {
  try {
    await ElMessageBox.confirm('确定要移除这台设备吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await axios.delete(`/api/2fa/trusted-devices/${deviceId}`)
    ElMessage.success('设备已移除')
    showTrustedDevices()
  } catch {
    // 用户取消
  }
}

// 复制密钥
const copySecret = () => {
  navigator.clipboard.writeText(secretKey.value)
  ElMessage.success('密钥已复制到剪贴板')
}

// 下载备用码
const downloadBackupCodes = () => {
  const content = backupCodes.value.join('\n')
  const blob = new Blob([content], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'backup-codes.txt'
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success('备用码已下载')
}

// 格式化时间
const formatTime = (time) => {
  return new Date(time).toLocaleString('zh-CN')
}

// 页面加载时获取状态
onMounted(() => {
  loadTwoFactorStatus()
})
</script>

<style scoped>
.two-factor-settings {
  padding: 20px;
}

.settings-card {
  max-width: 1200px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-header h2 {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0;
}

.enable-section,
.enabled-section {
  padding: 20px 0;
}

.enable-form {
  margin-top: 20px;
  max-width: 600px;
}

.action-buttons {
  margin: 20px 0;
  display: flex;
  gap: 10px;
}

.recent-attempts {
  margin-top: 30px;
}

.recent-attempts h3 {
  margin-bottom: 15px;
}

.qr-section {
  padding: 20px;
}

.step-content {
  margin-top: 30px;
}

.qr-code {
  display: flex;
  justify-content: center;
  margin: 30px 0;
}

.qr-code img {
  max-width: 300px;
}

.secret-key {
  margin: 20px 0;
}

.backup-codes {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 15px;
  margin: 20px 0;
}

.code-tag {
  font-family: 'Courier New', monospace;
  font-size: 16px;
  padding: 10px 15px;
}

.code-input {
  font-size: 24px;
  text-align: center;
  letter-spacing: 5px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 30px;
}

.mb-3 {
  margin-bottom: 15px;
}
</style>

