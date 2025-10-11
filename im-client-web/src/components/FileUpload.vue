<template>
  <el-upload
    ref="uploadRef"
    action="#"
    :auto-upload="false"
    :on-change="handleFileChange"
    :show-file-list="false"
    accept="image/*,video/*,audio/*,.pdf,.doc,.docx,.zip,.txt"
  >
    <el-button :icon="Paperclip" circle />
  </el-upload>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Paperclip } from '@element-plus/icons-vue'
import api from '@/api/client'

const props = defineProps({
  receiverId: {
    type: Number,
    required: true
  }
})

const emit = defineEmits(['uploaded'])

const uploadRef = ref(null)

const handleFileChange = async (file) => {
  // 检查文件大小
  if (file.size > 50 * 1024 * 1024) {
    ElMessage.error('文件大小不能超过50MB')
    return
  }
  
  // 准备上传
  const formData = new FormData()
  formData.append('file', file.raw)
  
  try {
    ElMessage.loading('正在上传...')
    
    const response = await api.post('/api/files/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
    
    if (response.data.success) {
      // 上传成功，发送消息
      const messageType = getMessageType(file.raw.type)
      const messageData = {
        receiver_id: props.receiverId,
        content: `[${messageType}] ${file.name}`,
        message_type: messageType,
        file_url: response.data.data.url,
        file_name: file.name
      }
      
      const msgResponse = await api.post('/api/messages/send', messageData)
      
      if (msgResponse.data.success) {
        ElMessage.success('文件发送成功')
        emit('uploaded', msgResponse.data.data)
      }
    }
  } catch (error) {
    ElMessage.error('文件上传失败')
  }
}

const getMessageType = (mimeType) => {
  if (mimeType.startsWith('image/')) return 'image'
  if (mimeType.startsWith('video/')) return 'video'
  if (mimeType.startsWith('audio/')) return 'audio'
  return 'file'
}
</script>

