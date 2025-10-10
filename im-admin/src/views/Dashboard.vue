<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <!-- 统计卡片 -->
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon users">
              <el-icon><User /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-number">{{ stats.totalUsers }}</div>
              <div class="stat-label">总用户数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon chats">
              <el-icon><ChatDotRound /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-number">{{ stats.totalChats }}</div>
              <div class="stat-label">总聊天数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon messages">
              <el-icon><Message /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-number">{{ stats.totalMessages }}</div>
              <div class="stat-label">总消息数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon online">
              <el-icon><Connection /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-number">{{ stats.onlineUsers }}</div>
              <div class="stat-label">在线用户</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    
    <el-row :gutter="20" style="margin-top: 20px;">
      <!-- 用户增长趋势 -->
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>用户增长趋势</span>
          </template>
          <div ref="userChartRef" style="height: 300px;"></div>
        </el-card>
      </el-col>
      
      <!-- 消息统计 -->
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>消息统计</span>
          </template>
          <div ref="messageChartRef" style="height: 300px;"></div>
        </el-card>
      </el-col>
    </el-row>
    
    <el-row :gutter="20" style="margin-top: 20px;">
      <!-- 最近用户 -->
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>最近注册用户</span>
          </template>
          <el-table :data="recentUsers" style="width: 100%">
            <el-table-column prop="username" label="用户名" />
            <el-table-column prop="phone" label="手机号" />
            <el-table-column prop="created_at" label="注册时间" />
            <el-table-column prop="status" label="状态">
              <template #default="{ row }">
                <el-tag :type="row.status === 'online' ? 'success' : 'info'">
                  {{ row.status === 'online' ? '在线' : '离线' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      
      <!-- 系统状态 -->
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>系统状态</span>
          </template>
          <div class="system-status">
            <div class="status-item">
              <span class="status-label">数据库连接</span>
              <el-tag :type="systemStatus.database ? 'success' : 'danger'">
                {{ systemStatus.database ? '正常' : '异常' }}
              </el-tag>
            </div>
            <div class="status-item">
              <span class="status-label">Redis 连接</span>
              <el-tag :type="systemStatus.redis ? 'success' : 'danger'">
                {{ systemStatus.redis ? '正常' : '异常' }}
              </el-tag>
            </div>
            <div class="status-item">
              <span class="status-label">MinIO 连接</span>
              <el-tag :type="systemStatus.minio ? 'success' : 'danger'">
                {{ systemStatus.minio ? '正常' : '异常' }}
              </el-tag>
            </div>
            <div class="status-item">
              <span class="status-label">系统负载</span>
              <el-tag :type="systemStatus.load < 0.8 ? 'success' : 'warning'">
                {{ (systemStatus.load * 100).toFixed(1) }}%
              </el-tag>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import request from '@/api/request'

const stats = ref({
  totalUsers: 0,
  totalChats: 0,
  totalMessages: 0,
  onlineUsers: 0
})

const recentUsers = ref([])

const systemStatus = ref({
  database: true,
  redis: true,
  minio: true,
  load: 0.65
})

const userChartRef = ref()
const messageChartRef = ref()

const initCharts = () => {
  // 用户增长趋势图
  const userChart = echarts.init(userChartRef.value)
  const userOption = {
    tooltip: { trigger: 'axis' },
    xAxis: {
      type: 'category',
      data: ['1月', '2月', '3月', '4月', '5月', '6月', '7月']
    },
    yAxis: { type: 'value' },
    series: [{
      name: '新增用户',
      type: 'line',
      data: [120, 200, 150, 300, 250, 400, 350],
      smooth: true
    }]
  }
  userChart.setOption(userOption)
  
  // 消息统计图
  const messageChart = echarts.init(messageChartRef.value)
  const messageOption = {
    tooltip: { trigger: 'item' },
    series: [{
      name: '消息类型',
      type: 'pie',
      radius: '50%',
      data: [
        { value: 1048, name: '文本消息' },
        { value: 735, name: '图片消息' },
        { value: 580, name: '语音消息' },
        { value: 484, name: '视频消息' },
        { value: 300, name: '文件消息' }
      ]
    }]
  }
  messageChart.setOption(messageOption)
}

// 加载仪表盘数据
const loadDashboardData = async () => {
  try {
    // 加载系统统计
    const statsResponse = await request.get('/super-admin/stats')
    stats.value = {
      totalUsers: statsResponse.data.total_users || 0,
      totalChats: statsResponse.data.total_chats || 0,
      totalMessages: statsResponse.data.total_messages || 0,
      onlineUsers: statsResponse.data.online_users || 0
    }
    
    // 加载最近用户（取前5个）
    const usersResponse = await request.get('/super-admin/users', {
      params: { page: 1, page_size: 5 }
    })
    recentUsers.value = (usersResponse.data || []).map(user => ({
      username: user.username,
      phone: user.phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2'), // 手机号脱敏
      created_at: user.created_at,
      status: user.online ? 'online' : 'offline'
    }))
  } catch (error) {
    ElMessage.error('加载数据失败: ' + (error.response?.data?.error || error.message))
  }
}

onMounted(() => {
  loadDashboardData()
  nextTick(() => {
    initCharts()
  })
})
</script>

<style lang="scss" scoped>
.dashboard {
  .stat-card {
    .stat-content {
      display: flex;
      align-items: center;
      
      .stat-icon {
        width: 60px;
        height: 60px;
        border-radius: 10px;
        display: flex;
        align-items: center;
        justify-content: center;
        margin-right: 15px;
        font-size: 24px;
        color: white;
        
        &.users { background: #409eff; }
        &.chats { background: #67c23a; }
        &.messages { background: #e6a23c; }
        &.online { background: #f56c6c; }
      }
      
      .stat-info {
        .stat-number {
          font-size: 28px;
          font-weight: bold;
          color: #303133;
        }
        
        .stat-label {
          font-size: 14px;
          color: #909399;
          margin-top: 5px;
        }
      }
    }
  }
  
  .system-status {
    .status-item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 10px 0;
      border-bottom: 1px solid #ebeef5;
      
      &:last-child {
        border-bottom: none;
      }
      
      .status-label {
        font-size: 14px;
        color: #606266;
      }
    }
  }
}
</style>
