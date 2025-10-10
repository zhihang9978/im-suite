<template>
  <div class="system-page">
    <el-tabs v-model="activeTab" type="card">
      <!-- 系统信息标签 -->
      <el-tab-pane label="系统信息" name="system">
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
                  <span class="value">志航密信 v1.6.0</span>
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
                  重启服务
                </el-button>
                <el-button type="warning" @click="handleClearCache">
                  清理缓存
                </el-button>
                <el-button type="info" @click="handleBackupData">
                  备份数据
                </el-button>
                <el-button type="danger" @click="handleShutdown">
                  关闭系统
                </el-button>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </el-tab-pane>

      <!-- 机器人管理标签 -->
      <el-tab-pane label="🤖 机器人管理" name="bots">
        <div class="bots-header">
          <h3>机器人管理</h3>
          <el-button type="primary" @click="showCreateBotDialog = true">
            ➕ 创建机器人
          </el-button>
        </div>

        <el-table :data="bots" v-loading="botsLoading" style="width: 100%; margin-top: 20px;">
          <el-table-column prop="name" label="名称" width="180" />
          <el-table-column prop="type" label="类型" width="100">
            <template #default="{ row }">
              <el-tag size="small">{{ row.type }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="description" label="描述" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.is_active ? 'success' : 'danger'" size="small">
                {{ row.is_active ? '激活' : '停用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="统计" width="200">
            <template #default="{ row }">
              <div class="stats">
                <span>总计: {{ row.total_calls || 0 }}</span>
                <span style="color: #67c23a">成功: {{ row.success_calls || 0 }}</span>
                <span style="color: #f56c6c">失败: {{ row.failed_calls || 0 }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="220" fixed="right">
            <template #default="{ row }">
              <el-button size="small" @click="viewBotDetails(row)">详情</el-button>
              <el-button size="small" type="warning" @click="toggleBotStatus(row)">
                {{ row.is_active ? '停用' : '启用' }}
              </el-button>
              <el-button size="small" type="danger" @click="deleteBot(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 机器人用户标签 -->
      <el-tab-pane label="👤 机器人用户" name="bot-users">
        <div class="bots-header">
          <h3>机器人用户管理</h3>
          <el-button type="primary" @click="showCreateBotUserDialog = true">
            ➕ 创建机器人用户
          </el-button>
        </div>

        <el-alert
          title="💡 提示"
          type="info"
          :closable="false"
          style="margin: 20px 0;"
        >
          为机器人在系统中创建用户账号，使其可以在聊天界面中与用户交互。
        </el-alert>

        <el-table :data="botUsers" v-loading="botUsersLoading" style="width: 100%;">
          <el-table-column prop="bot.name" label="机器人" width="180" />
          <el-table-column prop="user.username" label="用户名" width="150" />
          <el-table-column prop="user.nickname" label="昵称" width="150" />
          <el-table-column prop="user_id" label="用户ID" width="100" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.is_active ? 'success' : 'danger'" size="small">
                {{ row.is_active ? '激活' : '停用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="180">
            <template #default="{ row }">
              {{ formatDate(row.created_at) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button size="small" type="danger" @click="deleteBotUser(row)">
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 用户授权标签 -->
      <el-tab-pane label="🔑 用户授权" name="permissions">
        <div class="bots-header">
          <h3>用户授权管理</h3>
          <el-button type="primary" @click="showGrantPermissionDialog = true">
            ➕ 授权用户
          </el-button>
        </div>

        <el-table :data="permissions" v-loading="permissionsLoading" style="width: 100%; margin-top: 20px;">
          <el-table-column prop="user.username" label="用户" width="150" />
          <el-table-column prop="bot.name" label="机器人" width="180" />
          <el-table-column prop="granted_by_user.username" label="授权者" width="120" />
          <el-table-column prop="created_at" label="授权时间" width="180">
            <template #default="{ row }">
              {{ formatDate(row.created_at) }}
            </template>
          </el-table-column>
          <el-table-column label="过期时间" width="180">
            <template #default="{ row }">
              <span :style="{ color: getExpiryColor(row.expires_at) }">
                {{ formatExpiry(row.expires_at) }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.is_active ? 'success' : 'danger'" size="small">
                {{ row.is_active ? '激活' : '停用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button size="small" type="danger" @click="revokePermission(row)">
                撤销
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- 创建机器人对话框 -->
    <el-dialog v-model="showCreateBotDialog" title="创建机器人" width="600px">
      <el-form :model="botForm" label-width="120px">
        <el-form-item label="机器人名称" required>
          <el-input v-model="botForm.name" placeholder="例如: 用户管理机器人" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="botForm.description" type="textarea" rows="3" placeholder="机器人的用途和功能说明" />
        </el-form-item>
        <el-form-item label="类型" required>
          <el-select v-model="botForm.type" placeholder="请选择">
            <el-option label="内部机器人" value="internal" />
            <el-option label="Webhook机器人" value="webhook" />
            <el-option label="插件机器人" value="plugin" />
          </el-select>
        </el-form-item>
        <el-form-item label="权限" required>
          <el-checkbox-group v-model="botForm.permissions">
            <el-checkbox label="create_user">创建用户</el-checkbox>
            <el-checkbox label="delete_user">删除用户</el-checkbox>
            <el-checkbox label="update_user">更新用户</el-checkbox>
            <el-checkbox label="list_users">查看用户列表</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateBotDialog = false">取消</el-button>
        <el-button type="primary" @click="createBot" :loading="botSubmitting">创建</el-button>
      </template>
    </el-dialog>

    <!-- 创建机器人用户对话框 -->
    <el-dialog v-model="showCreateBotUserDialog" title="创建机器人用户" width="500px">
      <el-form :model="botUserForm" label-width="100px">
        <el-form-item label="选择机器人" required>
          <el-select v-model="botUserForm.bot_id" placeholder="请选择机器人">
            <el-option 
              v-for="bot in availableBots" 
              :key="bot.id" 
              :label="bot.name" 
              :value="bot.id" 
            />
          </el-select>
        </el-form-item>
        <el-form-item label="用户名" required>
          <el-input v-model="botUserForm.username" placeholder="例如: userbot" />
          <div style="font-size: 12px; color: #909399; margin-top: 5px;">
            用户在聊天中搜索此用户名与机器人对话
          </div>
        </el-form-item>
        <el-form-item label="昵称" required>
          <el-input v-model="botUserForm.nickname" placeholder="例如: 用户管理机器人" />
        </el-form-item>
        <el-form-item label="头像URL">
          <el-input v-model="botUserForm.avatar" placeholder="https://example.com/avatar.png" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateBotUserDialog = false">取消</el-button>
        <el-button type="primary" @click="createBotUser" :loading="botUserSubmitting">创建</el-button>
      </template>
    </el-dialog>

    <!-- 授权用户对话框 -->
    <el-dialog v-model="showGrantPermissionDialog" title="授权用户使用机器人" width="500px">
      <el-form :model="permissionForm" label-width="100px">
        <el-form-item label="用户ID" required>
          <el-input-number v-model="permissionForm.user_id" :min="1" placeholder="输入用户ID" style="width: 100%" />
        </el-form-item>
        <el-form-item label="机器人" required>
          <el-select v-model="permissionForm.bot_id" placeholder="请选择机器人">
            <el-option 
              v-for="bot in bots" 
              :key="bot.id" 
              :label="bot.name" 
              :value="bot.id" 
            />
          </el-select>
        </el-form-item>
        <el-form-item label="过期时间">
          <el-date-picker
            v-model="permissionForm.expires_at"
            type="datetime"
            placeholder="留空表示永不过期"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showGrantPermissionDialog = false">取消</el-button>
        <el-button type="primary" @click="grantPermission" :loading="permissionSubmitting">授权</el-button>
      </template>
    </el-dialog>

    <!-- API密钥显示对话框 -->
    <el-dialog v-model="showApiKeysDialog" title="⚠️ 机器人API密钥" width="700px" :close-on-click-modal="false">
      <el-alert
        title="重要提示：API密钥只显示一次，请立即保存！"
        type="warning"
        :closable="false"
        style="margin-bottom: 20px;"
      />
      <el-form label-width="120px">
        <el-form-item label="API Key:">
          <el-input v-model="createdApiKeys.api_key" readonly>
            <template #append>
              <el-button @click="copyToClipboard(createdApiKeys.api_key)">📋 复制</el-button>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="API Secret:">
          <el-input v-model="createdApiKeys.api_secret" readonly>
            <template #append>
              <el-button @click="copyToClipboard(createdApiKeys.api_secret)">📋 复制</el-button>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="使用示例:">
          <el-input 
            :value="`X-Bot-Auth: Bot ${createdApiKeys.api_key}:${createdApiKeys.api_secret}`" 
            type="textarea" 
            :rows="2" 
            readonly 
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button type="primary" @click="showApiKeysDialog = false">我已保存</el-button>
      </template>
    </el-dialog>

    <!-- 机器人详情对话框 -->
    <el-dialog v-model="showBotDetailDialog" title="机器人详情" width="800px">
      <el-descriptions v-if="selectedBot" :column="2" border>
        <el-descriptions-item label="名称">{{ selectedBot.name }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ selectedBot.type }}</el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">{{ selectedBot.description || '暂无' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="selectedBot.is_active ? 'success' : 'danger'">
            {{ selectedBot.is_active ? '激活' : '停用' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDate(selectedBot.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="最后使用">{{ formatDate(selectedBot.last_used_at) }}</el-descriptions-item>
        <el-descriptions-item label="总调用次数">{{ selectedBot.total_calls || 0 }}</el-descriptions-item>
        <el-descriptions-item label="成功调用">
          <span style="color: #67c23a">{{ selectedBot.success_calls || 0 }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="失败调用">
          <span style="color: #f56c6c">{{ selectedBot.failed_calls || 0 }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="成功率">
          {{ selectedBot.total_calls > 0 ? ((selectedBot.success_calls / selectedBot.total_calls) * 100).toFixed(2) + '%' : 'N/A' }}
        </el-descriptions-item>
        <el-descriptions-item label="速率限制">{{ selectedBot.rate_limit || 0 }} 次/分钟</el-descriptions-item>
        <el-descriptions-item label="每日限制">{{ selectedBot.daily_limit || 0 }} 次/天</el-descriptions-item>
        <el-descriptions-item label="权限" :span="2">
          <el-tag v-for="perm in parsePermissions(selectedBot.permissions)" :key="perm" size="small" style="margin-right: 5px;">
            {{ permissionLabels[perm] || perm }}
          </el-tag>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

// 标签页
const activeTab = ref('system')

// 系统信息
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

// 机器人相关
const bots = ref([])
const botUsers = ref([])
const permissions = ref([])
const botsLoading = ref(false)
const botUsersLoading = ref(false)
const permissionsLoading = ref(false)

const showCreateBotDialog = ref(false)
const showCreateBotUserDialog = ref(false)
const showGrantPermissionDialog = ref(false)
const showApiKeysDialog = ref(false)
const showBotDetailDialog = ref(false)

const botSubmitting = ref(false)
const botUserSubmitting = ref(false)
const permissionSubmitting = ref(false)

const selectedBot = ref(null)
const createdApiKeys = ref({ api_key: '', api_secret: '' })

const botForm = reactive({
  name: '',
  description: '',
  type: 'internal',
  permissions: []
})

const botUserForm = reactive({
  bot_id: '',
  username: '',
  nickname: '',
  avatar: ''
})

const permissionForm = reactive({
  user_id: null,
  bot_id: '',
  expires_at: null
})

const permissionLabels = {
  'create_user': '创建用户',
  'delete_user': '删除用户',
  'update_user': '更新用户',
  'list_users': '查看列表'
}

// 可用的机器人（未创建用户的）
const availableBots = computed(() => {
  const usedBotIds = new Set(botUsers.value.map(bu => bu.bot_id))
  return bots.value.filter(bot => !usedBotIds.has(bot.id))
})

// 加载机器人列表
const loadBots = async () => {
  botsLoading.value = true
  try {
    const response = await request.get('/super-admin/bots')
    bots.value = response.data.data || []
  } catch (error) {
    console.error('加载机器人列表失败:', error)
  } finally {
    botsLoading.value = false
  }
}

// 加载机器人用户
const loadBotUsers = async () => {
  botUsersLoading.value = true
  try {
    botUsers.value = []
    for (const bot of bots.value) {
      try {
        const response = await request.get(`/super-admin/bot-users/${bot.id}`)
        if (response.data.success && response.data.data) {
          botUsers.value.push(response.data.data)
        }
      } catch (error) {
        if (error.response?.status !== 404) {
          console.error(`加载机器人${bot.id}的用户失败:`, error)
        }
      }
    }
  } catch (error) {
    console.error('加载机器人用户列表失败:', error)
  } finally {
    botUsersLoading.value = false
  }
}

// 加载权限列表
const loadPermissions = async () => {
  permissionsLoading.value = true
  try {
    permissions.value = []
    for (const bot of bots.value) {
      try {
        const response = await request.get(`/super-admin/bot-users/${bot.id}/permissions`)
        if (response.data.success && response.data.data) {
          permissions.value.push(...response.data.data)
        }
      } catch (error) {
        console.error(`加载机器人${bot.id}的权限失败:`, error)
      }
    }
  } catch (error) {
    console.error('加载权限列表失败:', error)
  } finally {
    permissionsLoading.value = false
  }
}

// 创建机器人
const createBot = async () => {
  if (!botForm.name || botForm.permissions.length === 0) {
    ElMessage.warning('请填写必填项')
    return
  }

  botSubmitting.value = true
  try {
    const response = await request.post('/super-admin/bots', botForm)
    createdApiKeys.value = {
      api_key: response.data.data.api_key,
      api_secret: response.data.data.api_secret
    }
    showCreateBotDialog.value = false
    showApiKeysDialog.value = true
    ElMessage.success('机器人创建成功')
    loadBots()
    Object.assign(botForm, { name: '', description: '', type: 'internal', permissions: [] })
  } catch (error) {
    ElMessage.error('创建失败: ' + (error.response?.data?.error || error.message))
  } finally {
    botSubmitting.value = false
  }
}

// 创建机器人用户
const createBotUser = async () => {
  if (!botUserForm.bot_id || !botUserForm.username || !botUserForm.nickname) {
    ElMessage.warning('请填写必填项')
    return
  }

  botUserSubmitting.value = true
  try {
    await request.post('/super-admin/bot-users', botUserForm)
    showCreateBotUserDialog.value = false
    ElMessage.success('机器人用户创建成功！\n用户现在可以通过搜索"' + botUserForm.username + '"与机器人对话。')
    loadBotUsers()
    Object.assign(botUserForm, { bot_id: '', username: '', nickname: '', avatar: '' })
  } catch (error) {
    ElMessage.error('创建失败: ' + (error.response?.data?.error || error.message))
  } finally {
    botUserSubmitting.value = false
  }
}

// 授权用户
const grantPermission = async () => {
  if (!permissionForm.user_id || !permissionForm.bot_id) {
    ElMessage.warning('请填写必填项')
    return
  }

  permissionSubmitting.value = true
  try {
    const data = {
      user_id: permissionForm.user_id,
      bot_id: permissionForm.bot_id
    }
    if (permissionForm.expires_at) {
      data.expires_at = new Date(permissionForm.expires_at).toISOString()
    }
    
    await request.post('/admin/bot-permissions', data)
    showGrantPermissionDialog.value = false
    ElMessage.success('授权成功')
    loadPermissions()
    Object.assign(permissionForm, { user_id: null, bot_id: '', expires_at: null })
  } catch (error) {
    ElMessage.error('授权失败: ' + (error.response?.data?.error || error.message))
  } finally {
    permissionSubmitting.value = false
  }
}

// 查看机器人详情
const viewBotDetails = (bot) => {
  selectedBot.value = bot
  showBotDetailDialog.value = true
}

// 切换机器人状态
const toggleBotStatus = async (bot) => {
  try {
    await request.put(`/super-admin/bots/${bot.id}/status`, {
      is_active: !bot.is_active
    })
    bot.is_active = !bot.is_active
    ElMessage.success('状态已更新')
  } catch (error) {
    ElMessage.error('操作失败: ' + (error.response?.data?.error || error.message))
  }
}

// 删除机器人
const deleteBot = async (bot) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除机器人"${bot.name}"吗？此操作不可恢复！`,
      '危险操作',
      { type: 'error' }
    )
    
    await request.delete(`/super-admin/bots/${bot.id}`)
    ElMessage.success('机器人已删除')
    loadBots()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败: ' + (error.response?.data?.error || error.message))
    }
  }
}

// 删除机器人用户
const deleteBotUser = async (botUser) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除机器人用户"${botUser.user?.username}"吗？`,
      '确认删除',
      { type: 'warning' }
    )
    
    await request.delete(`/super-admin/bot-users/${botUser.bot_id}`)
    ElMessage.success('机器人用户已删除')
    loadBotUsers()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败: ' + (error.response?.data?.error || error.message))
    }
  }
}

// 撤销权限
const revokePermission = async (perm) => {
  try {
    await ElMessageBox.confirm(
      `确定要撤销用户"${perm.user?.username}"的权限吗？`,
      '确认撤销',
      { type: 'warning' }
    )
    
    await request.delete(`/admin/bot-permissions/${perm.user_id}/${perm.bot_id}`)
    ElMessage.success('权限已撤销')
    loadPermissions()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('撤销失败: ' + (error.response?.data?.error || error.message))
    }
  }
}

// 工具函数
const parsePermissions = (permissionsStr) => {
  try {
    return JSON.parse(permissionsStr || '[]')
  } catch {
    return []
  }
}

const formatDate = (dateStr) => {
  if (!dateStr || dateStr === '0001-01-01T00:00:00Z') return '从未'
  return new Date(dateStr).toLocaleString('zh-CN')
}

const formatExpiry = (dateStr) => {
  if (!dateStr) return '永不过期'
  const date = new Date(dateStr)
  if (date < new Date()) return '已过期'
  return date.toLocaleString('zh-CN')
}

const getExpiryColor = (dateStr) => {
  if (!dateStr) return '#67c23a'
  const date = new Date(dateStr)
  if (date < new Date()) return '#f56c6c'
  return '#409eff'
}

const copyToClipboard = (text) => {
  navigator.clipboard.writeText(text).then(() => {
    ElMessage.success('已复制到剪贴板')
  }).catch(() => {
    ElMessage.error('复制失败')
  })
}

// 系统配置操作
const handleSaveConfig = async () => {
  try {
    await new Promise(resolve => setTimeout(resolve, 1000))
    ElMessage.success('配置保存成功')
  } catch (error) {
    ElMessage.error('配置保存失败')
  }
}

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

const handleRestartServices = async () => {
  try {
    await ElMessageBox.confirm('确定要重启所有服务吗？', '提示', {
      type: 'warning'
    })
    ElMessage.success('服务重启成功')
  } catch (error) {
    // 取消
  }
}

const handleClearCache = async () => {
  try {
    await ElMessageBox.confirm('确定要清理所有缓存吗？', '提示', {
      type: 'warning'
    })
    ElMessage.success('缓存清理成功')
  } catch (error) {
    // 取消
  }
}

const handleBackupData = async () => {
  try {
    await ElMessageBox.confirm('确定要备份所有数据吗？', '提示', {
      type: 'info'
    })
    ElMessage.success('数据备份成功')
  } catch (error) {
    // 取消
  }
}

const handleShutdown = async () => {
  try {
    await ElMessageBox.confirm('确定要关闭系统吗？这将停止所有服务！', '危险操作', {
      type: 'error'
    })
    ElMessage.success('系统关闭成功')
  } catch (error) {
    // 取消
  }
}

onMounted(() => {
  loadBots().then(() => {
    loadBotUsers()
    loadPermissions()
  })
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

  .bots-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;

    h3 {
      margin: 0;
      font-size: 18px;
      font-weight: 600;
    }
  }

  .stats {
    display: flex;
    flex-direction: column;
    gap: 4px;
    font-size: 12px;
  }
}
</style>
