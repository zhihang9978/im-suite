<template>
  <div class="chat-container">
    <!-- 左侧联系人列表 -->
    <div class="sidebar">
      <div class="sidebar-header">
        <el-avatar :size="40" :src="userStore.user?.avatar">
          {{ userStore.user?.nickname?.charAt(0) || 'U' }}
        </el-avatar>
        <span class="username">{{ userStore.user?.nickname || '用户' }}</span>
        <el-button @click="handleLogout" link>
          <el-icon><SwitchButton /></el-icon>
        </el-button>
      </div>
      
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
          class="contact-item"
          :class="{ active: currentChat?.id === contact.id }"
          @click="selectContact(contact)"
        >
          <el-avatar :size="45" :src="contact.avatar">
            {{ contact.nickname?.charAt(0) || 'U' }}
          </el-avatar>
          <div class="contact-info">
            <div class="contact-name">{{ contact.nickname || contact.username }}</div>
            <div class="contact-last-message">{{ contact.lastMessage || '暂无消息' }}</div>
          </div>
          <div class="contact-time">
            {{ contact.lastTime || '' }}
          </div>
        </div>
        
        <el-empty v-if="contacts.length === 0" description="暂无联系人" />
      </div>
    </div>
    
    <!-- 右侧聊天区域 -->
    <div class="chat-area">
      <div v-if="currentChat" class="chat-content">
        <!-- 聊天头部 -->
        <div class="chat-header">
          <el-avatar :size="40" :src="currentChat.avatar">
            {{ currentChat.nickname?.charAt(0) || 'U' }}
          </el-avatar>
          <div class="chat-info">
            <div class="chat-name">{{ currentChat.nickname || currentChat.username }}</div>
            <div class="chat-status">{{ currentChat.online ? '在线' : '离线' }}</div>
          </div>
        </div>
        
        <!-- 消息列表 -->
        <div class="message-list" ref="messageListRef">
          <div
            v-for="message in messages"
            :key="message.id"
            class="message-item"
            :class="{ self: message.sender_id === userStore.user?.id }"
          >
            <el-avatar :size="35" :src="message.sender?.avatar">
              {{ message.sender?.nickname?.charAt(0) || 'U' }}
            </el-avatar>
            <div class="message-bubble">
              <div class="message-content">{{ message.content }}</div>
              <div class="message-time">{{ formatTime(message.created_at) }}</div>
            </div>
          </div>
          
          <el-empty v-if="messages.length === 0" description="暂无消息，开始聊天吧" />
        </div>
        
        <!-- 输入区域 -->
        <div class="input-area">
          <FileUpload
            v-if="currentChat"
            :receiver-id="currentChat.id"
            @uploaded="handleFileUploaded"
          />
          
          <el-input
            v-model="messageInput"
            placeholder="输入消息..."
            @keyup.enter="sendMessage"
            class="message-input"
          />
          
          <el-button type="primary" @click="sendMessage" :loading="sending">
            发送
          </el-button>
        </div>
      </div>
      
      <!-- 未选择聊天 -->
      <div v-else class="no-chat">
        <el-empty description="选择一个联系人开始聊天" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useChatStore } from '@/stores/chat'
import { useMessageStore } from '@/stores/message'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'
import MessageBubble from '@/components/MessageBubble.vue'
import FileUpload from '@/components/FileUpload.vue'

const router = useRouter()
const userStore = useUserStore()
const chatStore = useChatStore()
const messageStore = useMessageStore()

const searchKeyword = ref('')
const currentChat = ref(null)
const messageInput = ref('')
const sending = ref(false)
const messageListRef = ref(null)

const messages = computed(() => chatStore.messages)
const contacts = computed(() => chatStore.contacts)

const filteredContacts = computed(() => {
  if (!searchKeyword.value) return contacts.value
  return contacts.value.filter(c => 
    c.nickname?.includes(searchKeyword.value) || 
    c.username?.includes(searchKeyword.value)
  )
})

