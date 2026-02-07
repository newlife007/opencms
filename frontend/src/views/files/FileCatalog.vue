<template>
  <div class="file-catalog-container">
    <el-page-header @back="goBack" class="page-header">
      <template #content>
        <span class="page-title">{{ t("fileCatalog.title") }}</span>
      </template>
    </el-page-header>

    <el-card v-loading="loading" class="catalog-card">
      <!-- 文件基本信息 -->
      <template #header>
        <div class="card-header">
          <span>{{ fileInfo.title || '文件信息' }}</span>
          <el-tag v-if="fileInfo.status !== undefined" :type="getStatusType(fileInfo.status)">
            {{ getStatusText(fileInfo.status) }}
          </el-tag>
        </div>
      </template>

      <el-row :gutter="20">
        <!-- 左侧：文件预览 -->
        <el-col :span="8">
          <div class="preview-section">
            <h4>文件预览</h4>
            
            <!-- 视频预览 -->
            <div v-if="fileInfo.type === 1" class="video-preview">
              <video 
                v-if="fileInfo.preview_url" 
                :src="fileInfo.preview_url" 
                controls 
                class="preview-video"
              >
                您的浏览器不支持视频播放
              </video>
              <div v-else class="no-preview">
                <el-icon :size="60"><VideoCamera /></el-icon>
                <p>视频预览生成中...</p>
              </div>
            </div>

            <!-- 音频预览 -->
            <div v-else-if="fileInfo.type === 2" class="audio-preview">
              <audio 
                v-if="fileInfo.preview_url" 
                :src="fileInfo.preview_url" 
                controls 
                class="preview-audio"
              >
                您的浏览器不支持音频播放
              </audio>
              <div v-else class="no-preview">
                <el-icon :size="60"><Headset /></el-icon>
                <p>音频文件</p>
              </div>
            </div>

            <!-- 图片预览 -->
            <div v-else-if="fileInfo.type === 3" class="image-preview">
              <el-image 
                v-if="fileInfo.path" 
                :src="fileInfo.path" 
                fit="contain"
                class="preview-image"
              >
                <template #error>
                  <div class="image-error">
                    <el-icon :size="60"><Picture /></el-icon>
                    <p>图片加载失败</p>
                  </div>
                </template>
              </el-image>
            </div>

            <!-- 其他文件类型 -->
            <div v-else class="other-preview">
              <el-icon :size="60"><Document /></el-icon>
              <p>{{ getFileTypeName(fileInfo.type) }}</p>
            </div>

            <!-- 文件基本信息 -->
            <el-descriptions :column="1" border class="file-meta">
              <el-descriptions-item :label="t('files.fileName')">{{ fileInfo.name }}</el-descriptions-item>
              <el-descriptions-item :label="t('files.fileSize')">{{ formatFileSize(fileInfo.size) }}</el-descriptions-item>
              <el-descriptions-item :label="t('files.uploadTime')">{{ formatDate(fileInfo.upload_at) }}</el-descriptions-item>
              <el-descriptions-item :label="t('files.uploader')">{{ fileInfo.upload_username }}</el-descriptions-item>
            </el-descriptions>
          </div>
        </el-col>

        <!-- 右侧：编目表单 -->
        <el-col :span="16">
          <div class="catalog-form-section">
            <h4>{{ t("fileDetail.catalogInfo") }}</h4>
            
            <el-form 
              ref="catalogFormRef" 
              :model="catalogForm" 
              :rules="catalogRules"
              label-width="120px"
            >
              <!-- 基本信息 -->
              <el-divider content-position="left">{{ t("fileDetail.basicInfo") }}</el-divider>
              
              <el-form-item label="标题" prop="title" required>
                <el-input 
                  v-model="catalogForm.title" 
                  placeholder="请输入文件标题"
                  clearable
                />
              </el-form-item>

              <el-form-item :label="t('files.category')" prop="category_id" required>
                <el-tree-select
                  v-model="catalogForm.category_id"
                  :data="categoryTree"
                  :props="{ label: 'name', value: 'id' }"
                  placeholder="请选择分类"
                  check-strictly
                  clearable
                />
              </el-form-item>

              <el-form-item :label="t('common.description')" prop="description">
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

              <!-- 动态编目字段 -->
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
              <el-empty v-else description="该文件类型暂无扩展属性配置" />

              <!-- 操作按钮 -->
              <el-form-item>
                <el-button type="primary" @click="submitCatalog" :loading="submitting">
                  <el-icon><Check /></el-icon>
                  保存并提交审核
                </el-button>
                <el-button @click="saveDraft" :loading="submitting">
                  <el-icon><Document /></el-icon>{{ t("fileCatalog.saveDraft") }}</el-button>
                <el-button @click="goBack">
                  <el-icon><Close /></el-icon>{{ t("common.cancel") }}</el-button>
              </el-form-item>
            </el-form>
          </div>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  VideoCamera, 
  Headset, 
  Picture, 
  Document, 
  Check, 
  Close 
} from '@element-plus/icons-vue'
import axios from '@/utils/request'
import filesApi from '@/api/files'


