<template>
  <div class="file-detail">
    <el-page-header @back="$router.back()" title="返回">
      <template #content>
        <span>{{ t("fileDetail.title") }}</span>
      </template>
    </el-page-header>

    <el-row :gutter="20" style="margin-top: 20px">
      <!-- Left Column - Preview and Info -->
      <el-col :span="16">
        <!-- Preview -->
        <el-card v-loading="loading">
          <template #header>
            <span>文件预览</span>
          </template>
          
          <div class="preview-container">
            <!-- Video/Audio Preview -->
            <div v-if="fileInfo.id && (fileInfo.type === 1 || fileInfo.type === 2) && previewUrl" class="video-wrapper">
              <VideoPlayer
                :src="previewUrl"
                :type="videoType"
              />
            </div>
            
            <!-- Image Preview -->
            <el-image
              v-else-if="fileInfo.type === 3"
              :src="`/api/v1/files/${fileId}/preview`"
              fit="contain"
              style="max-width: 100%; max-height: 600px"
            />
            
            <!-- Other types -->
            <div v-else class="no-preview">
              <el-icon :size="100" color="#ccc"><Document /></el-icon>
              <p>该文件类型不支持在线预览</p>
            </div>
          </div>
        </el-card>

        <!-- File Info -->
        <el-card style="margin-top: 20px">
          <template #header>
            <span>文件信息</span>
          </template>
          
          <el-descriptions :column="2" border>
            <el-descriptions-item label="文件ID">{{ fileInfo.id }}</el-descriptions-item>
            <el-descriptions-item label="文件标题">{{ fileInfo.title }}</el-descriptions-item>
            <el-descriptions-item :label="t('files.fileType')">
              <el-tag>{{ getFileTypeName(fileInfo.type) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="文件状态">
              <el-tag :type="getStatusType(fileInfo.status)">
                {{ getStatusName(fileInfo.status) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="t('files.fileSize')">
              {{ formatFileSize(fileInfo.size) }}
            </el-descriptions-item>
            <el-descriptions-item label="文件格式">{{ fileInfo.ext }}</el-descriptions-item>
            <el-descriptions-item label="所属分类">{{ fileInfo.category_name }}</el-descriptions-item>
            <el-descriptions-item :label="t('files.uploader')">{{ fileInfo.upload_username }}</el-descriptions-item>
            <el-descriptions-item :label="t('files.uploadTime')">
              {{ formatDate(fileInfo.upload_at) }}
            </el-descriptions-item>
            <el-descriptions-item label="编目时间">
              {{ formatDate(fileInfo.catalog_at) }}
            </el-descriptions-item>
            <el-descriptions-item label="编目人">{{ fileInfo.catalog_username }}</el-descriptions-item>
            <el-descriptions-item label="发布时间">
              {{ formatDate(fileInfo.putout_at) }}
            </el-descriptions-item>
          </el-descriptions>
        </el-card>

        <!-- Catalog Info -->
        <el-card style="margin-top: 20px" v-if="fileInfo.catalog_info">
          <template #header>
            <span>{{ t("fileDetail.catalogInfo") }}</span>
          </template>
          
          <el-descriptions :column="1" border>
            <el-descriptions-item
              v-for="(value, key) in catalogInfo"
              :key="key"
              :label="key"
            >
              {{ value }}
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>

      <!-- Right Column - Actions -->
      <el-col :span="8">
        <el-card>
          <template #header>
            <span>操作</span>
          </template>
          
          <div class="action-buttons">
            <el-button
              type="primary"
              icon="Edit"
              @click="handleCatalog"
              :disabled="!canEdit"
            >
              编目
            </el-button>
            
            <el-button
              type="success"
              icon="Check"
              @click="handleSubmit"
              v-if="fileInfo.status === 0"
            >{{ t("fileCatalog.submitForReview") }}</el-button>
            
            <el-button
              type="success"
              icon="CircleCheck"
              @click="handlePublish"
              v-if="fileInfo.status === 1 && canPublish"
            >
              发布
            </el-button>
            
            <el-button
              type="warning"
              icon="CircleClose"
              @click="handleReject"
              v-if="fileInfo.status === 1 && canPublish"
            >{{ t("fileApproval.reject") }}</el-button>
            
            <el-button
              type="info"
              icon="Download"
              @click="handleDownload"
              :disabled="!fileInfo.is_download"
            >{{ t("fileDetail.downloadFile") }}</el-button>
            
            <el-button
              type="danger"
              icon="Delete"
              @click="handleDelete"
            >{{ t("fileList.deleteFile") }}</el-button>
          </div>
        </el-card>

        <!-- History -->
        <el-card style="margin-top: 20px">
          <template #header>
            <span>操作历史</span>
          </template>
          
          <el-timeline>
            <el-timeline-item
              v-if="fileInfo.upload_at"
              :timestamp="formatDate(fileInfo.upload_at)"
              placement="top"
            >
              <p>{{ fileInfo.upload_username }} 上传了文件</p>
            </el-timeline-item>
            <el-timeline-item
              v-if="fileInfo.catalog_at"
              :timestamp="formatDate(fileInfo.catalog_at)"
              placement="top"
            >
              <p>{{ fileInfo.catalog_username }} 完成编目</p>
            </el-timeline-item>
            <el-timeline-item
              v-if="fileInfo.putout_at"
              :timestamp="formatDate(fileInfo.putout_at)"
              placement="top"
            >
              <p>{{ fileInfo.putout_username }} 发布了文件</p>
            </el-timeline-item>
          </el-timeline>
        </el-card>
      </el-col>
    </el-row>

    <!-- Catalog Dialog -->
    <el-dialog v-model="catalogDialogVisible" title="文件编目" width="900px" :close-on-click-modal="false">
      <el-form :model="catalogForm" :rules="catalogRules" ref="catalogFormRef" label-width="120px">
        <!-- 基本信息 -->
        <el-divider content-position="left">{{ t("fileDetail.basicInfo") }}</el-divider>
        
        <el-form-item label="文件标题" prop="title">
          <el-input v-model="catalogForm.title" placeholder="请输入文件标题" clearable />
        </el-form-item>
        
        <el-form-item label="所属分类" prop="category_id">
          <el-tree-select
            v-model="catalogForm.category_id"
            :data="categoryTree"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            placeholder="请选择分类"
            check-strictly
            clearable
            :render-after-expand="false"
          />
        </el-form-item>
        
        <el-form-item label="描述信息">
          <el-input
            v-model="catalogForm.description"
            type="textarea"
            :rows="4"
            placeholder="请输入文件描述"
          />
        </el-form-item>
        
        <!-- 权限控制 -->
        <el-divider content-position="left">权限控制</el-divider>
        
        <el-form-item label="浏览等级" prop="level">
          <el-select v-model="catalogForm.level" placeholder="请选择浏览等级" clearable>
            <el-option
              v-for="level in levels"
              :key="level.id"
              :label="level.name"
              :value="level.id"
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="可访问组" prop="groups">
          <el-select
            v-model="catalogForm.groups"
            multiple
            placeholder="请选择可访问的组（不选则所有组可访问）"
            clearable
          >
            <el-option
              v-for="group in groups"
              :key="group.id"
              :label="group.name"
              :value="group.id"
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="允许下载">
          <el-switch v-model="catalogForm.is_download" />
        </el-form-item>
        
        <!-- 扩展属性 -->
        <el-divider content-position="left">扩展属性</el-divider>
        
        <template v-if="catalogFields.length > 0">
          <el-form-item
            v-for="field in catalogFields"
            :key="field.id"
            :label="field.label"
            :prop="'catalog_info.' + field.name"
          >
            <!-- 文本输入 -->
            <el-input
              v-if="field.type === 'text'"
              v-model="catalogForm.catalog_info[field.name]"
              :placeholder="'请输入' + field.label"
            />
            
            <!-- 数字输入 -->
            <el-input-number
              v-else-if="field.type === 'number'"
              v-model="catalogForm.catalog_info[field.name]"
              :placeholder="'请输入' + field.label"
            />
            
            <!-- 日期选择 -->
            <el-date-picker
              v-else-if="field.type === 'date'"
              v-model="catalogForm.catalog_info[field.name]"
              type="date"
              :placeholder="'请选择' + field.label"
            />
            
            <!-- 下拉选择 -->
            <el-select
              v-else-if="field.type === 'select'"
              v-model="catalogForm.catalog_info[field.name]"
              :placeholder="'请选择' + field.label"
            >
              <el-option
                v-for="option in field.options"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              />
            </el-select>
            
            <!-- 多行文本 -->
            <el-input
              v-else-if="field.type === 'textarea'"
              v-model="catalogForm.catalog_info[field.name]"
              type="textarea"
              :rows="3"
              :placeholder="'请输入' + field.label"
            />
          </el-form-item>
        </template>
        <el-empty v-else description="该文件类型暂无扩展属性配置" :image-size="80" />
      </el-form>
      
      <template #footer>
        <el-button @click="catalogDialogVisible = false">{{ t("common.cancel") }}</el-button>
        <el-button @click="saveDraft" :loading="catalogLoading">{{ t("fileCatalog.saveDraft") }}</el-button>
        <el-button type="primary" @click="submitCatalog" :loading="catalogLoading">
          保存并提交审核
        </el-button>
      </template>
    </el-dialog>

    <!-- Reject Dialog -->
    <el-dialog v-model="rejectDialogVisible" title="拒绝原因" width="500px">
      <el-input
        v-model="rejectReason"
        type="textarea"
        :rows="4"
        placeholder="请输入拒绝原因"
      />
      <template #footer>
        <el-button @click="rejectDialogVisible = false">{{ t("common.cancel") }}</el-button>
        <el-button type="primary" @click="confirmReject">{{ t("common.confirm") }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/user'
import filesApi from '@/api/files'
import categoryApi from '@/api/category'
import catalogApi from '@/api/catalog'
import request from '@/utils/request'
import VideoPlayer from '@/components/VideoPlayer.vue'


const { t } = useI18n()

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const loading = ref(false)
const fileInfo = ref({})
const rejectDialogVisible = ref(false)
const rejectReason = ref('')

// Catalog dialog states
const catalogDialogVisible = ref(false)
const catalogLoading = ref(false)
const catalogFormRef = ref(null)
const categoryTree = ref([])
const levels = ref([])
const groups = ref([])
const catalogFields = ref([])

// Catalog form
const catalogForm = reactive({
  title: '',
  category_id: null,
  description: '',
  level: null,
  groups: [],
  is_download: true,
  catalog_info: {}
})

// Catalog form validation rules
const catalogRules = {
  title: [
    { required: true, message: '请输入文件标题', trigger: 'blur' },
    { min: 2, max: 200, message: '标题长度在 2 到 200 个字符', trigger: 'blur' }
  ],
  category_id: [
    { required: true, message: '请选择所属分类', trigger: 'change' }
  ]
}

const fileId = computed(() => route.params.id)

// 视频预览URL和类型
const videoUrl = ref('')
const videoType = ref('video/x-flv') // Preview files are transcoded to FLV format

const previewUrl = computed(() => {
  return videoUrl.value
})

// 根据文件扩展名确定MIME类型
const getMimeType = (ext) => {
  const mimeTypes = {
    '.mp4': 'video/mp4',
    '.flv': 'video/x-flv',
    '.avi': 'video/x-msvideo',
    '.mov': 'video/quicktime',
    '.wmv': 'video/x-ms-wmv',
    '.mp3': 'audio/mpeg',
    '.wav': 'audio/wav',
    '.ogg': 'audio/ogg',
  }
  return mimeTypes[ext.toLowerCase()] || 'video/mp4'
}

// 检测并设置视频URL
const setupVideoUrl = async () => {
  if (fileInfo.value.type !== 1 && fileInfo.value.type !== 2) {
    return
  }

  // 先尝试预览文件（转码后的FLV）
  const previewFileUrl = filesApi.getPreviewUrl(fileId.value)
  
  try {
    const response = await fetch(previewFileUrl, { method: 'HEAD' })
    if (response.ok) {
      // 预览文件存在，使用FLV格式（FLV.js已集成）
      videoUrl.value = previewFileUrl
      videoType.value = 'video/x-flv'
      console.log('Using preview file (FLV):', previewFileUrl)
      return
    }
  } catch (e) {
    console.warn('Preview file not available, will use original file:', e)
  }

  // 使用原始文件
  videoUrl.value = filesApi.getDownloadUrl(fileId.value)
  videoType.value = getMimeType(fileInfo.value.ext || '.mp4')
  console.log('Using original file:', videoUrl.value, 'type:', videoType.value)
}

const catalogInfo = computed(() => {
  try {
    return JSON.parse(fileInfo.value.catalog_info || '{}')
  } catch (error) {
    return {}
  }
})

const canEdit = computed(() => {
  return fileInfo.value.status !== 2 // Can't edit published files
})

const canPublish = computed(() => {
  return userStore.isAdmin() || userStore.hasPermission('admin.fileputout.publish')
})

const getFileTypeName = (type) => {
  const types = { 1: '视频', 2: '音频', 3: '图片', 4: '富媒体' }
  return types[type] || '未知'
}

const getStatusName = (status) => {
  const statuses = {
    0: '新上传',
    1: '待审核',
    2: '已发布',
    3: '已拒绝',
    4: '已删除',
  }
  return statuses[status] || '未知'
}

const getStatusType = (status) => {
  const types = {
    0: 'info',
    1: 'warning',
    2: 'success',
    3: 'danger',
    4: 'info',
  }
  return types[status] || 'info'
}

const formatFileSize = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

const formatDate = (timestamp) => {
  if (!timestamp) return '-'
  return new Date(timestamp * 1000).toLocaleString('zh-CN')
}

const loadFileDetail = async () => {
  loading.value = true
  try {
    const res = await filesApi.getDetail(fileId.value)
    if (res.success) {
      fileInfo.value = res.data
      // 设置视频URL
      await setupVideoUrl()
    }
  } catch (error) {
    ElMessage.error('加载文件详情失败')
  } finally {
    loading.value = false
  }
}

// Load category tree
const loadCategoryTree = async () => {
  try {
    const res = await categoryApi.getTree()
    if (res.success) {
      categoryTree.value = res.data || []
    }
  } catch (error) {
    console.error('加载分类树失败:', error)
  }
}

// Load levels
const loadLevels = async () => {
  try {
    const res = await request.get('/admin/levels')
    levels.value = res.data || []
  } catch (error) {
    console.error('加载等级失败:', error)
  }
}

// Load groups
const loadGroups = async () => {
  try {
    const res = await request.get('/admin/groups')
    groups.value = res.data || []
  } catch (error) {
    console.error('加载组失败:', error)
  }
}

// Load catalog fields based on file type
const loadCatalogFields = async (fileType) => {
  if (!fileType) return
  
  try {
    console.log('Loading catalog fields for type:', fileType)
    const res = await request.get(`/catalog`, { params: { type: fileType } })
    console.log('Catalog API response:', res)
    
    // Backend returns { success: true, type: X, catalog: [...] }
    if (res.success && res.catalog) {
      console.log('Raw catalog tree:', res.catalog)
      const flattenedFields = flattenCatalogTree(res.catalog)
      console.log('Flattened catalog fields:', flattenedFields)
      catalogFields.value = flattenedFields
    } else {
      console.warn('No catalog data in response')
      catalogFields.value = []
    }
  } catch (error) {
    console.error('加载编目字段失败:', error)
    console.error('Error details:', error.response)
    catalogFields.value = []
  }
}

// Flatten catalog tree to simple field list
const flattenCatalogTree = (tree) => {
  const fields = []
  
  const traverse = (nodes) => {
    if (!nodes || !Array.isArray(nodes)) return
    
    for (const node of nodes) {
      // Skip group nodes, only process actual fields
      if (node.field_type === 'group') {
        // Just traverse children of group nodes
        if (node.children && node.children.length > 0) {
          traverse(node.children)
        }
        continue
      }
      
      // Add current node as field (non-group nodes)
      if (node.name && node.label) {
        fields.push({
          id: node.id,
          name: node.name,
          label: node.label,
          type: node.field_type || 'text', // Use field_type from backend
          required: node.required || false,
          options: node.options ? JSON.parse(node.options) : []
        })
      }
      
      // Traverse children
      if (node.children && node.children.length > 0) {
        traverse(node.children)
      }
    }
  }
  
  traverse(tree)
  return fields
}

const handleCatalog = async () => {
  // Load necessary data if not loaded
  if (categoryTree.value.length === 0) {
    await loadCategoryTree()
  }
  if (levels.value.length === 0) {
    await loadLevels()
  }
  if (groups.value.length === 0) {
    await loadGroups()
  }
  
  // Load catalog fields for this file type
  await loadCatalogFields(fileInfo.value.type)
  
  // Initialize catalog form with current file info
  catalogForm.title = fileInfo.value.title || ''
  catalogForm.category_id = fileInfo.value.category_id || null
  catalogForm.description = fileInfo.value.description || ''
  catalogForm.level = fileInfo.value.level || null
  catalogForm.groups = fileInfo.value.groups ? fileInfo.value.groups.split(',').map(Number).filter(Boolean) : []
  catalogForm.is_download = fileInfo.value.is_download !== 0
  
  // Parse catalog_info
  try {
    const catalogInfo = typeof fileInfo.value.catalog_info === 'string' 
      ? JSON.parse(fileInfo.value.catalog_info || '{}')
      : (fileInfo.value.catalog_info || {})
    catalogForm.catalog_info = catalogInfo
  } catch (e) {
    catalogForm.catalog_info = {}
  }
  
  catalogDialogVisible.value = true
}

// Save draft
const saveDraft = async () => {
  if (!catalogFormRef.value) return
  
  try {
    await catalogFormRef.value.validate()
  } catch (error) {
    ElMessage.warning('请检查表单内容')
    return
  }
  
  catalogLoading.value = true
  try {
    const updateData = {
      title: catalogForm.title,
      category_id: catalogForm.category_id,
      description: catalogForm.description,
      level: catalogForm.level,
      groups: catalogForm.groups.join(','),
      is_download: catalogForm.is_download ? 1 : 0,
      catalog_info: JSON.stringify(catalogForm.catalog_info),
      status: 0 // 草稿状态
    }
    
    const res = await filesApi.update(fileId.value, updateData)
    if (res.success) {
      ElMessage.success('草稿保存成功')
      await loadFileDetail()
    } else {
      ElMessage.error(res.message || '保存失败')
    }
  } catch (error) {
    console.error('Save draft error:', error)
    ElMessage.error(t('message.saveFailed'))
  } finally {
    catalogLoading.value = false
  }
}

// Submit catalog
const submitCatalog = async () => {
  if (!catalogFormRef.value) return
  
  try {
    await catalogFormRef.value.validate()
  } catch (error) {
    ElMessage.warning('请检查表单内容')
    return
  }
  
  catalogLoading.value = true
  try {
    const updateData = {
      title: catalogForm.title,
      category_id: catalogForm.category_id,
      description: catalogForm.description,
      level: catalogForm.level,
      groups: catalogForm.groups.join(','),
      is_download: catalogForm.is_download ? 1 : 0,
      catalog_info: JSON.stringify(catalogForm.catalog_info)
    }
    
    // 步骤1：保存编目信息
    const res = await filesApi.update(fileId.value, updateData)
    if (!res.success) {
      ElMessage.error(res.message || '保存失败')
      return
    }
    console.log('✓ 编目信息已保存')
    
    // 步骤2：提交审核
    await filesApi.submit(fileId.value)
    console.log('✓ 文件已提交审核')
    
    ElMessage.success('编目信息已保存并提交审核')
    catalogDialogVisible.value = false
    await loadFileDetail()
    
  } catch (error) {
    console.error('Submit catalog error:', error)
    ElMessage.error('提交失败：' + (error.response?.data?.message || error.message))
  } finally {
    catalogLoading.value = false
  }
}

const handleSubmit = async () => {
  try {
    await ElMessageBox.confirm('确定要提交此文件进行审核吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    
    await filesApi.submit(fileId.value)
    ElMessage.success('已提交审核')
    loadFileDetail()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('提交失败')
    }
  }
}

const handlePublish = async () => {
  try {
    await ElMessageBox.confirm('确定要发布此文件吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    
    await filesApi.publish(fileId.value)
    ElMessage.success('发布成功')
    loadFileDetail()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('发布失败')
    }
  }
}

const handleReject = () => {
  rejectDialogVisible.value = true
  rejectReason.value = ''
}

const confirmReject = async () => {
  if (!rejectReason.value.trim()) {
    ElMessage.warning('请输入拒绝原因')
    return
  }
  
  try {
    await filesApi.reject(fileId.value, { reason: rejectReason.value })
    ElMessage.success('已拒绝')
    rejectDialogVisible.value = false
    loadFileDetail()
  } catch (error) {
    ElMessage.error('拒绝失败')
  }
}

const handleDownload = async () => {
  try {
    // 直接创建下载链接（与文件列表页保持一致）
    const downloadUrl = `/api/v1/files/${fileId.value}/download`
    const link = document.createElement('a')
    link.href = downloadUrl
    link.style.display = 'none'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    
    ElMessage.success('开始下载文件')
  } catch (error) {
    console.error('Download error:', error)
    ElMessage.error('下载失败')
  }
}

const handleDelete = async () => {
  try {
    await ElMessageBox.confirm('确定要删除此文件吗？删除后无法恢复！', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error',
    })
    
    await filesApi.delete(fileId.value)
    ElMessage.success(t('message.deleteSuccess'))
    router.push('/files')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(t('message.deleteFailed'))
    }
  }
}

onMounted(() => {
  loadFileDetail()
})
</script>

<style scoped>
.file-detail {
  padding: 20px;
}

.preview-container {
  min-height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f5f5;
}

.video-wrapper {
  width: 100%;
}

.no-preview {
  text-align: center;
  color: #999;
}

.no-preview p {
  margin-top: 20px;
  font-size: 14px;
}

.action-buttons {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.action-buttons .el-button {
  width: 100%;
}
</style>