// 选择联系人
const selectContact = (contact) => {
  currentChat.value = contact
  chatStore.switchChat(contact)
  nextTick(() => {
    scrollToBottom()
  })
}

// 发送消息
const sendMessage = async () => {
  if (!messageInput.value.trim() || !currentChat.value) return
  
  sending.value = true
  const result = await chatStore.sendMessage(messageInput.value, currentChat.value.id)
  sending.value = false
  
  if (result.success) {
    messageInput.value = ''
    nextTick(() => {
      scrollToBottom()
    })
  }
}

// 处理文件上传完成
const handleFileUploaded = (message) => {
  messages.value.push(message)
  nextTick(() => scrollToBottom())
}

// 登出
const handleLogout = async () => {
  await userStore.logout()
  router.push('/login')
}

// 格式化时间
const formatTime = (time) => {
  return dayjs(time).format('HH:mm')
}

// 滚动到底部
const scrollToBottom = () => {
  if (messageListRef.value) {
    messageListRef.value.scrollTop = messageListRef.value.scrollHeight
  }
}

// 初始化
onMounted(() => {
  chatStore.initWebSocket()
  chatStore.getContacts()
})
</script>

<style scoped>
.chat-container {
  display: flex;
  width: 100%;
  height: 100vh;
}

.sidebar {
  width: 300px;
  border-right: 1px solid #e0e0e0;
  display: flex;
  flex-direction: column;
  background: #f5f5f5;
}

.sidebar-header {
  display: flex;
  align-items: center;
  padding: 15px;
  gap: 10px;
  background: white;
  border-bottom: 1px solid #e0e0e0;
}

.username {
  flex: 1;
  font-weight: 500;
}

.search-input {
  margin: 10px;
}

.contact-list {
  flex: 1;
  overflow-y: auto;
}

.contact-item {
  display: flex;
  align-items: center;
  padding: 12px 15px;
  cursor: pointer;
  gap: 10px;
  background: white;
  margin-bottom: 1px;
  transition: background 0.2s;
}

.contact-item:hover {
  background: #f0f0f0;
}

.contact-item.active {
  background: #e6f7ff;
}

.contact-info {
  flex: 1;
  min-width: 0;
}

.contact-name {
  font-weight: 500;
  font-size: 14px;
  color: #333;
}

.contact-last-message {
  font-size: 12px;
  color: #999;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.contact-time {
  font-size: 12px;
  color: #999;
}

.chat-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #fff;
}

.chat-header {
  display: flex;
  align-items: center;
  padding: 15px;
  gap: 10px;
  border-bottom: 1px solid #e0e0e0;
}

.chat-info {
  flex: 1;
}

.chat-name {
  font-weight: 500;
  font-size: 16px;
}

.chat-status {
  font-size: 12px;
  color: #999;
}

.message-list {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background: #f9f9f9;
}

.message-item {
  display: flex;
  gap: 10px;
  margin-bottom: 15px;
}

.message-item.self {
  flex-direction: row-reverse;
}

.message-bubble {
  max-width: 60%;
  background: white;
  padding: 10px 15px;
  border-radius: 10px;
  box-shadow: 0 1px 2px rgba(0,0,0,0.1);
}

.message-item.self .message-bubble {
  background: #1890ff;
  color: white;
}

.message-content {
  word-break: break-word;
}

.message-time {
  font-size: 11px;
  color: #999;
  margin-top: 5px;
}

.message-item.self .message-time {
  color: rgba(255,255,255,0.8);
  text-align: right;
}

.input-area {
  display: flex;
  align-items: center;
  padding: 15px;
  gap: 10px;
  border-top: 1px solid #e0e0e0;
  background: white;
}

.message-input {
  flex: 1;
}

.no-chat {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
}

.login-card {
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
}

.login-button {
  width: 100%;
  margin-top: 10px;
}
</style>

