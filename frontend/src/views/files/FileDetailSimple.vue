<template>
  <div class="file-detail-simple" style="padding: 20px;">
    <h1>文件详情页 - 简化测试版</h1>
    
    <el-card style="margin: 20px 0;">
      <h2>页面状态</h2>
      <p>✓ 页面已加载</p>
      <p>文件ID: {{ $route.params.id }}</p>
      <p>Loading: {{ loading }}</p>
      <p>fileInfo存在: {{ !!fileInfo.id }}</p>
    </el-card>

    <el-card v-if="fileInfo.id" style="margin: 20px 0;">
      <h2>文件信息</h2>
      <p>ID: {{ fileInfo.id }}</p>
      <p>标题: {{ fileInfo.title }}</p>
      <p>类型: {{ fileInfo.type }}</p>
      <p>扩展名: {{ fileInfo.ext }}</p>
    </el-card>

    <el-card style="margin: 20px 0;">
      <h2>预览URL</h2>
      <p>计算结果: {{ previewUrl || '(空)' }}</p>
      <p>条件检查: type={{ fileInfo.type }}, isVideo={{ fileInfo.type === 1 }}</p>
    </el-card>

    <el-card v-if="showVideoPlayer" style="margin: 20px 0;">
      <h2>VideoPlayer组件测试</h2>
      <p style="background: #d4edda; padding: 10px;">✓ 条件满足，VideoPlayer应该在下方显示</p>
      <div style="background: #000; min-height: 400px; padding: 20px;">
        <VideoPlayer
          v-if="previewUrl"
          :src="previewUrl"
          type="video/x-flv"
        />
      </div>
    </el-card>

    <el-card v-else style="margin: 20px 0;">
      <h2>VideoPlayer未显示</h2>
      <p style="background: #f8d7da; padding: 10px;">
        条件不满足: 
        fileInfo.id={{ !!fileInfo.id }}, 
        type in [1,2]={{ [1,2].includes(fileInfo.type) }}, 
        previewUrl={{ !!previewUrl }}
      </p>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import VideoPlayer from '@/components/VideoPlayer.vue'

const route = useRoute()
const loading = ref(false)
const fileInfo = ref({})

const previewUrl = computed(() => {
  if (fileInfo.value.type === 1 || fileInfo.value.type === 2) {
    return `/api/v1/files/${route.params.id}/preview`
  }
  return null
})

const showVideoPlayer = computed(() => {
  return !!fileInfo.value.id && 
         (fileInfo.value.type === 1 || fileInfo.value.type === 2) && 
         !!previewUrl.value
})

const loadFileDetail = async () => {
  loading.value = true
  
  try {
    const response = await fetch(`/api/v1/files/${route.params.id}`, {
      credentials: 'include'
    })
    
    if (response.ok) {
      const data = await response.json()
      fileInfo.value = data.data || data
    } else {
      console.error('Failed to load file:', response.status)
    }
  } catch (error) {
    console.error('Error loading file:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadFileDetail()
})
</script>

<style scoped>
.file-detail-simple h2 {
  color: #333;
  border-bottom: 2px solid #4CAF50;
  padding-bottom: 10px;
  margin-top: 0;
}
</style>
