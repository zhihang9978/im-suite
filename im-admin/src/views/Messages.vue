<template>
  <div class="messages-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>消息管理</span>
        </div>
      </template>
      
      <!-- 搜索栏 -->
      <div class="search-bar">
        <el-form :model="searchForm" inline>
          <el-form-item label="消息类型">
            <el-select v-model="searchForm.type" placeholder="请选择类型" clearable>
              <el-option label="文本" value="text" />
              <el-option label="图片" value="image" />
              <el-option label="语音" value="audio" />
              <el-option label="视频" value="video" />
              <el-option label="文件" value="file" />
            </el-select>
          </el-form-item>
          <el-form-item label="发送者">
            <el-input v-model="searchForm.sender" placeholder="请输入发送者" clearable />
          </el-form-item>
          <el-form-item label="内容">
            <el-input v-model="searchForm.content" placeholder="请输入消息内容" clearable />
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
      
      <!-- 消息表格 -->
      <el-table :data="messages" v-loading="loading" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="getTypeTag(row.type)">
              {{ getTypeText(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sender" label="发送者" />
        <el-table-column prop="content" label="内容" show-overflow-tooltip />
        <el-table-column prop="chat_id" label="聊天ID" width="100" />
        <el-table-column prop="created_at" label="发送时间" />
        <el-table-column prop="is_deleted" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_deleted ? 'danger' : 'success'">
              {{ row.is_deleted ? '已删除' : '正常' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleView(row)">
              查看
            </el-button>
            <el-button type="danger" size="small" @click="handleDelete(row)" v-if="!row.is_deleted">
              删除
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
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

const loading = ref(false)
const messages = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)

const searchForm = reactive({
  type: '',
  sender: '',
  content: ''
})

// 获取消息列表
const getMessages = async () => {
  loading.value = true
  try {
    const response = await request.get('/super-admin/messages', {
      params: {
        page: currentPage.value,
        page_size: pageSize.value,
        type: searchForm.type,
        sender: searchForm.sender,
        content: searchForm.content
      }
    })
    
    messages.value = response.data || []
    total.value = response.pagination?.total || 0
  } catch (error) {
    ElMessage.error('获取消息列表失败: ' + (error.response?.data?.error || error.message))
  } finally {
    loading.value = false
  }
}

// 获取类型标签
const getTypeTag = (type) => {
  const tags = {
    text: 'success',
    image: 'primary',
    audio: 'warning',
    video: 'danger',
    file: 'info'
  }
  return tags[type] || 'info'
}

// 获取类型文本
const getTypeText = (type) => {
  const texts = {
    text: '文本',
    image: '图片',
    audio: '语音',
    video: '视频',
    file: '文件'
  }
  return texts[type] || '未知'
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  getMessages()
}

// 重置
const handleReset = () => {
  Object.assign(searchForm, {
    type: '',
    sender: '',
    content: ''
  })
  handleSearch()
}

// 查看消息
const handleView = (row) => {
  ElMessage.info(`查看消息: ${row.content}`)
}

// 删除消息
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该消息吗？此操作不可撤销！', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await request.delete(`/messages/${row.id}`)
    ElMessage.success('删除成功')
    getMessages()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败: ' + (error.response?.data?.error || error.message))
    }
  }
}

// 分页
const handleSizeChange = (val) => {
  pageSize.value = val
  getMessages()
}

const handleCurrentChange = (val) => {
  currentPage.value = val
  getMessages()
}

onMounted(() => {
  getMessages()
})
</script>

<style lang="scss" scoped>
.messages-page {
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
}
</style>
