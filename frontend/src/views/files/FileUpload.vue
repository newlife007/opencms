<template>
  <div class="file-upload">
    <el-card>
      <template #header>
        <span>{{ t("fileUpload.title") }}</span>
      </template>

      <el-form :model="uploadForm" :rules="uploadRules" ref="uploadFormRef" label-width="120px">
        <el-form-item label="选择文件" prop="file" required>
          <el-upload
            ref="uploadRef"
            class="upload-demo"
            drag
            :auto-upload="false"
            :on-change="handleFileChange"
            :on-remove="handleFileRemove"
            :limit="1"
            :file-list="fileList"
            accept="*"
          >
            <el-icon class="el-icon--upload"><upload-filled /></el-icon>
            <div class="el-upload__text">
              拖拽文件到此处或 <em>点击上传</em>
            </div>
            <template #tip>
              <div class="el-upload__tip">
                支持视频、音频、图片、富媒体文件，最大500MB
              </div>
            </template>
          </el-upload>
        </el-form-item>

        <el-form-item :label="t('files.fileType')" prop="type">
          <el-select v-model="uploadForm.type" placeholder="请选择文件类型">
            <el-option label="视频" :value="1" />
            <el-option label="音频" :value="2" />
            <el-option label="图片" :value="3" />
            <el-option label="富媒体" :value="4" />
          </el-select>
        </el-form-item>

        <el-form-item label="文件标题" prop="title">
          <el-input v-model="uploadForm.title" placeholder="请输入文件标题" />
        </el-form-item>

        <el-form-item label="所属分类" prop="category_id">
          <el-tree-select
            v-model="uploadForm.category_id"
            :data="categoryTree"
            :props="{ label: 'name', value: 'id' }"
            placeholder="请选择分类"
            check-strictly
          />
        </el-form-item>

        <el-form-item label="文件描述">
          <el-input
            v-model="uploadForm.description"
            type="textarea"
            :rows="4"
            placeholder="请输入文件描述（可选）"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleUpload" :loading="uploading" :disabled="!uploadForm.file">
            <el-icon><Upload /></el-icon>
            {{ uploading ? '上传中...' : '开始上传' }}
          </el-button>
          <el-button @click="handleReset">{{ t("common.reset") }}</el-button>
          <el-button @click="$router.back()">返回</el-button>
        </el-form-item>

        <!-- Upload Progress -->
        <el-form-item v-if="uploading">
          <el-progress :percentage="uploadProgress" :status="uploadStatus" />
          <div class="upload-info">
            <span>已上传: {{ formatFileSize(uploadedSize) }}</span>
            <span>总大小: {{ formatFileSize(totalSize) }}</span>
            <span>速度: {{ uploadSpeed }}</span>
          </div>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 本次上传成功的文件列表 -->
    <el-card v-if="uploadedFiles.length > 0" style="margin-top: 20px;">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <span>本次上传成功的文件 ({{ uploadedFiles.length }})</span>
          <el-button size="small" @click="clearUploadedFiles">清空列表</el-button>
        </div>
      </template>

      <el-table :data="uploadedFiles" style="width: 100%" stripe>
        <el-table-column type="index" label="#" width="60" />
        
        <el-table-column prop="title" label="文件标题" min-width="200">
          <template #default="{ row }">
            <el-link :href="`/files/${row.id}`" type="primary">
              {{ row.title }}
            </el-link>
          </template>
        </el-table-column>

        <el-table-column prop="name" :label="t('files.fileName')" min-width="150">
          <template #default="{ row }">
            <el-tooltip :content="row.name + row.ext" placement="top">
              <span>{{ row.name.substring(0, 12) }}...{{ row.ext }}</span>
            </el-tooltip>
          </template>
        </el-table-column>

        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.type === 1" type="danger">视频</el-tag>
            <el-tag v-else-if="row.type === 2" type="warning">音频</el-tag>
            <el-tag v-else-if="row.type === 3" type="success">图片</el-tag>
            <el-tag v-else-if="row.type === 4" type="info">富媒体</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="size" :label="t('files.fileSize')" width="120">
          <template #default="{ row }">
            {{ formatFileSize(row.size) }}
          </template>
        </el-table-column>

        <el-table-column prop="uploaded" :label="t('files.uploadTime')" width="180">
          <template #default="{ row }">
            {{ formatUploadTime(row.uploaded) }}
          </template>
        </el-table-column>

        <el-table-column prop="status" :label="t('files.status')" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.status === 0" type="info">新建</el-tag>
            <el-tag v-else-if="row.status === 1" type="warning">{{ t("fileList.pendingFiles") }}</el-tag>
            <el-tag v-else-if="row.status === 2" type="success">{{ t("fileList.publishedFiles") }}</el-tag>
            <el-tag v-else-if="row.status === 3" type="danger">已拒绝</el-tag>
            <el-tag v-else-if="row.status === 4" type="info">已删除</el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('common.actions')" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="viewFile(row.id)">{{ t("fileList.viewDetail") }}</el-button>
            <el-button size="small" type="success" @click="catalogFile(row.id)">
              编目
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import filesApi from '@/api/files'
import categoryApi from '@/api/category'


