<template>
  <div class="loading-skeleton">
    <el-skeleton :rows="rows" :animated="animated" :loading="loading">
      <template #template>
        <div v-for="i in rows" :key="i" class="skeleton-row">
          <el-skeleton-item variant="text" :style="{ width: getWidth(i) }" />
        </div>
      </template>
      <template #default>
        <slot></slot>
      </template>
    </el-skeleton>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  loading: {
    type: Boolean,
    default: true
  },
  rows: {
    type: Number,
    default: 5
  },
  animated: {
    type: Boolean,
    default: true
  }
})

// 每行宽度变化，更自然
const getWidth = (index) => {
  const widths = ['100%', '90%', '95%', '85%', '100%']
  return widths[index % widths.length]
}
</script>

<style lang="scss" scoped>
.loading-skeleton {
  padding: 20px;
  
  .skeleton-row {
    margin-bottom: 16px;
  }
}
</style>

