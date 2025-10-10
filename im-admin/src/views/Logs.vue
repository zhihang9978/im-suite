<template>
  <div class="logs-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>日志管理</span>
          <el-button type="primary" @click="handleRefresh">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>
      
      <!-- 搜索栏 -->
      <div class="search-bar">
        <el-form :model="searchForm" inline>
          <el-form-item label="日志级别">
            <el-select v-model="searchForm.level" placeholder="请选择级别" clearable>
              <el-option label="DEBUG" value="debug" />
              <el-option label="INFO" value="info" />
              <el-option label="WARN" value="warn" />
              <el-option label="ERROR" value="error" />
            </el-select>
          </el-form-item>
          <el-form-item label="模块">
            <el-select v-model="searchForm.module" placeholder="请选择模块" clearable>
              <el-option label="认证" value="auth" />
              <el-option label="用户" value="user" />
              <el-option label="消息" value="message" />
              <el-option label="系统" value="system" />
            </el-select>
          </el-form-item>
          <el-form-item label="关键词">
            <el-input v-model="searchForm.keyword" placeholder="请输入关键词" clearable />
          </el-form-item>
          <el-form-item label="时间范围">
            <el-date-picker
              v-model="searchForm.dateRange"
              type="datetimerange"
              range-separator="至"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              format="YYYY-MM-DD HH:mm:ss"
              value-format="YYYY-MM-DD HH:mm:ss"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">
              <el-icon><Search /></el-icon>
              搜索
            </el-button>
            <el-button @click="handleReset">
              <el-icon><Refresh /></el-icon>
              重置
            </el-button>
          </el-form-item>
        </el-form>
      </div>
      
      <!-- 日志表格 -->
      <el-table :data="logs" v-loading="loading" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="level" label="级别" width="100">
          <template #default="{ row }">
            <el-tag :type="getLevelTag(row.level)">
              {{ row.level.toUpperCase() }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="module" label="模块" width="100" />
        <el-table-column prop="message" label="消息" show-overflow-tooltip />
        <el-table-column prop="user_id" label="用户ID" width="100" />
        <el-table-column prop="ip" label="IP地址" width="120" />
        <el-table-column prop="created_at" label="时间" width="180" />
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleView(row)">
              查看
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>
    
    <!-- 日志详情对话框 -->
    <el-dialog
      v-model="dialogVisible"
      title="日志详情"
      width="800px"
    >
      <div class="log-detail">
        <div class="detail-item">
          <span class="label">ID:</span>
          <span class="value">{{ currentLog.id }}</span>
        </div>
        <div class="detail-item">
          <span class="label">级别:</span>
          <el-tag :type="getLevelTag(currentLog.level)">
            {{ currentLog.level.toUpperCase() }}
          </el-tag>
        </div>
        <div class="detail-item">
          <span class="label">模块:</span>
          <span class="value">{{ currentLog.module }}</span>
        </div>
        <div class="detail-item">
          <span class="label">消息:</span>
          <span class="value">{{ currentLog.message }}</span>
        </div>
        <div class="detail-item">
          <span class="label">用户ID:</span>
          <span class="value">{{ currentLog.user_id || 'N/A' }}</span>
        </div>
        <div class="detail-item">
          <span class="label">IP地址:</span>
          <span class="value">{{ currentLog.ip || 'N/A' }}</span>
        </div>
        <div class="detail-item">
          <span class="label">时间:</span>
          <span class="value">{{ currentLog.created_at }}</span>
        </div>
        <div class="detail-item">
          <span class="label">详细信息:</span>
          <pre class="value">{{ currentLog.details || '无' }}</pre>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

const loading = ref(false)
const logs = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)

const searchForm = reactive({
  level: '',
  module: '',
  keyword: '',
  dateRange: []
})

const dialogVisible = ref(false)
const currentLog = ref({})

// 获取日志列表
const getLogs = async () => {
  loading.value = true
  try {
    const response = await request.get('/super-admin/logs', {
      params: {
        page: currentPage.value,
        page_size: pageSize.value,
        level: searchForm.level,
        module: searchForm.module,
        keyword: searchForm.keyword,
        start_date: searchForm.dateRange[0],
        end_date: searchForm.dateRange[1]
      }
    })
    
    logs.value = response.data || []
    total.value = response.total || 0
  } catch (error) {
    ElMessage.error('获取日志列表失败: ' + (error.response?.data?.error || error.message))
  } finally {
    loading.value = false
  }
}

// 获取级别标签
const getLevelTag = (level) => {
  const tags = {
    debug: 'info',
    info: 'success',
    warn: 'warning',
    error: 'danger'
  }
  return tags[level] || 'info'
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  getLogs()
}

// 重置
const handleReset = () => {
  Object.assign(searchForm, {
    level: '',
    module: '',
    keyword: '',
    dateRange: []
  })
  handleSearch()
}

// 刷新
const handleRefresh = () => {
  getLogs()
}

// 查看日志
const handleView = (row) => {
  currentLog.value = row
  dialogVisible.value = true
}

// 分页
const handleSizeChange = (val) => {
  pageSize.value = val
  getLogs()
}

const handleCurrentChange = (val) => {
  currentPage.value = val
  getLogs()
}

onMounted(() => {
  getLogs()
})
</script>

<style lang="scss" scoped>
.logs-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  
  .search-bar {
    margin-bottom: 20px;
    padding: 20px;
    background: #f5f5f5;
    border-radius: 6px;
  }
  
  .pagination {
    margin-top: 20px;
    text-align: right;
  }
  
  .log-detail {
    .detail-item {
      display: flex;
      margin-bottom: 15px;
      
      .label {
        font-weight: 500;
        color: #606266;
        width: 100px;
        flex-shrink: 0;
      }
      
      .value {
        color: #303133;
        flex: 1;
        
        pre {
          background: #f5f5f5;
          padding: 10px;
          border-radius: 4px;
          font-size: 12px;
          white-space: pre-wrap;
          word-break: break-all;
        }
      }
    }
  }
}
</style>