const { t } = useI18n()

const router = useRouter()
const uploadFormRef = ref(null)
const uploadRef = ref(null)
const uploading = ref(false)
const uploadProgress = ref(0)
const uploadStatus = ref('')
const uploadedSize = ref(0)
const totalSize = ref(0)
const uploadSpeed = ref('0 KB/s')
const fileList = ref([])
const categoryTree = ref([])
const uploadedFiles = ref([]) // 存储本次登录期间上传成功的文件

const uploadForm = reactive({
  file: null,
  type: null,
  title: '',
  category_id: null,
  description: '',
})

const uploadRules = {
  type: [{ required: true, message: '请选择文件类型', trigger: 'change' }],
  title: [
    { required: true, message: '请输入文件标题', trigger: 'blur' },
    { min: 2, max: 200, message: '标题长度在 2 到 200 个字符', trigger: 'blur' },
  ],
  category_id: [{ required: true, message: '请选择分类', trigger: 'change' }],
}

const handleFileChange = (file) => {
  uploadForm.file = file.raw
  if (!uploadForm.title) {
    // Auto-fill title from filename (without extension)
    uploadForm.title = file.name.replace(/\.[^/.]+$/, '')
  }
  
  // Auto-detect file type
  const ext = file.name.split('.').pop().toLowerCase()
  if (['mp4', 'avi', 'mov', 'wmv', 'flv', 'mkv', 'mpg', 'mpeg'].includes(ext)) {
    uploadForm.type = 1 // Video
  } else if (['mp3', 'wav', 'wma', 'aac', 'flac', 'ogg', 'm4a'].includes(ext)) {
    uploadForm.type = 2 // Audio
  } else if (['jpg', 'jpeg', 'png', 'gif', 'bmp', 'tiff', 'webp'].includes(ext)) {
    uploadForm.type = 3 // Image
  } else if (['swf', 'pdf', 'doc', 'docx', 'ppt', 'pptx'].includes(ext)) {
    uploadForm.type = 4 // Rich media
  }
}

const handleFileRemove = () => {
  uploadForm.file = null
}

const formatFileSize = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

