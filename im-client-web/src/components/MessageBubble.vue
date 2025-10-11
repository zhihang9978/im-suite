<template>
  <div class="message-bubble" :class="{ self: isSelf }">
    <!-- 头像 -->
    <el-avatar :size="35" :src="message.sender?.avatar">
      {{ message.sender?.nickname?.charAt(0) || 'U' }}
    </el-avatar>
    
    <!-- 消息内容 -->
    <div class="bubble-content">
      <!-- 文本消息 -->
      <div v-if="message.message_type === 'text'" class="text-message">
        {{ message.content }}
      </div>
      
      <!-- 图片消息 -->
      <div v-else-if="message.message_type === 'image'" class="image-message">
        <el-image
          :src="message.file_url"
          :preview-src-list="[message.file_url]"
          fit="cover"
          style="max-width: 300px; max-height: 300px;"
        />
      </div>
      
      <!-- 文件消息 -->
      <div v-else-if="message.message_type === 'file'" class="file-message">
        <el-icon><Document /></el-icon>
        <div class="file-info">
          <div class="file-name">{{ message.file_name || '文件' }}</div>
          <el-link :href="message.file_url" target="_blank" type="primary">
            下载
          </el-link>
        </div>
      </div>
      
      <!-- 语音消息 -->
      <div v-else-if="message.message_type === 'audio'" class="audio-message">
        <el-icon><Microphone /></el-icon>
        <audio controls :src="message.file_url"></audio>
      </div>
      
      <!-- 视频消息 -->
      <div v-else-if="message.message_type === 'video'" class="video-message">
        <video controls :src="message.file_url" style="max-width: 300px;"></video>
      </div>
      
      <!-- 时间戳 -->
      <div class="message-time">
        {{ formatTime(message.created_at) }}
        <el-icon v-if="message.is_read" class="read-icon"><Check /></el-icon>
      </div>
      
      <!-- 操作按钮 -->
      <div v-if="isSelf" class="message-actions">
        <el-dropdown @command="handleAction">
          <el-icon><MoreFilled /></el-icon>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="recall">撤回</el-dropdown-item>
              <el-dropdown-item command="delete">删除</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useUserStore } from '@/stores/user'
import { useMessageStore } from '@/stores/message'
import dayjs from 'dayjs'
import { ElMessage, ElMessageBox } from 'element-plus'

const props = defineProps({
  message: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['update'])

const userStore = useUserStore()
const messageStore = useMessageStore()

const isSelf = computed(() => props.message.sender_id === userStore.user?.id)

const formatTime = (time) => {
  return dayjs(time).format('HH:mm')
}

const handleAction = async (command) => {
  if (command === 'recall') {
    try {
      await ElMessageBox.confirm('确定要撤回这条消息吗？', '撤回消息', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
      
      const result = await messageStore.recallMessage(props.message.id)
      if (result.success) {
        ElMessage.success('消息已撤回')
        emit('update')
      } else {
        ElMessage.error('撤回失败')
      }
    } catch (error) {
      // 用户取消
    }
  } else if (command === 'delete') {
    try {
      await ElMessageBox.confirm('确定要删除这条消息吗？', '删除消息', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
      
      const result = await messageStore.deleteMessage(props.message.id)
      if (result.success) {
        ElMessage.success('消息已删除')
        emit('update')
      } else {
        ElMessage.error('删除失败')
      }
    } catch (error) {
      // 用户取消
    }
  }
}
</script>

<style scoped>
.message-bubble {
  display: flex;
  gap: 10px;
  margin-bottom: 15px;
  align-items: flex-start;
}

.message-bubble.self {
  flex-direction: row-reverse;
}

.bubble-content {
  max-width: 60%;
  background: white;
  padding: 10px 15px;
  border-radius: 10px;
  box-shadow: 0 1px 2px rgba(0,0,0,0.1);
  position: relative;
}

.message-bubble.self .bubble-content {
  background: #1890ff;
  color: white;
}

.text-message {
  word-break: break-word;
  line-height: 1.5;
}

.image-message,
.file-message,
.audio-message,
.video-message {
  margin-bottom: 5px;
}

.file-message {
  display: flex;
  align-items: center;
  gap: 10px;
}

.file-info {
  flex: 1;
}

.file-name {
  font-weight: 500;
  margin-bottom: 5px;
}

.message-time {
  font-size: 11px;
  color: #999;
  margin-top: 5px;
  display: flex;
  align-items: center;
  gap: 3px;
}

.message-bubble.self .message-time {
  color: rgba(255,255,255,0.8);
  justify-content: flex-end;
}

.read-icon {
  color: #52c41a;
}

.message-actions {
  position: absolute;
  top: 5px;
  right: 5px;
  opacity: 0;
  transition: opacity 0.2s;
}

.bubble-content:hover .message-actions {
  opacity: 1;
}
</style>

