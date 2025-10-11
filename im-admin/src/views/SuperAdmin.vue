<template>
  <div class="super-admin-dashboard">
    <el-row :gutter="20">
      <el-col :span="24">
        <div class="page-header">
          <h1>
            <el-icon><Promotion /></el-icon>
            超级管理后台
          </h1>
          <el-button type="primary" :icon="Refresh" @click="refreshAll">
            刷新数据
          </el-button>
        </div>
      </el-col>
    </el-row>

    <!-- 系统统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <el-icon class="stat-icon primary"><User /></el-icon>
            <div class="stat-content">
              <div class="stat-value">{{ formatNumber(stats.total_users) }}</div>
              <div class="stat-label">总用户数</div>
              <div class="stat-sub">
                在线: {{ formatNumber(stats.online_users) }}
              </div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <el-icon class="stat-icon success"><ChatDotRound /></el-icon>
            <div class="stat-content">
              <div class="stat-value">{{ formatNumber(stats.total_messages) }}</div>
              <div class="stat-label">总消息数</div>
              <div class="stat-sub">
                今日: {{ formatNumber(stats.today_messages) }}
              </div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <el-icon class="stat-icon warning"><Folder /></el-icon>
            <div class="stat-content">
              <div class="stat-value">{{ formatBytes(stats.storage_used) }}</div>
              <div class="stat-label">存储使用</div>
              <div class="stat-sub">
                数据库: {{ formatBytes(stats.database_size) }}
              </div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <el-icon class="stat-icon danger"><Cpu /></el-icon>
            <div class="stat-content">
              <div class="stat-value">{{ stats.cpu_usage?.toFixed(1) }}%</div>
              <div class="stat-label">CPU使用率</div>
              <div class="stat-sub">
                内存: {{ stats.memory_usage?.toFixed(1) }}%
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 功能标签页 -->
    <el-card class="main-card">
      <el-tabs v-model="activeTab" type="border-card">
        <!-- 在线用户 -->
        <el-tab-pane label="在线用户" name="online">
          <el-table :data="onlineUsers" style="width: 100%" v-loading="loading">
            <el-table-column prop="user_id" label="ID" width="80" />
            <el-table-column label="用户" width="200">
              <template #default="{ row }">
                <div class="user-info">
                  <el-avatar :src="row.avatar" :size="32">{{ row.username[0] }}</el-avatar>
                  <div class="user-details">
                    <div class="username">{{ row.username }}</div>
                    <div class="nickname">{{ row.nickname }}</div>
                  </div>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag
                  :type="row.online_status === 'online' ? 'success' : 'info'"
                  size="small"
                >
                  {{ row.online_status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="ip_address" label="IP地址" width="150" />
            <el-table-column prop="device" label="设备" width="200" />
            <el-table-column label="登录时间" width="180">
              <template #default="{ row }">
                {{ formatTime(row.login_time) }}
              </template>
            </el-table-column>
            <el-table-column prop="session_count" label="会话数" width="80" />
            <el-table-column label="操作" width="250" fixed="right">
              <template #default="{ row }">
                <el-button
                  type="info"
                  size="small"
                  :icon="View"
                  @click="viewUserAnalysis(row.user_id)"
                >
                  详情
                </el-button>
                <el-button
                  type="warning"
                  size="small"
                  :icon="SwitchButton"
                  @click="forceLogout(row)"
                >
                  下线
                </el-button>
                <el-button
                  type="danger"
                  size="small"
                  :icon="CircleClose"
                  @click="openBanDialog(row)"
                >
                  封禁
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 用户分析 -->
        <el-tab-pane label="用户分析" name="analysis">
          <div v-if="userAnalysis">
            <el-row :gutter="20">
              <el-col :span="8">
                <el-card>
                  <div class="risk-score">
                    <h3>风险评分</h3>
                    <el-progress
                      type="circle"
                      :percentage="userAnalysis.risk_score"
                      :color="getRiskColor(userAnalysis.risk_score)"
                      :width="150"
                    />
                    <el-alert
                      v-if="userAnalysis.is_suspicious"
                      type="warning"
                      title="可疑用户"
                      description="该用户行为异常，建议重点关注"
                      :closable="false"
                      style="margin-top: 20px"
                    />
                  </div>
                </el-card>
              </el-col>

              <el-col :span="16">
                <el-card>
                  <h3>用户行为统计</h3>
                  <el-row :gutter="20" style="margin-top: 20px">
                    <el-col :span="6">
                      <div class="stat-item">
                        <div class="stat-value">{{ formatNumber(userAnalysis.message_count) }}</div>
                        <div class="stat-label">消息数量</div>
                      </div>
                    </el-col>
                    <el-col :span="6">
                      <div class="stat-item">
                        <div class="stat-value">{{ userAnalysis.group_count }}</div>
                        <div class="stat-label">群组数量</div>
                      </div>
                    </el-col>
                    <el-col :span="6">
                      <div class="stat-item">
                        <div class="stat-value">{{ userAnalysis.file_upload_count }}</div>
                        <div class="stat-label">文件上传</div>
                      </div>
                    </el-col>
                    <el-col :span="6">
                      <div class="stat-item">
                        <div class="stat-value">{{ formatDuration(userAnalysis.online_time) }}</div>
                        <div class="stat-label">在线时长</div>
                      </div>
                    </el-col>
                  </el-row>

                  <el-row :gutter="20" style="margin-top: 20px">
                    <el-col :span="6">
                      <div class="stat-item">
                        <div class="stat-value danger">{{ userAnalysis.violation_count }}</div>
                        <div class="stat-label">违规次数</div>
                      </div>
                    </el-col>
                    <el-col :span="6">
                      <div class="stat-item">
                        <div class="stat-value warning">{{ userAnalysis.reported_count }}</div>
                        <div class="stat-label">被举报次数</div>
                      </div>
                    </el-col>
                    <el-col :span="6">
                      <div class="stat-item">
                        <el-tag :type="userAnalysis.is_blacklisted ? 'danger' : 'success'">
                          {{ userAnalysis.is_blacklisted ? '已拉黑' : '正常' }}
                        </el-tag>
                        <div class="stat-label">黑名单状态</div>
                      </div>
                    </el-col>
                    <el-col :span="6">
                      <div class="stat-item">
                        <div class="stat-value">{{ formatTime(userAnalysis.last_login_time) }}</div>
                        <div class="stat-label">最后登录</div>
                      </div>
                    </el-col>
                  </el-row>
                </el-card>
              </el-col>
            </el-row>
          </div>
          <el-empty v-else description="请选择用户查看分析" />
        </el-tab-pane>

        <!-- 内容审核 -->
        <el-tab-pane label="内容审核" name="moderation">
          <el-table :data="moderationQueue" v-loading="loading">
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="content_type" label="类型" width="100" />
            <el-table-column prop="username" label="用户" width="150" />
            <el-table-column prop="content" label="内容" show-overflow-tooltip />
            <el-table-column label="违规类型" width="120">
              <template #default="{ row }">
                <el-tag size="small">{{ row.violation_type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="严重程度" width="100">
              <template #default="{ row }">
                <el-tag
                  :type="getSeverityType(row.severity)"
                  size="small"
                >
                  {{ row.severity }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="250" fixed="right">
              <template #default="{ row }">
                <el-button
                  type="success"
                  size="small"
                  @click="moderateContent(row.id, 'approve')"
                >
                  通过
                </el-button>
                <el-button
                  type="warning"
                  size="small"
                  @click="moderateContent(row.id, 'warn')"
                >
                  警告
                </el-button>
                <el-button
                  type="danger"
                  size="small"
                  @click="moderateContent(row.id, 'delete')"
                >
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 系统日志 -->
        <el-tab-pane label="系统日志" name="logs">
          <el-table :data="systemLogs" v-loading="loading">
            <el-table-column label="级别" width="100">
              <template #default="{ row }">
                <el-tag :type="getLogType(row.level)" size="small">
                  {{ row.level }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="message" label="消息" show-overflow-tooltip />
            <el-table-column prop="user_id" label="用户ID" width="100" />
            <el-table-column prop="ip_address" label="IP地址" width="150" />
            <el-table-column label="时间" width="180">
              <template #default="{ row }">
                {{ formatTime(row.timestamp) }}
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 封禁用户对话框 -->
    <el-dialog v-model="banDialogVisible" title="封禁用户" width="500px">
      <el-form :model="banForm" label-width="100px">
        <el-form-item label="用户">
          <el-input :value="selectedUser?.username" disabled />
        </el-form-item>
        <el-form-item label="封禁时长">
          <el-select v-model="banForm.duration" placeholder="请选择封禁时长">
            <el-option label="1小时" :value="1" />
            <el-option label="24小时" :value="24" />
            <el-option label="7天" :value="168" />
            <el-option label="30天" :value="720" />
            <el-option label="永久" :value="8760" />
          </el-select>
        </el-form-item>
        <el-form-item label="封禁原因">
          <el-input
            v-model="banForm.reason"
            type="textarea"
            :rows="3"
            placeholder="请输入封禁原因"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="banDialogVisible = false">取消</el-button>
        <el-button type="danger" @click="confirmBan">确认封禁</el-button>
      </template>
    </el-dialog>

    <!-- 广播消息对话框 -->
    <el-dialog v-model="broadcastDialogVisible" title="广播系统消息" width="600px">
      <el-form :model="broadcastForm" label-width="100px">
        <el-form-item label="目标类型">
          <el-radio-group v-model="broadcastForm.target_type">
            <el-radio label="all">所有用户</el-radio>
            <el-radio label="users">指定用户</el-radio>
            <el-radio label="groups">指定群组</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item
          v-if="broadcastForm.target_type !== 'all'"
          label="目标ID"
        >
          <el-input
            v-model="broadcastForm.target_ids_str"
            placeholder="多个ID用逗号分隔，例如: 1,2,3"
          />
        </el-form-item>
        <el-form-item label="消息内容">
          <el-input
            v-model="broadcastForm.message"
            type="textarea"
            :rows="4"
            placeholder="请输入要广播的消息内容"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="broadcastDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmBroadcast">确认广播</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import {
  Refresh,
  Promotion,
  User,
  ChatDotRound,
  Folder,
  Cpu,
  View,
  SwitchButton,
  CircleClose,
} from '@element-plus/icons-vue';
import request from '@/api/request';

const activeTab = ref('online');
const loading = ref(false);
const banDialogVisible = ref(false);
const broadcastDialogVisible = ref(false);

const stats = ref({
  total_users: 0,
  online_users: 0,
  total_messages: 0,
  today_messages: 0,
  total_groups: 0,
  active_groups: 0,
  total_files: 0,
  storage_used: 0,
  database_size: 0,
  cpu_usage: 0,
  memory_usage: 0,
});

const onlineUsers = ref<any[]>([]);
const userAnalysis = ref<any>(null);
const moderationQueue = ref<any[]>([]);
const systemLogs = ref<any[]>([]);
const selectedUser = ref<any>(null);

const banForm = ref({
  duration: 24,
  reason: '',
});

const broadcastForm = ref({
  target_type: 'all',
  target_ids_str: '',
  message: '',
});

onMounted(() => {
  refreshAll();
});

const refreshAll = async () => {
  loading.value = true;
  await Promise.all([
    loadSystemStats(),
    loadOnlineUsers(),
    loadModerationQueue(),
    loadSystemLogs(),
  ]);
  loading.value = false;
};

const loadSystemStats = async () => {
  try {
    const data = await request.get('/super-admin/stats');
    stats.value = data.data;
  } catch (err) {
    console.error('加载系统统计失败:', err);
  }
};

const loadOnlineUsers = async () => {
  try {
    const data = await request.get('/super-admin/users/online');
    onlineUsers.value = data.data || [];
  } catch (err) {
    console.error('加载在线用户失败:', err);
  }
};

const loadModerationQueue = async () => {
  try {
    // 内容审核队列（如果后端未实现，跳过）
    // const data = await request.get('/moderation/reports/pending');
    // moderationQueue.value = data.data || [];
    moderationQueue.value = []; // 临时：等待后端实现
  } catch (err) {
    console.error('加载审核队列失败:', err);
  }
};

const loadSystemLogs = async () => {
  try {
    const data = await request.get('/super-admin/logs');
    systemLogs.value = data.data || [];
  } catch (err) {
    console.error('加载系统日志失败:', err);
  }
};

const viewUserAnalysis = async (userId: number) => {
  try {
    const data = await request.get(`/super-admin/users/${userId}/analysis`);
    userAnalysis.value = data.data;
    activeTab.value = 'analysis';
  } catch (err) {
    console.error('加载用户分析失败:', err);
  }
};

const forceLogout = async (user: any) => {
  try {
    await ElMessageBox.confirm('确定要强制该用户下线吗？', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    });

    await request.post(`/super-admin/users/${user.user_id}/logout`);

    ElMessage.success('用户已强制下线');
    loadOnlineUsers();
  } catch (err) {
    if (err !== 'cancel') {
      console.error('强制下线失败:', err);
    }
  }
};

const openBanDialog = (user: any) => {
  selectedUser.value = user;
  banDialogVisible.value = true;
};

const confirmBan = async () => {
  try {
    await request.post(`/super-admin/users/${selectedUser.value.user_id}/ban`, banForm.value);

    ElMessage.success('用户已封禁');
    banDialogVisible.value = false;
    banForm.value.reason = '';
    loadOnlineUsers();
  } catch (err) {
    console.error('封禁用户失败:', err);
  }
};

const confirmBroadcast = async () => {
  try {
    const targetIds = broadcastForm.value.target_ids_str
      ? broadcastForm.value.target_ids_str.split(',').map(id => parseInt(id.trim()))
      : [];

    await request.post('/super-admin/broadcast', {
      message: broadcastForm.value.message,
    });

    ElMessage.success('系统消息已广播');
    broadcastDialogVisible.value = false;
    broadcastForm.value.message = '';
  } catch (err) {
    console.error('广播消息失败:', err);
  }
};

const moderateContent = async (contentId: number, action: string) => {
  try {
    const reason = await ElMessageBox.prompt('请输入操作原因', '内容审核', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputPattern: /.+/,
      inputErrorMessage: '原因不能为空',
    });

    // 使用正确的审核API路径
    await request.post(`/moderation/reports/${contentId}/handle`, {
      action,
      reason: reason.value,
    });

    ElMessage.success('审核完成');
    loadModerationQueue();
  } catch (err) {
    if (err !== 'cancel') {
      console.error('审核失败:', err);
    }
  }
};

const formatNumber = (num: number) => {
  return num?.toLocaleString() || '0';
};

const formatBytes = (bytes: number) => {
  if (!bytes) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

const formatTime = (time: string) => {
  return new Date(time).toLocaleString('zh-CN');
};

const formatDuration = (seconds: number) => {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return `${hours}h${minutes}m`;
};

const getRiskColor = (score: number) => {
  if (score >= 80) return '#f56c6c';
  if (score >= 60) return '#e6a23c';
  if (score >= 40) return '#409eff';
  return '#67c23a';
};

const getSeverityType = (severity: string) => {
  const map: any = {
    critical: 'danger',
    high: 'danger',
    medium: 'warning',
    low: 'info',
  };
  return map[severity] || 'info';
};

const getLogType = (level: string) => {
  const map: any = {
    error: 'danger',
    warning: 'warning',
    info: 'info',
    debug: '',
  };
  return map[level] || '';
};
</script>

<style scoped lang="scss">
.super-admin-dashboard {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;

    h1 {
      display: flex;
      align-items: center;
      gap: 10px;
      margin: 0;
    }
  }

  .stats-row {
    margin-bottom: 20px;
  }

  .stat-card {
    display: flex;
    align-items: center;
    gap: 15px;

    .stat-icon {
      font-size: 48px;
      
      &.primary { color: #409eff; }
      &.success { color: #67c23a; }
      &.warning { color: #e6a23c; }
      &.danger { color: #f56c6c; }
    }

    .stat-content {
      flex: 1;

      .stat-value {
        font-size: 28px;
        font-weight: bold;
        line-height: 1;
      }

      .stat-label {
        font-size: 14px;
        color: #909399;
        margin-top: 5px;
      }

      .stat-sub {
        font-size: 12px;
        color: #c0c4cc;
        margin-top: 3px;
      }
    }
  }

  .main-card {
    margin-top: 20px;
  }

  .user-info {
    display: flex;
    align-items: center;
    gap: 10px;

    .user-details {
      .username {
        font-weight: 500;
      }

      .nickname {
        font-size: 12px;
        color: #909399;
      }
    }
  }

  .risk-score {
    text-align: center;
    padding: 20px;

    h3 {
      margin-bottom: 20px;
    }
  }

  .stat-item {
    text-align: center;
    padding: 10px;

    .stat-value {
      font-size: 24px;
      font-weight: bold;
      line-height: 1;

      &.danger { color: #f56c6c; }
      &.warning { color: #e6a23c; }
    }

    .stat-label {
      font-size: 12px;
      color: #909399;
      margin-top: 8px;
    }
  }
}
</style>