const handleUpload = async () => {
  if (!uploadFormRef.value) return

  await uploadFormRef.value.validate(async (valid) => {
    if (!valid) return

    uploading.value = true
    uploadProgress.value = 0
    uploadStatus.value = ''
    totalSize.value = uploadForm.file.size
    
    const startTime = Date.now()
    let lastLoaded = 0
    let lastTime = startTime

    try {
      const formData = new FormData()
      formData.append('file', uploadForm.file)
      formData.append('title', uploadForm.title)
      formData.append('type', uploadForm.type)
      formData.append('category_id', uploadForm.category_id)
      if (uploadForm.description) {
        formData.append('description', uploadForm.description)
      }

      const res = await filesApi.upload(formData, (progressEvent) => {
        uploadedSize.value = progressEvent.loaded
        uploadProgress.value = Math.round((progressEvent.loaded * 100) / progressEvent.total)
        
        // Calculate upload speed
        const now = Date.now()
        const timeDiff = (now - lastTime) / 1000 // seconds
        const bytesDiff = progressEvent.loaded - lastLoaded
        
        if (timeDiff > 0) {
          const speed = bytesDiff / timeDiff
          uploadSpeed.value = formatFileSize(speed) + '/s'
          lastTime = now
          lastLoaded = progressEvent.loaded
        }
      })

      if (res.success) {
        uploadStatus.value = 'success'
        ElMessage.success('上传成功！')
        
        // 将上传成功的文件添加到列表
        uploadedFiles.value.unshift({
          id: res.file.id,
          name: res.file.name,
          ext: res.file.ext,
          title: res.file.title,
          type: res.file.type,
          size: res.file.size,
          status: res.file.status,
          path: res.file.path,
          uploaded: res.file.uploaded || Date.now() / 1000,
        })
        
        // 保存到 localStorage，以便刷新页面后仍然显示
        saveUploadedFilesToStorage()
        
        // 重置表单
        setTimeout(() => {
          handleReset()
        }, 1000)
      }
    } catch (error) {
      uploadStatus.value = 'exception'
      ElMessage.error('上传失败: ' + error.message)
    } finally {
      uploading.value = false
    }
  })
}

const handleReset = () => {
  uploadFormRef.value?.resetFields()
  uploadRef.value?.clearFiles()
  fileList.value = []
  uploadForm.file = null
  uploadProgress.value = 0
  uploadedSize.value = 0
  totalSize.value = 0
  uploadSpeed.value = '0 KB/s'
}

const loadCategories = async () => {
  try {
    console.log('Loading categories tree...')
    const res = await categoryApi.getTree()
    console.log('Categories response:', res)
    
    if (res.success) {
      categoryTree.value = res.data || []
      if (categoryTree.value.length === 0) {
        ElMessage.warning('暂无分类数据，请先在分类管理中添加分类')
      }
    } else {
      ElMessage.error(res.message || '加载分类失败')
      categoryTree.value = []
    }
  } catch (error) {
    console.error('Load categories error:', error)
    ElMessage.error('加载分类失败: ' + (error.message || '网络错误'))
    categoryTree.value = []
  }
}

// 保存上传文件列表到 localStorage
const saveUploadedFilesToStorage = () => {
  try {
    const sessionKey = 'uploaded_files_session_' + new Date().toDateString()
    localStorage.setItem(sessionKey, JSON.stringify(uploadedFiles.value))
  } catch (error) {
    console.error('Failed to save uploaded files to localStorage:', error)
  }
}

// 从 localStorage 加载上传文件列表
const loadUploadedFilesFromStorage = () => {
  try {
    const sessionKey = 'uploaded_files_session_' + new Date().toDateString()
    const stored = localStorage.getItem(sessionKey)
    if (stored) {
      uploadedFiles.value = JSON.parse(stored)
    }
  } catch (error) {
    console.error('Failed to load uploaded files from localStorage:', error)
  }
}

// 清空上传文件列表
const clearUploadedFiles = () => {
  uploadedFiles.value = []
  const sessionKey = 'uploaded_files_session_' + new Date().toDateString()
  localStorage.removeItem(sessionKey)
  ElMessage.success('已清空上传列表')
}

// 格式化上传时间
const formatUploadTime = (timestamp) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

// 查看文件详情
const viewFile = (fileId) => {
  router.push(`/files/${fileId}`)
}

// 编辑文件编目
const catalogFile = (fileId) => {
  router.push(`/files/${fileId}/catalog`)
}

onMounted(() => {
  loadCategories()
  loadUploadedFilesFromStorage() // 加载本次会话的上传记录
})
</script>

<style scoped>
.file-upload {
  padding: 20px;
}

.upload-demo {
  width: 100%;
}

.upload-demo :deep(.el-upload) {
  width: 100%;
}

.upload-demo :deep(.el-upload-dragger) {
  width: 100%;
}

.upload-info {
  display: flex;
  justify-content: space-between;
  margin-top: 10px;
  font-size: 14px;
  color: #666;
}
</style>
