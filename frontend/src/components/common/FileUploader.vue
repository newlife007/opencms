<template>
  <div class="file-uploader">
    <div
      class="upload-dragger"
      :class="{ 'is-dragover': isDragover }"
      @drop.prevent="handleDrop"
      @dragover.prevent="handleDragover"
      @dragleave.prevent="handleDragleave"
      @click="triggerFileInput"
    >
      <input
        ref="fileInputRef"
        type="file"
        :accept="accept"
        :multiple="multiple"
        style="display: none"
        @change="handleFileSelect"
      />
      
      <div class="upload-dragger-icon">
        <el-icon :size="60"><UploadFilled /></el-icon>
      </div>
      
      <div class="upload-dragger-text">
        <div class="primary-text">
          {{ isDragover ? '释放以上传文件' : '拖拽文件到此处或点击上传' }}
        </div>
        <div class="secondary-text">
          <slot name="tip">
            <template v-if="accept">
              支持格式: {{ accept }}
            </template>
            <template v-if="maxSize">
              | 单个文件最大: {{ formatFileSize(maxSize) }}
            </template>
            <template v-if="maxFiles && multiple">
              | 最多上传: {{ maxFiles }} 个文件
            </template>
          </slot>
        </div>
      </div>
    </div>

    <!-- File List -->
    <div v-if="fileList.length > 0" class="file-list">
      <div
        v-for="item in fileList"
        :key="item.uid"
        class="file-item"
        :class="{ 'is-uploading': item.status === 'uploading' }"
      >
        <div class="file-item-info">
          <el-icon class="file-icon">
            <Document v-if="item.file.type.startsWith('application')" />
            <VideoPlay v-else-if="item.file.type.startsWith('video')" />
            <Microphone v-else-if="item.file.type.startsWith('audio')" />
            <Picture v-else-if="item.file.type.startsWith('image')" />
            <Document v-else />
          </el-icon>
          <div class="file-name">
            {{ item.file.name }}
            <span class="file-size">({{ formatFileSize(item.file.size) }})</span>
          </div>
        </div>

        <div class="file-item-actions">
          <template v-if="item.status === 'ready'">
            <el-button
              v-if="!autoUpload"
              size="small"
              type="primary"
              @click="$emit('upload', item)"
            >
              上传
            </el-button>
            <el-button
              size="small"
              type="danger"
              icon="Delete"
              @click="removeFile(item.uid)"
            >
              移除
            </el-button>
          </template>

          <template v-else-if="item.status === 'uploading'">
            <div class="upload-progress">
              <el-progress
                :percentage="item.progress"
                :status="item.progress === 100 ? 'success' : undefined"
              />
            </div>
            <el-button
              size="small"
              type="warning"
              icon="Close"
              @click="cancelUpload(item.uid)"
            >
              取消
            </el-button>
          </template>

          <template v-else-if="item.status === 'success'">
            <el-tag type="success">上传成功</el-tag>
            <el-button
              size="small"
              type="danger"
              icon="Delete"
              @click="removeFile(item.uid)"
            >
              移除
            </el-button>
          </template>

          <template v-else-if="item.status === 'error'">
            <el-tooltip :content="item.error" placement="top">
              <el-tag type="danger">上传失败</el-tag>
            </el-tooltip>
            <el-button
              size="small"
              type="primary"
              @click="retryUpload(item.uid)"
            >
              重试
            </el-button>
            <el-button
              size="small"
              type="danger"
              icon="Delete"
              @click="removeFile(item.uid)"
            >
              移除
            </el-button>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import {
  UploadFilled,
  Document,
  VideoPlay,
  Microphone,
  Picture,
} from '@element-plus/icons-vue'

const props = defineProps({
  accept: {
    type: String,
    default: '',
  },
  maxSize: {
    type: Number,
    default: 0, // bytes, 0 means no limit
  },
  multiple: {
    type: Boolean,
    default: false,
  },
  maxFiles: {
    type: Number,
    default: 0, // 0 means no limit
  },
  autoUpload: {
    type: Boolean,
    default: true,
  },
})

const emit = defineEmits(['file-select', 'file-success', 'file-error', 'upload'])

const fileInputRef = ref(null)
const isDragover = ref(false)
const fileList = ref([])
let uid = 0

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