const { t } = useI18n()

const route = useRoute()
const router = useRouter()

// 数据
const loading = ref(false)
const submitting = ref(false)
const fileInfo = ref({})
const categoryTree = ref([])
const levels = ref([])
const groups = ref([])
const catalogFields = ref([])
const catalogFormRef = ref(null)

// 编目表单
const catalogForm = reactive({
  title: '',
  category_id: null,
  description: '',
  level: null,
  groups: [],
  is_download: true,
  catalog_info: {}
})

// 表单验证规则
const catalogRules = {
  title: [
    { required: true, message: '请输入文件标题', trigger: 'blur' },
    { min: 2, max: 200, message: '标题长度在 2 到 200 个字符', trigger: 'blur' }
  ],
  category_id: [
    { required: true, message: '请选择分类', trigger: 'change' }
  ]
}

// 获取文件信息
const fetchFileInfo = async () => {
  const fileId = route.params.id
  loading.value = true
  try {
    const response = await axios.get(`/files/${fileId}`)
    fileInfo.value = response.data
    
    // 初始化表单数据
    catalogForm.title = response.data.title || response.data.name
    catalogForm.category_id = response.data.category_id
    catalogForm.description = response.data.description || ''
    catalogForm.level = response.data.level
    catalogForm.groups = response.data.groups ? response.data.groups.split(',').map(Number) : []
    catalogForm.is_download = response.data.is_download !== 0
    catalogForm.catalog_info = response.data.catalog_info || {}
    
  } catch (error) {
    ElMessage.error('获取文件信息失败：' + (error.response?.data?.message || error.message))
  } finally {
    loading.value = false
  }
}

// 获取分类树
const fetchCategories = async () => {
  try {
    const response = await axios.get('/categories')
    categoryTree.value = buildTree(response.data)
  } catch (error) {
    console.error('获取分类失败：', error)
  }
}

// 获取等级列表
const fetchLevels = async () => {
  try {
    const response = await axios.get('/admin/levels')
    levels.value = response.data
  } catch (error) {
    console.error('获取等级失败：', error)
  }
}

// 获取组列表
const fetchGroups = async () => {
  try {
    const response = await axios.get('/admin/groups')
    groups.value = response.data
  } catch (error) {
    console.error('获取组失败：', error)
  }
}

