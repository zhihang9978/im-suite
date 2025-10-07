<template>
  <div class="system-page">
    <el-row :gutter="20">
      <!-- 系统信息 -->
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>系统信息</span>
          </template>
          <div class="system-info">
            <div class="info-item">
              <span class="label">系统版本:</span>
              <span class="value">志航密信 v1.0.0</span>
            </div>
            <div class="info-item">
              <span class="label">运行时间:</span>
              <span class="value">{{ systemInfo.uptime }}</span>
            </div>
            <div class="info-item">
              <span class="label">CPU 使用率:</span>
              <span class="value">{{ systemInfo.cpu }}%</span>
            </div>
            <div class="info-item">
              <span class="label">内存使用率:</span>
              <span class="value">{{ systemInfo.memory }}%</span>
            </div>
            <div class="info-item">
              <span class="label">磁盘使用率:</span>
              <span class="value">{{ systemInfo.disk }}%</span>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <!-- 服务状态 -->
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>服务状态</span>
          </template>
          <div class="service-status">
            <div class="status-item">
              <span class="label">数据库:</span>
              <el-tag :type="serviceStatus.database ? 'success' : 'danger'">
                {{ serviceStatus.database ? '正常' : '异常' }}
              </el-tag>
            </div>
            <div class="status-item">
              <span class="label">Redis:</span>
              <el-tag :type="serviceStatus.redis ? 'success' : 'danger'">
                {{ serviceStatus.redis ? '正常' : '异常' }}
              </el-tag>
            </div>
            <div class="status-item">
              <span class="label">MinIO:</span>
              <el-tag :type="serviceStatus.minio ? 'success' : 'danger'">
                {{ serviceStatus.minio ? '正常' : '异常' }}
              </el-tag>
            </div>
            <div class="status-item">
              <span class="label">后端服务:</span>
              <el-tag :type="serviceStatus.backend ? 'success' : 'danger'">
                {{ serviceStatus.backend ? '正常' : '异常' }}
              </el-tag>
            </div>
            <div class="status-item">
              <span class="label">Web 服务:</span>
              <el-tag :type="serviceStatus.web ? 'success' : 'danger'">
                {{ serviceStatus.web ? '正常' : '异常' }}
              </el-tag>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    
    <el-row :gutter="20" style="margin-top: 20px;">
      <!-- 系统配置 -->
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>系统配置</span>
          </template>
          <el-form :model="configForm" label-width="120px">
            <el-form-item label="系统名称">
              <el-input v-model="configForm.systemName" />
            </el-form-item>
            <el-form-item label="系统描述">
              <el-input v-model="configForm.systemDesc" type="textarea" />
            </el-form-item>
            <el-form-item label="最大用户数">
              <el-input-number v-model="configForm.maxUsers" :min="1" :max="100000" />
            </el-form-item>
            <el-form-item label="消息保留天数">
              <el-input-number v-model="configForm.messageRetentionDays" :min="1" :max="365" />
            </el-form-item>
            <el-form-item label="文件大小限制(MB)">
              <el-input-number v-model="configForm.maxFileSize" :min="1" :max="1000" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleSaveConfig">保存配置</el-button>
              <el-button @click="handleResetConfig">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
      
      <!-- 系统操作 -->
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>系统操作</span>
          </template>
          <div class="system-actions">
            <el-button type="primary" @click="handleRestartServices">
              <el-icon><Refresh /></el-icon>
              重启服务
            </el-button>
            <el-button type="warning" @click="handleClearCache">
              <el-icon><Delete /></el-icon>
              清理缓存
            </el-button>
            <el-button type="info" @click="handleBackupData">
              <el-icon><Download /></el-icon>
              备份数据
            </el-button>
            <el-button type="danger" @click="handleShutdown">
              <el-icon><SwitchButton /></el-icon>
              关闭系统
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

const systemInfo = ref({
  uptime: '7天 12小时 30分钟',
  cpu: 45.6,
  memory: 67.8,
  disk: 23.4
})

const serviceStatus = ref({
  database: true,
  redis: true,
  minio: true,
  backend: true,
  web: true
})

const configForm = reactive({
  systemName: '志航密信',
  systemDesc: '基于 Telegram 的私有通讯系统',
  maxUsers: 10000,
  messageRetentionDays: 30,
  maxFileSize: 100
})

// 保存配置
const handleSaveConfig = async () => {
  try {
    // 模拟保存
    await new Promise(resolve => setTimeout(resolve, 1000))
    ElMessage.success('配置保存成功')
  } catch (error) {
    ElMessage.error('配置保存失败')
  }
}

// 重置配置
const handleResetConfig = () => {
  Object.assign(configForm, {
    systemName: '志航密信',
    systemDesc: '基于 Telegram 的私有通讯系统',
    maxUsers: 10000,
    messageRetentionDays: 30,
    maxFileSize: 100
  })
  ElMessage.info('配置已重置')
}

// 重启服务
const handleRestartServices = async () => {
  try {
    await ElMessageBox.confirm('确定要重启所有服务吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    ElMessage.success('服务重启成功')
  } catch (error) {
    // 用户取消
  }
}

// 清理缓存
const handleClearCache = async () => {
  try {
    await ElMessageBox.confirm('确定要清理所有缓存吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    ElMessage.success('缓存清理成功')
  } catch (error) {
    // 用户取消
  }
}

// 备份数据
const handleBackupData = async () => {
  try {
    await ElMessageBox.confirm('确定要备份所有数据吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'info'
    })
    
    ElMessage.success('数据备份成功')
  } catch (error) {
    // 用户取消
  }
}

// 关闭系统
const handleShutdown = async () => {
  try {
    await ElMessageBox.confirm('确定要关闭系统吗？这将停止所有服务！', '危险操作', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error'
    })
    
    ElMessage.success('系统关闭成功')
  } catch (error) {
    // 用户取消
  }
}

onMounted(() => {
  // 初始化系统信息
})
</script>

<style lang="scss" scoped>
.system-page {
  .system-info, .service-status {
    .info-item, .status-item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 10px 0;
      border-bottom: 1px solid #ebeef5;
      
      &:last-child {
        border-bottom: none;
      }
      
      .label {
        font-weight: 500;
        color: #606266;
      }
      
      .value {
        color: #303133;
      }
    }
  }
  
  .system-actions {
    display: flex;
    flex-direction: column;
    gap: 15px;
    
    .el-button {
      width: 100%;
      height: 45px;
      font-size: 16px;
    }
  }
}
</style>