const validateFile = (file) => {
  // Check file type
  if (props.accept) {
    const acceptTypes = props.accept.split(',').map(t => t.trim())
    const fileExt = '.' + file.name.split('.').pop().toLowerCase()
    const isValid = acceptTypes.some(type => {
      if (type.startsWith('.')) {
        return fileExt === type
      } else if (type.endsWith('/*')) {
        return file.type.startsWith(type.replace('/*', ''))
      }
      return file.type === type
    })
    
    if (!isValid) {
      ElMessage.error(`不支持的文件格式: ${file.name}`)
      return false
    }
  }

  // Check file size
  if (props.maxSize && file.size > props.maxSize) {
    ElMessage.error(`文件 ${file.name} 超过最大限制 ${formatFileSize(props.maxSize)}`)
    return false
  }

  // Check max files
  if (props.maxFiles && fileList.value.length >= props.maxFiles) {
    ElMessage.error(`最多只能上传 ${props.maxFiles} 个文件`)
    return false
  }

  return true
}

const addFiles = (files) => {
  const validFiles = Array.from(files).filter(validateFile)
  
  validFiles.forEach(file => {
    const fileItem = {
      uid: ++uid,
      file,
      status: 'ready',
      progress: 0,
      error: null,
    }
    fileList.value.push(fileItem)
    emit('file-select', fileItem)
    
    // Auto upload if enabled
    if (props.autoUpload) {
      emit('upload', fileItem)
    }
  })
}

const handleDrop = (e) => {
  isDragover.value = false
  const files = e.dataTransfer.files
  if (files.length > 0) {
    addFiles(files)
  }
}

const handleDragover = () => {
  isDragover.value = true
}

const handleDragleave = () => {
  isDragover.value = false
}

const triggerFileInput = () => {
  fileInputRef.value?.click()
}

const handleFileSelect = (e) => {
  const files = e.target.files
  if (files.length > 0) {
    addFiles(files)
  }
  // Reset input
  e.target.value = ''
}

const removeFile = (uid) => {
  const index = fileList.value.findIndex(item => item.uid === uid)
  if (index > -1) {
    fileList.value.splice(index, 1)
  }
}

const cancelUpload = (uid) => {
  const item = fileList.value.find(f => f.uid === uid)
  if (item) {
    item.status = 'ready'
    item.progress = 0
  }
}

const retryUpload = (uid) => {
  const item = fileList.value.find(f => f.uid === uid)
  if (item) {
    item.status = 'ready'
    item.progress = 0
    item.error = null
    emit('upload', item)
  }
}

const updateFileStatus = (uid, status, progress = 0, error = null) => {
  const item = fileList.value.find(f => f.uid === uid)
  if (item) {
    item.status = status
    item.progress = progress
    item.error = error
  }
}

const clearFiles = () => {
  fileList.value = []
}

// Expose methods
defineExpose({
  updateFileStatus,
  clearFiles,
})
</script>

<style scoped>
.file-uploader {
  width: 100%;
}

.upload-dragger {
  border: 2px dashed #d9d9d9;
  border-radius: 6px;
  padding: 40px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s;
  background-color: #fafafa;
}

.upload-dragger:hover,
.upload-dragger.is-dragover {
  border-color: #409eff;
  background-color: #f0f9ff;
}

.upload-dragger-icon {
  color: #c0c4cc;
  margin-bottom: 16px;
}

.upload-dragger.is-dragover .upload-dragger-icon {
  color: #409eff;
}

.primary-text {
  font-size: 16px;
  color: #606266;
  margin-bottom: 8px;
}

.secondary-text {
  font-size: 14px;
  color: #909399;
}

.file-list {
  margin-top: 16px;
}

.file-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  margin-bottom: 8px;
  transition: all 0.3s;
}

.file-item:hover {
  background-color: #f5f7fa;
}

.file-item.is-uploading {
  background-color: #f0f9ff;
}

.file-item-info {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
}

.file-icon {
  font-size: 24px;
  margin-right: 12px;
  color: #606266;
}

.file-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 14px;
  color: #303133;
}

.file-size {
  color: #909399;
  margin-left: 8px;
}

.file-item-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.upload-progress {
  width: 200px;
}
</style>
