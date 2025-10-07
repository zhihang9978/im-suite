<template>
  <div class="plugin-management">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1>插件管理</h1>
      <p>管理志航密信插件系统</p>
    </div>

    <!-- 插件统计 -->
    <div class="plugin-stats">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-number">{{ pluginStats.total }}</div>
              <div class="stat-label">总插件数</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-number">{{ pluginStats.enabled }}</div>
              <div class="stat-label">已启用</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-number">{{ pluginStats.disabled }}</div>
              <div class="stat-label">已禁用</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-number">{{ pluginStats.updates }}</div>
              <div class="stat-label">可更新</div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 插件操作栏 -->
    <div class="plugin-actions">
      <el-button type="primary" @click="showInstallDialog = true">
        <el-icon><Plus /></el-icon>
        安装插件
      </el-button>
      <el-button @click="refreshPlugins">
        <el-icon><Refresh /></el-icon>
        刷新列表
      </el-button>
      <el-button @click="checkUpdates">
        <el-icon><Download /></el-icon>
        检查更新
      </el-button>
    </div>

    <!-- 插件列表 -->
    <div class="plugin-list">
      <el-table :data="plugins" v-loading="loading">
        <el-table-column prop="name" label="插件名称" width="200">
          <template #default="{ row }">
            <div class="plugin-name">
              <el-icon v-if="row.icon"><component :is="row.icon" /></el-icon>
              <span>{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="version" label="版本" width="100" />
        
        <el-table-column prop="author" label="作者" width="150" />
        
        <el-table-column prop="description" label="描述" min-width="200" />
        
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'enabled' ? 'success' : 'info'">
              {{ row.status === 'enabled' ? '已启用' : '已禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="compatibility" label="兼容性" width="100">
          <template #default="{ row }">
            <el-tag :type="row.compatibility === 'compatible' ? 'success' : 'warning'">
              {{ row.compatibility === 'compatible' ? '兼容' : '不兼容' }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button 
              v-if="row.status === 'disabled'" 
              type="success" 
              size="small" 
              @click="enablePlugin(row)"
            >
              启用
            </el-button>
            <el-button 
              v-if="row.status === 'enabled'" 
              type="warning" 
              size="small" 
              @click="disablePlugin(row)"
            >
              禁用
            </el-button>
            <el-button 
              type="primary" 
              size="small" 
              @click="showPluginSettings(row)"
            >
              设置
            </el-button>
            <el-button 
              type="danger" 
              size="small" 
              @click="uninstallPlugin(row)"
            >
              卸载
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 安装插件对话框 -->
    <el-dialog 
      v-model="showInstallDialog" 
      title="安装插件" 
      width="600px"
    >
      <div class="install-form">
        <el-form :model="installForm" label-width="100px">
          <el-form-item label="插件文件">
            <el-upload
              ref="uploadRef"
              :auto-upload="false"
              :on-change="handleFileChange"
              :file-list="fileList"
              accept=".zip,.tar.gz"
            >
              <el-button type="primary">选择文件</el-button>
            </el-upload>
          </el-form-item>
          
          <el-form-item label="插件URL">
            <el-input 
              v-model="installForm.url" 
              placeholder="输入插件下载URL"
            />
          </el-form-item>
          
          <el-form-item label="插件名称">
            <el-input 
              v-model="installForm.name" 
              placeholder="输入插件名称"
            />
          </el-form-item>
          
          <el-form-item label="插件描述">
            <el-input 
              v-model="installForm.description" 
              type="textarea" 
              placeholder="输入插件描述"
            />
          </el-form-item>
        </el-form>
      </div>
      
      <template #footer>
        <el-button @click="showInstallDialog = false">取消</el-button>
        <el-button type="primary" @click="installPlugin">安装</el-button>
      </template>
    </el-dialog>

    <!-- 插件设置对话框 -->
    <el-dialog 
      v-model="showSettingsDialog" 
      title="插件设置" 
      width="800px"
    >
      <div class="plugin-settings">
        <div class="plugin-info">
          <h3>{{ selectedPlugin?.name }}</h3>
          <p>{{ selectedPlugin?.description }}</p>
          <p>版本: {{ selectedPlugin?.version }}</p>
          <p>作者: {{ selectedPlugin?.author }}</p>
        </div>
        
        <div class="plugin-config">
          <h4>插件配置</h4>
          <el-form :model="pluginConfig" label-width="120px">
            <el-form-item 
              v-for="(config, key) in pluginConfig" 
              :key="key" 
              :label="config.label"
            >
              <el-input 
                v-if="config.type === 'text'" 
                v-model="config.value" 
                :placeholder="config.placeholder"
              />
              <el-switch 
                v-else-if="config.type === 'boolean'" 
                v-model="config.value"
              />
              <el-select 
                v-else-if="config.type === 'select'" 
                v-model="config.value"
              >
                <el-option 
                  v-for="option in config.options" 
                  :key="option.value" 
                  :label="option.label" 
                  :value="option.value"
                />
              </el-select>
            </el-form-item>
          </el-form>
        </div>
      </div>
      
      <template #footer>
        <el-button @click="showSettingsDialog = false">取消</el-button>
        <el-button type="primary" @click="savePluginSettings">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Download } from '@element-plus/icons-vue'

// 响应式数据
const loading = ref(false)
const showInstallDialog = ref(false)
const showSettingsDialog = ref(false)
const selectedPlugin = ref(null)
const fileList = ref([])

// 插件统计
const pluginStats = reactive({
  total: 0,
  enabled: 0,
  disabled: 0,
  updates: 0
})

// 插件列表
const plugins = ref([])

// 安装表单
const installForm = reactive({
  file: null,
  url: '',
  name: '',
  description: ''
})

// 插件配置
const pluginConfig = ref({})

// 初始化
onMounted(() => {
  loadPlugins()
  loadPluginStats()
})

// 加载插件列表
const loadPlugins = async () => {
  loading.value = true
  try {
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    plugins.value = [
      {
        id: 1,
        name: '消息加密插件',
        version: '1.0.0',
        author: '志航密信团队',
        description: '提供端到端加密功能',
        status: 'enabled',
        compatibility: 'compatible',
        icon: 'Lock'
      },
      {
        id: 2,
        name: '文件管理插件',
        version: '2.1.0',
        author: '第三方开发者',
        description: '增强文件管理功能',
        status: 'disabled',
        compatibility: 'compatible',
        icon: 'Folder'
      },
      {
        id: 3,
        name: '主题定制插件',
        version: '1.5.0',
        author: '主题设计师',
        description: '自定义主题和样式',
        status: 'enabled',
        compatibility: 'incompatible',
        icon: 'Palette'
      }
    ]
  } catch (error) {
    ElMessage.error('加载插件列表失败')
  } finally {
    loading.value = false
  }
}

// 加载插件统计
const loadPluginStats = () => {
  pluginStats.total = plugins.value.length
  pluginStats.enabled = plugins.value.filter(p => p.status === 'enabled').length
  pluginStats.disabled = plugins.value.filter(p => p.status === 'disabled').length
  pluginStats.updates = plugins.value.filter(p => p.hasUpdate).length
}

// 刷新插件列表
const refreshPlugins = () => {
  loadPlugins()
  loadPluginStats()
}

// 检查更新
const checkUpdates = async () => {
  loading.value = true
  try {
    // 模拟检查更新
    await new Promise(resolve => setTimeout(resolve, 1000))
    ElMessage.success('检查完成，发现 2 个插件可更新')
  } catch (error) {
    ElMessage.error('检查更新失败')
  } finally {
    loading.value = false
  }
}

// 启用插件
const enablePlugin = async (plugin) => {
  try {
    await ElMessageBox.confirm('确定要启用此插件吗？', '确认操作')
    
    plugin.status = 'enabled'
    ElMessage.success('插件已启用')
    loadPluginStats()
  } catch (error) {
    // 用户取消
  }
}

// 禁用插件
const disablePlugin = async (plugin) => {
  try {
    await ElMessageBox.confirm('确定要禁用此插件吗？', '确认操作')
    
    plugin.status = 'disabled'
    ElMessage.success('插件已禁用')
    loadPluginStats()
  } catch (error) {
    // 用户取消
  }
}

// 卸载插件
const uninstallPlugin = async (plugin) => {
  try {
    await ElMessageBox.confirm('确定要卸载此插件吗？此操作不可撤销。', '确认操作', {
      type: 'warning'
    })
    
    const index = plugins.value.findIndex(p => p.id === plugin.id)
    if (index > -1) {
      plugins.value.splice(index, 1)
      ElMessage.success('插件已卸载')
      loadPluginStats()
    }
  } catch (error) {
    // 用户取消
  }
}

// 显示插件设置
const showPluginSettings = (plugin) => {
  selectedPlugin.value = plugin
  
  // 模拟插件配置
  pluginConfig.value = {
    enabled: {
      label: '启用插件',
      type: 'boolean',
      value: plugin.status === 'enabled'
    },
    autoStart: {
      label: '自动启动',
      type: 'boolean',
      value: true
    },
    logLevel: {
      label: '日志级别',
      type: 'select',
      value: 'info',
      options: [
        { label: '调试', value: 'debug' },
        { label: '信息', value: 'info' },
        { label: '警告', value: 'warn' },
        { label: '错误', value: 'error' }
      ]
    }
  }
  
  showSettingsDialog.value = true
}

// 保存插件设置
const savePluginSettings = () => {
  ElMessage.success('插件设置已保存')
  showSettingsDialog.value = false
}

// 处理文件选择
const handleFileChange = (file) => {
  installForm.file = file.raw
}

// 安装插件
const installPlugin = async () => {
  if (!installForm.file && !installForm.url) {
    ElMessage.warning('请选择插件文件或输入插件URL')
    return
  }
  
  try {
    loading.value = true
    
    // 模拟安装过程
    await new Promise(resolve => setTimeout(resolve, 2000))
    
    ElMessage.success('插件安装成功')
    showInstallDialog.value = false
    
    // 重置表单
    installForm.file = null
    installForm.url = ''
    installForm.name = ''
    installForm.description = ''
    fileList.value = []
    
    // 刷新列表
    loadPlugins()
    loadPluginStats()
  } catch (error) {
    ElMessage.error('插件安装失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.plugin-management {
  padding: 20px;
}

.page-header {
  margin-bottom: 30px;
}

.page-header h1 {
  margin: 0 0 10px 0;
  color: #303133;
}

.page-header p {
  margin: 0;
  color: #909399;
}

.plugin-stats {
  margin-bottom: 30px;
}

.stat-card {
  text-align: center;
}

.stat-content {
  padding: 20px;
}

.stat-number {
  font-size: 32px;
  font-weight: bold;
  color: #409EFF;
  margin-bottom: 10px;
}

.stat-label {
  color: #909399;
  font-size: 14px;
}

.plugin-actions {
  margin-bottom: 20px;
}

.plugin-actions .el-button {
  margin-right: 10px;
}

.plugin-list {
  background: white;
  border-radius: 4px;
  padding: 20px;
}

.plugin-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.install-form {
  padding: 20px 0;
}

.plugin-settings {
  padding: 20px 0;
}

.plugin-info {
  margin-bottom: 30px;
  padding-bottom: 20px;
  border-bottom: 1px solid #EBEEF5;
}

.plugin-info h3 {
  margin: 0 0 10px 0;
  color: #303133;
}

.plugin-info p {
  margin: 5px 0;
  color: #606266;
}

.plugin-config h4 {
  margin: 0 0 20px 0;
  color: #303133;
}
</style>