// 获取编目字段配置
const fetchCatalogFields = async () => {
  if (!fileInfo.value.type) return
  
  try {
    console.log('Fetching catalog fields for type:', fileInfo.value.type)
    const response = await axios.get(`/catalog?type=${fileInfo.value.type}`)
    console.log('Catalog API response:', response)
    
    // Backend returns { success: true, type: X, catalog: [...] }
    if (response.success && response.catalog) {
      console.log('Raw catalog tree:', response.catalog)
      const flattenedFields = flattenCatalogTree(response.catalog)
      console.log('Flattened catalog fields:', flattenedFields)
      catalogFields.value = flattenedFields
    } else {
      console.warn('No catalog data in response')
      catalogFields.value = []
    }
  } catch (error) {
    console.error('获取编目字段失败：', error)
    console.error('Error details:', error.response)
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

// 构建树结构
const buildTree = (items, parentId = null) => {
  return items
    .filter(item => item.parent_id === parentId)
    .map(item => ({
      ...item,
      children: buildTree(items, item.id)
    }))
}

// 保存草稿
const saveDraft = async () => {
  submitting.value = true
  try {
    const fileId = route.params.id
    const data = {
      ...catalogForm,
      groups: catalogForm.groups.join(','),
      is_download: catalogForm.is_download ? 1 : 0,
      status: 0 // 草稿状态
    }
    
    await axios.put(`/files/${fileId}`, data)
    ElMessage.success('草稿保存成功')
    
  } catch (error) {
    ElMessage.error('保存失败：' + (error.response?.data?.message || error.message))
  } finally {
    submitting.value = false
  }
}

// 提交编目
const submitCatalog = async () => {
  // 表单验证
  if (!catalogFormRef.value) return
  
  await catalogFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    submitting.value = true
    try {
      const fileId = route.params.id
      
      // 步骤1：保存编目信息
      const data = {
        ...catalogForm,
        groups: catalogForm.groups.join(','),
        is_download: catalogForm.is_download ? 1 : 0
      }
      
      await axios.put(`/files/${fileId}`, data)
      console.log('✓ 编目信息已保存')
      
      // 步骤2：提交审核（调用 submit API）
      await filesApi.submit(fileId)
      console.log('✓ 文件已提交审核')
      
      ElMessage.success('编目信息已保存并提交审核')
      
      // 延迟跳转
      setTimeout(() => {
        router.push('/files')
      }, 1500)
      
    } catch (error) {
      console.error('提交失败:', error)
      ElMessage.error('提交失败：' + (error.response?.data?.message || error.message))
    } finally {
      submitting.value = false
    }
  })
}

// 返回
const goBack = () => {
  router.back()
}

// 格式化文件大小
const formatFileSize = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

// 格式化日期
const formatDate = (timestamp) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleString('zh-CN')
}

// 获取文件类型名称
const getFileTypeName = (type) => {
  const types = { 1: '视频', 2: '音频', 3: '图片', 4: '富媒体' }
  return types[type] || '未知'
}

// 获取状态类型
const getStatusType = (status) => {
  const types = { 0: 'info', 1: 'warning', 2: 'success', 3: 'danger', 4: 'info' }
  return types[status] || 'info'
}

// 获取状态文本
const getStatusText = (status) => {
  const texts = { 0: '新上传', 1: '待审核', 2: '已发布', 3: '已拒绝', 4: '已删除' }
  return texts[status] || '未知'
}

// 初始化
onMounted(async () => {
  await fetchFileInfo()
  await Promise.all([
    fetchCategories(),
    fetchLevels(),
    fetchGroups(),
    fetchCatalogFields()
  ])
})
</script>

<style scoped>
.file-catalog-container {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.page-title {
  font-size: 18px;
  font-weight: bold;
}

.catalog-card {
  min-height: 600px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.preview-section {
  position: sticky;
  top: 20px;
}

.preview-section h4 {
  margin: 0 0 15px 0;
  font-size: 16px;
  color: #303133;
}

.video-preview,
.audio-preview,
.image-preview,
.other-preview {
  margin-bottom: 20px;
  background: #f5f7fa;
  border-radius: 8px;
  overflow: hidden;
}

.preview-video {
  width: 100%;
  max-height: 300px;
}

.preview-audio {
  width: 100%;
  margin: 20px 0;
}

.preview-image {
  width: 100%;
  max-height: 400px;
}

.no-preview,
.image-error,
.other-preview {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: #909399;
}

.file-meta {
  margin-top: 20px;
}

.catalog-form-section h4 {
  margin: 0 0 20px 0;
  font-size: 16px;
  color: #303133;
}

.el-divider {
  margin: 30px 0 20px 0;
}

.el-divider:first-of-type {
  margin-top: 0;
}

/* 响应式 */
@media (max-width: 768px) {
  .preview-section {
    position: relative;
    top: 0;
    margin-bottom: 20px;
  }
}
</style>
