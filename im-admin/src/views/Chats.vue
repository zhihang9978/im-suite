<template>
  <div class="chats-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>聊天管理</span>
        </div>
      </template>
      
      <!-- 搜索栏 -->
      <div class="search-bar">
        <el-form :model="searchForm" inline>
          <el-form-item label="聊天类型">
            <el-select v-model="searchForm.type" placeholder="请选择类型" clearable>
              <el-option label="私聊" value="private" />
              <el-option label="群聊" value="group" />
              <el-option label="频道" value="channel" />
            </el-select>
          </el-form-item>
          <el-form-item label="标题">
            <el-input v-model="searchForm.title" placeholder="请输入标题" clearable />
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
      
      <!-- 聊天表格 -->
      <el-table :data="chats" v-loading="loading" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="getTypeTag(row.type)">
              {{ getTypeText(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" />
        <el-table-column prop="member_count" label="成员数" width="100" />
        <el-table-column prop="last_message" label="最后消息" />
        <el-table-column prop="last_message_at" label="最后消息时间" />
        <el-table-column prop="created_at" label="创建时间" />
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleView(row)">
              查看
            </el-button>
            <el-button type="danger" size="small" @click="handleDelete(row)">
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
const chats = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)

const searchForm = reactive({
  type: '',
  title: ''
})

// 获取聊天列表
const getChats = async () => {
  loading.value = true
  try {
    const response = await request.get('/super-admin/chats', {
      params: {
        page: currentPage.value,
        page_size: pageSize.value,
        type: searchForm.type,
        title: searchForm.title
      }
    })
    
    chats.value = response.data || []
    total.value = response.pagination?.total || 0
  } catch (error) {
    ElMessage.error('获取聊天列表失败: ' + (error.response?.data?.error || error.message))
  } finally {
    loading.value = false
  }
}

// 获取类型标签
const getTypeTag = (type) => {
  const tags = {
    private: 'success',
    group: 'primary',
    channel: 'warning'
  }
  return tags[type] || 'info'
}

// 获取类型文本
const getTypeText = (type) => {
  const texts = {
    private: '私聊',
    group: '群聊',
    channel: '频道'
  }
  return texts[type] || '未知'
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  getChats()
}

// 重置
const handleReset = () => {
  Object.assign(searchForm, {
    type: '',
    title: ''
  })
  handleSearch()
}

// 查看聊天
const handleView = (row) => {
  ElMessage.info(`查看聊天: ${row.title}`)
}

// 删除聊天
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该聊天吗？此操作将删除所有相关消息！', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await request.delete(`/chats/${row.id}`)
    ElMessage.success('删除成功')
    getChats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败: ' + (error.response?.data?.error || error.message))
    }
  }
}

// 分页
const handleSizeChange = (val) => {
  pageSize.value = val
  getChats()
}

const handleCurrentChange = (val) => {
  currentPage.value = val
  getChats()
}

onMounted(() => {
  getChats()
})
</script>

<style lang="scss" scoped>
.chats-page {
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
