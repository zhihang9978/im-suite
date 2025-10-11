<template>
  <div class="contacts-container">
    <el-card>
      <template #header>
        <div class="header">
          <h3>联系人</h3>
          <el-button type="primary" @click="addContactDialogVisible = true">
            添加好友
          </el-button>
        </div>
      </template>
      
      <el-input
        v-model="searchKeyword"
        placeholder="搜索联系人"
        prefix-icon="Search"
        class="search-input"
      />
      
      <div class="contact-list">
        <div
          v-for="contact in filteredContacts"
          :key="contact.id"
          class="contact-card"
        >
          <el-avatar :size="50" :src="contact.avatar">
            {{ contact.nickname?.charAt(0) || 'U' }}
          </el-avatar>
          <div class="contact-info">
            <div class="contact-name">{{ contact.nickname || contact.username }}</div>
            <div class="contact-phone">{{ contact.phone }}</div>
          </div>
          <el-button type="primary" @click="startChat(contact)">
            发消息
          </el-button>
        </div>
        
        <el-empty v-if="contacts.length === 0" description="暂无联系人" />
      </div>
    </el-card>
    
    <!-- 添加好友对话框 -->
    <el-dialog v-model="addContactDialogVisible" title="添加好友" width="400px">
      <el-form :model="addContactForm">
        <el-form-item label="手机号">
          <el-input v-model="addContactForm.phone" placeholder="请输入对方手机号" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="addContactDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAddContact">添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useChatStore } from '@/stores/chat'
import { ElMessage } from 'element-plus'
import { searchUserByPhone } from '@/api/search'

const router = useRouter()
const chatStore = useChatStore()

const searchKeyword = ref('')
const addContactDialogVisible = ref(false)
const addContactForm = ref({ phone: '' })

const contacts = computed(() => chatStore.contacts)

const filteredContacts = computed(() => {
  if (!searchKeyword.value) return contacts.value
  return contacts.value.filter(c => 
    c.nickname?.includes(searchKeyword.value) || 
    c.username?.includes(searchKeyword.value) ||
    c.phone?.includes(searchKeyword.value)
  )
})

const startChat = (contact) => {
  chatStore.switchChat(contact)
  router.push('/')
}

const handleAddContact = async () => {
  if (!addContactForm.value.phone) {
    ElMessage.warning('请输入手机号')
    return
  }

  try {
    // 搜索用户
    const result = await searchUserByPhone(addContactForm.value.phone)
    
    if (result.success && result.data) {
      // 添加到联系人（暂时添加到本地，实际应该调用后端API）
      const newContact = result.data
      if (!contacts.value.find(c => c.id === newContact.id)) {
        contacts.value.push(newContact)
        ElMessage.success('添加好友成功')
      } else {
        ElMessage.info('该用户已在联系人列表中')
      }
      addContactDialogVisible.value = false
      addContactForm.value.phone = ''
    } else {
      ElMessage.warning(result.message || '未找到该用户')
    }
  } catch (error) {
    ElMessage.error('添加好友失败: ' + error.message)
  }
}
</script>

<style scoped>
.contacts-container {
  padding: 20px;
  height: 100vh;
  background: #f5f5f5;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-input {
  margin-bottom: 20px;
}

.contact-list {
  max-height: calc(100vh - 250px);
  overflow-y: auto;
}

.contact-card {
  display: flex;
  align-items: center;
  padding: 15px;
  gap: 15px;
  background: white;
  margin-bottom: 10px;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.contact-info {
  flex: 1;
}

.contact-name {
  font-weight: 500;
  font-size: 15px;
  margin-bottom: 5px;
}

.contact-phone {
  font-size: 13px;
  color: #999;
}
</style>

