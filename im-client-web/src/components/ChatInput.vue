<template>
  <div class="chat-input">
    <div class="toolbar">
      <!-- 表情按钮 -->
      <el-popover placement="top-start" :width="300" trigger="click">
        <template #reference>
          <el-button :icon="Sunny" circle />
        </template>
        <div class="emoji-panel">
          <span
            v-for="emoji in commonEmojis"
            :key="emoji"
            class="emoji-item"
            @click="insertEmoji(emoji)"
          >
            {{ emoji }}
          </span>
        </div>
      </el-popover>
      
      <!-- 文件上传 -->
      <FileUpload
        v-if="receiverId"
        :receiver-id="receiverId"
        @uploaded="handleFileUploaded"
      />
      
      <!-- 图片上传 -->
      <el-button :icon="Picture" circle @click="handleImageUpload" />
      
      <!-- 语音通话 -->
      <el-button :icon="Phone" circle @click="handleVoiceCall" />
      
      <!-- 视频通话 -->
      <el-button :icon="VideoCamera" circle @click="handleVideoCall" />
    </div>
    
    <div class="input-wrapper">
      <el-input
        v-model="message"
        type="textarea"
        :rows="3"
        placeholder="输入消息..."
        @keyup.ctrl.enter="handleSend"
        @input="handleTyping"
      />
      
      <div class="send-button">
        <el-button type="primary" @click="handleSend" :loading="sending">
          发送 (Ctrl+Enter)
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Sunny, Picture, Phone, VideoCamera } from '@element-plus/icons-vue'
import FileUpload from './FileUpload.vue'

const props = defineProps({
  receiverId: Number
})

const emit = defineEmits(['send', 'uploaded', 'voiceCall', 'videoCall'])

const message = ref('')
const sending = ref(false)

const commonEmojis = [
  '😀', '😃', '😄', '😁', '😆', '😅', '🤣', '😂',
  '🙂', '🙃', '😉', '😊', '😇', '🥰', '😍', '🤩',
  '😘', '😗', '😚', '😙', '😋', '😛', '😜', '🤪',
  '😝', '🤑', '🤗', '🤭', '🤫', '🤔', '🤐', '🤨',
  '😐', '😑', '😶', '😏', '😒', '🙄', '😬', '🤥',
  '😌', '😔', '😪', '🤤', '😴', '😷', '🤒', '🤕',
  '🤢', '🤮', '🤧', '🥵', '🥶', '😶‍🌫️', '😵', '😵‍💫',
  '👍', '👎', '👏', '🙏', '💪', '🎉', '🎊', '❤️'
]

const insertEmoji = (emoji) => {
  message.value += emoji
}

const handleSend = () => {
  if (!message.value.trim()) {
    ElMessage.warning('请输入消息内容')
    return
  }
  
  emit('send', message.value)
  message.value = ''
}

const handleFileUploaded = (file) => {
  emit('uploaded', file)
}

const handleImageUpload = () => {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/*'
  input.click()
  // 实际上传逻辑由FileUpload组件处理
}

const handleVoiceCall = () => {
  emit('voiceCall')
}

const handleVideoCall = () => {
  emit('videoCall')
}

let typingTimer = null
const handleTyping = () => {
  if (typingTimer) clearTimeout(typingTimer)
  
  // 发送打字状态（节流）
  typingTimer = setTimeout(() => {
    // TODO: 通过WebSocket发送打字状态
  }, 500)
}
</script>

<style scoped>
.chat-input {
  border-top: 1px solid #e0e0e0;
  background: white;
}

.toolbar {
  display: flex;
  gap: 5px;
  padding: 10px 15px;
  border-bottom: 1px solid #f0f0f0;
}

.emoji-panel {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 5px;
}

.emoji-item {
  font-size: 24px;
  cursor: pointer;
  text-align: center;
  padding: 5px;
  border-radius: 5px;
  transition: background 0.2s;
}

.emoji-item:hover {
  background: #f0f0f0;
}

.input-wrapper {
  padding: 10px 15px;
}

.send-button {
  margin-top: 10px;
  text-align: right;
}
</style>

